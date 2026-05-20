package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerResources registers the static (always-present) resources and
// templates. Skill-contributed resources are registered dynamically by
// refreshCapabilities() on every refresh() call.
func (s *Server) registerResources() {
	s.sdkServer.AddResource(&mcp.Resource{
		URI:         ScopeOverviewURI,
		Name:        "Scope overview",
		Description: "Session selection state and next recommended resources",
		MIMEType:    "application/json",
	}, s.handleScopeOverview)

	s.sdkServer.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "workbench:///projects/{id}",
		Name:        "Project",
		Description: "Full project details including system prompt",
		MIMEType:    "application/json",
	}, s.handleProject)

	s.sdkServer.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "workbench:///roles/{id}",
		Name:        "Role",
		Description: "Role details including system prompt",
		MIMEType:    "application/json",
	}, s.handleRoleResource)

	s.sdkServer.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "workbench:///boards/{id}",
		Name:        "Board",
		Description: "Board details",
		MIMEType:    "application/json",
	}, s.handleBoardResource)

	s.sdkServer.AddResource(&mcp.Resource{
		URI:         "workbench:///github/config",
		Name:        "GitHub config",
		Description: "GitHub org integration config supplied by MCP stdio env",
		MIMEType:    "application/json",
	}, s.handleGitHubConfigResource)

	s.sdkServer.AddResource(&mcp.Resource{
		URI:         "workbench:///context/project-snapshot",
		Name:        "Project snapshot",
		Description: "README/docs/tech-stack snapshot for the configured project root",
		MIMEType:    "application/json",
	}, s.handleProjectSnapshotResource)

	s.sdkServer.AddResource(&mcp.Resource{
		URI:         "workbench:///tasks",
		Name:        "Tasks",
		Description: "Tracked task state for the active Workbench server",
		MIMEType:    "application/json",
	}, s.handleTasksResource)

	s.sdkServer.AddResource(&mcp.Resource{
		URI:         "workbench:///knowledge",
		Name:        "Knowledge",
		Description: "Feedback and notes captured as queryable Workbench knowledge",
		MIMEType:    "application/json",
	}, s.handleKnowledgeResource)
}

// ---- scope/overview ----

func (s *Server) handleScopeOverview(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	s.mu.Lock()
	sel := s.sel
	dynamicTools := append([]string(nil), s.dynamicTools...)
	s.mu.Unlock()
	navigation := s.composeNavigation(sel)
	body := s.composeScopeOverview(sel, navigation, dynamicTools)
	return jsonResource(req.Params.URI, body)
}

// ---- projects/{id} ----

func (s *Server) handleProject(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	idStr := strings.TrimPrefix(req.Params.URI, "workbench:///projects/")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, mcp.ResourceNotFoundError(req.Params.URI)
	}
	p, err := s.store.Get(id)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			return nil, mcp.ResourceNotFoundError(req.Params.URI)
		}
		return nil, err
	}
	return jsonResource(req.Params.URI, projectToWire(p))
}

func (s *Server) handleRoleResource(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	idStr := strings.TrimPrefix(req.Params.URI, "workbench:///roles/")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, mcp.ResourceNotFoundError(req.Params.URI)
	}
	s.mu.Lock()
	role, ok := s.roles[id]
	s.mu.Unlock()
	if !ok {
		return nil, mcp.ResourceNotFoundError(req.Params.URI)
	}
	return jsonResource(req.Params.URI, map[string]any{"id": role.ID.String(), "name": role.Name, "description": role.Description, "system_prompt": role.SystemPrompt, "created_at": role.CreatedAt})
}

func (s *Server) handleBoardResource(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	idStr := strings.TrimPrefix(req.Params.URI, "workbench:///boards/")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, mcp.ResourceNotFoundError(req.Params.URI)
	}
	s.mu.Lock()
	b, ok := s.boards[id]
	s.mu.Unlock()
	if !ok {
		return nil, mcp.ResourceNotFoundError(req.Params.URI)
	}
	return jsonResource(req.Params.URI, map[string]any{"id": b.ID.String(), "project_id": b.ProjectID.String(), "namespace_id": b.NamespaceID.String(), "name": b.Name, "description": b.Description, "created_at": b.CreatedAt})
}

func (s *Server) handleGitHubConfigResource(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	s.mu.Lock()
	cfg := s.github
	s.mu.Unlock()
	if cfg.Organization == "" {
		return jsonResource(req.Params.URI, map[string]any{"configured": false})
	}
	return jsonResource(req.Params.URI, map[string]any{"configured": true, "organization": cfg.Organization, "token_configured": cfg.Token != ""})
}

func (s *Server) handleProjectSnapshotResource(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	return jsonResource(req.Params.URI, SnapshotProject(s.projectRoot))
}

// ---- skills/{name}/manifest ----
// Registered dynamically per active skill bundle by refreshCapabilities().

type skillFileEntry struct {
	Path     string `json:"path"`
	URI      string `json:"uri"`
	MIMEType string `json:"mime_type"`
}

type skillManifestBody struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Version     string           `json:"version"`
	SkillMDURI  string           `json:"skill_md_uri"`
	Files       []skillFileEntry `json:"files"`
}

func (s *Server) handleSkillManifest(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	name := skillNameFrom(req.Params.URI, "/manifest")
	b, ok := s.skills.Get(name)
	if !ok {
		return nil, mcp.ResourceNotFoundError(req.Params.URI)
	}
	base := "skill://" + b.Name + "/"
	var files []skillFileEntry
	for _, f := range b.Files {
		if f.RelPath == "SKILL.md" {
			continue
		}
		files = append(files, skillFileEntry{
			Path:     f.RelPath,
			URI:      base + f.RelPath,
			MIMEType: f.MIMEType,
		})
	}
	if files == nil {
		files = []skillFileEntry{}
	}
	return jsonResource(req.Params.URI, skillManifestBody{
		Name:        b.Name,
		Description: b.Description,
		Version:     b.Version,
		SkillMDURI:  base + "SKILL.md",
		Files:       files,
	})
}

// ---- skills/{name}/SKILL.md ----
// Registered dynamically per active skill bundle by refreshCapabilities().

func (s *Server) handleSkillSKILLMD(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	name := skillNameFrom(req.Params.URI, "/SKILL.md")
	b, ok := s.skills.Get(name)
	if !ok {
		return nil, mcp.ResourceNotFoundError(req.Params.URI)
	}

	sel := s.getSelection()
	var content []byte
	var mimeType string
	found := false
	for _, f := range b.Files {
		if f.RelPath != "SKILL.md" {
			continue
		}
		s.mu.Lock()
		projCtx := s.projectContext(sel)
		s.mu.Unlock()
		content = f.Content(projCtx)
		mimeType = f.MIMEType
		found = true
		break
	}
	if !found {
		return nil, mcp.ResourceNotFoundError(req.Params.URI)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      req.Params.URI,
			MIMEType: mimeType,
			Text:     string(content),
		}},
	}, nil
}

// ---- helpers ----

func jsonResource(uri string, v any) (*mcp.ReadResourceResult, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      uri,
			MIMEType: "application/json",
			Text:     string(b),
		}},
	}, nil
}

// skillNameFrom extracts the bundle name from a skill resource URI.
// e.g. skillNameFrom("skill://foo/manifest", "/manifest") → "foo"
func skillNameFrom(uri, suffix string) string {
	s := strings.TrimPrefix(uri, "skill://")
	s = strings.TrimPrefix(s, "workbench:///skills/")
	return strings.TrimSuffix(s, suffix)
}
