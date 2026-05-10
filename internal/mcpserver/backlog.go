package mcpserver

import (
	"context"
	"errors"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/manuelibar/workbench/internal/domain"
	"github.com/manuelibar/workbench/internal/id"
	"github.com/manuelibar/workbench/internal/pgstore"
)

// backlogArtifactType is the artifact type the backlog verbs operate on.
const backlogArtifactType = "task"

// backlog.add ------------------------------------------------------------

type BacklogAddArgs struct {
	Title string `json:"title" jsonschema:"task title (becomes the artifact's content_text and content.title)"`
	Body  string `json:"body,omitempty" jsonschema:"longer markdown description"`
}

type BacklogAddResult struct {
	Artifact ArtifactWire `json:"artifact"`
}

func (s *Server) handleBacklogAdd(ctx context.Context, _ *mcp.CallToolRequest, args BacklogAddArgs) (*mcp.CallToolResult, BacklogAddResult, error) {
	pID, err := s.resolveProjectID("")
	if err != nil {
		return nil, BacklogAddResult{}, err
	}
	if args.Title == "" {
		return nil, BacklogAddResult{}, errors.New("backlog.add: title required")
	}
	a := domain.Artifact{
		ProjectID: pID,
		Type:      backlogArtifactType,
		Status:    domain.ArtifactStatusDraft,
	}
	if x, ok := id.FromContext(ctx); ok {
		a.IdempotencyKey = x.IdempotencyKey
	}
	out, err := s.store.CreateArtifact(ctx, pgstore.CreateArtifactInput{
		Artifact:    a,
		Content:     map[string]any{"title": args.Title, "body": args.Body},
		ContentText: args.Title,
	})
	if err != nil {
		return nil, BacklogAddResult{}, err
	}
	return nil, BacklogAddResult{Artifact: artifactToWire(out)}, nil
}

// backlog.list -----------------------------------------------------------

type BacklogListArgs struct {
	Status string `json:"status,omitempty" jsonschema:"filter by status (defaults to all). Common values: draft, reviewing, signed_off, archived"`
	Limit  int    `json:"limit,omitempty"`
}

type BacklogListResult struct {
	Tasks []ArtifactWire `json:"tasks"`
	Count int            `json:"count"`
}

func (s *Server) handleBacklogList(ctx context.Context, _ *mcp.CallToolRequest, args BacklogListArgs) (*mcp.CallToolResult, BacklogListResult, error) {
	pID, err := s.resolveProjectID("")
	if err != nil {
		return nil, BacklogListResult{}, err
	}
	list, err := s.store.ListArtifacts(ctx, pID, pgstore.ListArtifactsFilter{
		Type:   backlogArtifactType,
		Status: args.Status,
		Limit:  args.Limit,
	})
	if err != nil {
		return nil, BacklogListResult{}, err
	}
	out := make([]ArtifactWire, len(list))
	for i, a := range list {
		out[i] = artifactToWire(a)
	}
	return nil, BacklogListResult{Tasks: out, Count: len(out)}, nil
}

// backlog.take_next ------------------------------------------------------

type BacklogTakeNextResult struct {
	Task    ArtifactWire `json:"task"`
	Found   bool         `json:"found"`
	Message string       `json:"message,omitempty"`
}

func (s *Server) handleBacklogTakeNext(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, BacklogTakeNextResult, error) {
	pID, err := s.resolveProjectID("")
	if err != nil {
		return nil, BacklogTakeNextResult{}, err
	}
	// "Take next" returns the OLDEST draft task. ListArtifacts is sorted
	// by updated_at DESC, so we just walk the list backwards. For a real
	// backlog at scale, a dedicated query with ORDER BY created_at ASC
	// LIMIT 1 belongs in the store; v0 reuses ListArtifacts.
	list, err := s.store.ListArtifacts(ctx, pID, pgstore.ListArtifactsFilter{
		Type:   backlogArtifactType,
		Status: domain.ArtifactStatusDraft,
		Limit:  pgstore.MaxNoteListLimit,
	})
	if err != nil {
		return nil, BacklogTakeNextResult{}, err
	}
	if len(list) == 0 {
		return nil, BacklogTakeNextResult{
			Found:   false,
			Message: fmt.Sprintf("no draft tasks in this project"),
		}, nil
	}
	oldest := list[len(list)-1]
	return nil, BacklogTakeNextResult{Found: true, Task: artifactToWire(oldest)}, nil
}

// registerBacklog wires backlog.* on srv (visible while a project is selected).
func (s *Server) registerBacklog(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "backlog.add",
		Description: "Add a task to the project backlog. Internally creates an artifact with type='task'; the title is stored as content_text + content.title.",
	}, s.handleBacklogAdd)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "backlog.list",
		Description: "List tasks (artifacts of type='task') in the currently-selected project, optionally filtered by status.",
	}, s.handleBacklogList)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "backlog.take_next",
		Description: "Return the oldest draft task in the current project, or {found: false} if none exist.",
	}, s.handleBacklogTakeNext)
}
