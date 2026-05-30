package mcpserver

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/manuelibar/workbench/internal/mcpserver/skills"
)

// selection holds the active project selection for this server instance.
// With stdio transport there is exactly one session per process.
type selection struct {
	ProjectID   *uuid.UUID
	NamespaceID *uuid.UUID
	RoleID      *uuid.UUID
	BoardID     *uuid.UUID
}

// Server is the Workbench MCP server. One instance = one agent session.
type Server struct {
	sdkServer   *mcp.Server
	log         *slog.Logger
	store       ProjectStore
	skills      skills.SkillRegistry
	projectRoot string

	mu             sync.Mutex
	sel            selection
	dynamicURIs    []string // skill-contributed resource URIs currently registered
	dynamicTools   []string // scope-contributed tool names currently registered
	namespaces     map[uuid.UUID]Namespace
	roles          map[uuid.UUID]Role
	boards         map[uuid.UUID]Board
	tasks          map[uuid.UUID]Task
	knowledge      map[uuid.UUID]KnowledgeItem
	kbRetriever    KBRetriever
	querySynth     QuerySynthesizer
	queryResources map[string]querySkillResource
	github         GitHubConfig

	syncMu               sync.Mutex
	refreshListWait      time.Duration
	capabilityGeneration int64
	capabilitySync       *capabilitySyncTracker
}

const serverInstructions = `This is the Workbench MCP server — an adaptive capability kernel.

Call refresh() at the start of every session, when changing selected scope, and after context compaction.
Pass project_id to scope your session to a specific project.

refresh() is the synchronization point: it updates all capabilities (tools, resources,
prompts) for your selection, waits briefly for tools/list and resources/list calls,
then returns your working context and capability_sync status:
- selection: the active scope
- overview: the same logical briefing available later at workbench:///scope/overview
- skills[]: every relevant skill bundle with instructions already rendered
- capability_sync: client_relisted or fallback_index

Use workbench:///scope/overview only to re-read the current briefing without side effects.
Use MCP list methods for full capability discovery. If capability_sync.status is fallback_index,
use the inline index as the fresh capability surface.`

// New returns a configured Server ready to run.
func New(log *slog.Logger, store ProjectStore, registry skills.SkillRegistry) *Server {
	s := &Server{
		log:             log,
		store:           store,
		skills:          registry,
		namespaces:      map[uuid.UUID]Namespace{},
		roles:           map[uuid.UUID]Role{},
		boards:          map[uuid.UUID]Board{},
		tasks:           map[uuid.UUID]Task{},
		knowledge:       map[uuid.UUID]KnowledgeItem{},
		queryResources:  map[string]querySkillResource{},
		github:          GitHubConfig{},
		refreshListWait: defaultRefreshListWait,
	}
	// Seed default namespace for GitHub org mapping.
	nsID := uuid.New()
	s.namespaces[nsID] = Namespace{ID: nsID, Name: "organization", CreatedAt: time.Now().UTC()}
	s.sdkServer = mcp.NewServer(
		&mcp.Implementation{Name: "workbench", Version: "0.1.0"},
		&mcp.ServerOptions{Instructions: serverInstructions},
	)
	s.registerCoreTools()
	s.registerResources()
	s.installCapabilityListMiddleware()
	return s
}

// Run connects via stdio and blocks until the context is cancelled.
func (s *Server) Run(ctx context.Context) error {
	ss, err := s.sdkServer.Connect(ctx, &mcp.StdioTransport{}, nil)
	if err != nil {
		return fmt.Errorf("stdio connect: %w", err)
	}
	defer ss.Close()
	<-ctx.Done()
	return nil
}

// SDKServer returns the underlying mcp.Server, used in tests to connect via
// in-memory transport.
func (s *Server) SDKServer() *mcp.Server { return s.sdkServer }

// SetGitHubConfig applies integration config supplied by the MCP stdio client
// environment, not by a runtime tool.
func (s *Server) SetGitHubConfig(cfg GitHubConfig) {
	s.mu.Lock()
	s.github = cfg
	s.mu.Unlock()
}

// SetProjectRoot configures the repository/document root used for adaptive context resources.
func (s *Server) SetProjectRoot(root string) { s.projectRoot = root }

// getSelection returns a snapshot of the current selection.
func (s *Server) getSelection() selection {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sel
}

// refreshCapabilities removes all skill-contributed resources registered by the
// previous selection and registers the ones for the new selection.
// Must be called with s.mu held.
func (s *Server) refreshCapabilities(sel selection) {
	if len(s.dynamicURIs) > 0 {
		s.sdkServer.RemoveResources(s.dynamicURIs...)
		s.dynamicURIs = s.dynamicURIs[:0]
	}
	s.registerScopeTools(sel)

	for _, b := range s.skills.For(sel.ProjectID != nil) {
		manifestURI := "skill://" + b.Name + "/manifest"
		skillMDURI := "skill://" + b.Name + "/SKILL.md"

		s.sdkServer.AddResource(&mcp.Resource{
			URI:         manifestURI,
			Name:        b.Name + " manifest",
			Description: b.Description,
			MIMEType:    "application/json",
		}, s.handleSkillManifest)

		s.sdkServer.AddResource(&mcp.Resource{
			URI:         skillMDURI,
			Name:        b.Name + " SKILL.md",
			Description: "Skill instructions for " + b.Name,
			MIMEType:    "text/markdown",
		}, s.handleSkillSKILLMD)

		s.dynamicURIs = append(s.dynamicURIs, manifestURI, skillMDURI)
	}
}
