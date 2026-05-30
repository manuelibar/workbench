package mcpserver

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/manuelibar/workbench/internal/backlogclient"
	"github.com/manuelibar/workbench/internal/domain"
	"github.com/manuelibar/workbench/internal/mcpserver/middleware"
	"github.com/manuelibar/workbench/internal/pgstore"
)

// Implementation reported to MCP clients during initialize.
const (
	implName    = "workbench"
	implVersion = "v0.1.0"
)

// Server is the workbench MCP front door. It wraps a singleton [*mcp.Server]
// shared across every connected MCP-protocol session: when the selection
// changes, the tool surface is mutated on this single server and the SDK
// fans out `notifications/tools/list_changed` to every active session for free.
//
// Concurrency model: a [sync.Mutex] guards [Server.selection] +
// [Server.activeTools]. The mutex is held only briefly during selection
// changes; tool handlers themselves do not hold it.
type Server struct {
	store         *pgstore.Store
	user          domain.User
	workSessionID uuid.UUID
	log           *slog.Logger

	// backlog is the typed HTTP client for the standalone backlog-service.
	// nil is allowed: backlog.* handlers then return a clear "not
	// configured" error, leaving the rest of the surface functional.
	backlog *backlogclient.Client

	sdkServer *mcp.Server // singleton; receives every session

	selectionMu sync.Mutex
	selection   domain.Selection
	activeTools map[string]bool // currently registered on sdkServer
}

// New constructs a [Server] backed by store. The user and the WorkSession
// id are recorded so handlers can attribute writes correctly. backlog may
// be nil to disable the `backlog.*` surface (handlers return a clear
// error). The initial selection is loaded from ws and the corresponding
// tool surface is registered eagerly so the first connecting client sees
// the right tools without having to call refresh first.
func New(store *pgstore.Store, user domain.User, ws domain.WorkSession, backlog *backlogclient.Client, log *slog.Logger) *Server {
	if log == nil {
		log = slog.Default()
	}
	s := &Server{
		store:         store,
		user:          user,
		workSessionID: ws.ID,
		log:           log,
		backlog:       backlog,
		selection:     ws.Selection,
		activeTools:   map[string]bool{},
	}
	s.sdkServer = s.buildSDKServer()
	s.applyVisibility(s.selection)
	return s
}

// Handler returns the streamable-HTTP handler suitable for mounting on a
// net/http mux at /mcp. Every new MCP-protocol session shares the same
// underlying [*mcp.Server].
func (s *Server) Handler() http.Handler {
	return mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return s.sdkServer
	}, &mcp.StreamableHTTPOptions{
		Logger: s.log,
	})
}

// SDKServer exposes the singleton [*mcp.Server] for in-memory testing,
// where tests connect [mcp.NewInMemoryTransports] pairs directly to it.
func (s *Server) SDKServer() *mcp.Server { return s.sdkServer }

// currentSelection returns a snapshot of the current selection. Safe for
// concurrent use.
func (s *Server) currentSelection() domain.Selection {
	s.selectionMu.Lock()
	defer s.selectionMu.Unlock()
	return s.selection
}

// buildSDKServer constructs the singleton [*mcp.Server] with the workbench
// implementation info, instructions, middleware, and resources. Tools are
// registered later by [Server.applyVisibility] (called from [New]).
func (s *Server) buildSDKServer() *mcp.Server {
	srv := mcp.NewServer(
		&mcp.Implementation{Name: implName, Version: implVersion},
		&mcp.ServerOptions{
			Instructions: "Workbench MCP — read resource workbench://skill for orientation; call refresh to sync state and change selection; use artifact.begin plus artifact.guidance/validate/elicit for typed artifact authoring; call ask to elicit general input from the user.",
			Logger:       s.log,
		},
	)
	srv.AddReceivingMiddleware(
		middleware.IDs(),
		middleware.Events(s.store, s.workSessionID, s.log),
		middleware.Slog(s.log),
	)
	s.registerResources(srv)
	return srv
}
