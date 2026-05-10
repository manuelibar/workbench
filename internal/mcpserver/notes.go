package mcpserver

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/manuelibar/workbench/internal/domain"
	"github.com/manuelibar/workbench/internal/id"
	"github.com/manuelibar/workbench/internal/pgstore"
)

// NoteWire is the JSON shape the MCP tool surface uses for notes. UUIDs
// travel as canonical strings so the auto-generated JSON Schema for tool
// outputs is unambiguous to clients.
type NoteWire struct {
	ID          string    `json:"id"`
	BodyMD      string    `json:"body_md"`
	Tags        []string  `json:"tags"`
	NamespaceID string    `json:"namespace_id,omitempty"`
	ProjectID   string    `json:"project_id,omitempty"`
	PromotedTo  string    `json:"promoted_to,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func noteToWire(n domain.Note) NoteWire {
	w := NoteWire{
		ID:        n.ID.String(),
		BodyMD:    n.BodyMD,
		Tags:      n.Tags,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
	}
	if n.NamespaceID != nil {
		w.NamespaceID = n.NamespaceID.String()
	}
	if n.ProjectID != nil {
		w.ProjectID = n.ProjectID.String()
	}
	if n.PromotedTo != nil {
		w.PromotedTo = n.PromotedTo.String()
	}
	return w
}

func notesToWire(ns []domain.Note) []NoteWire {
	out := make([]NoteWire, len(ns))
	for i, n := range ns {
		out[i] = noteToWire(n)
	}
	return out
}

// note.add ---------------------------------------------------------------

type NoteAddArgs struct {
	Text string   `json:"text" jsonschema:"the note body, in markdown"`
	Tags []string `json:"tags,omitempty" jsonschema:"optional tag list for filtering and grouping"`
}

type NoteAddResult struct {
	Note NoteWire `json:"note"`
}

func (s *Server) handleNoteAdd(ctx context.Context, _ *mcp.CallToolRequest, args NoteAddArgs) (*mcp.CallToolResult, NoteAddResult, error) {
	if args.Text == "" {
		return nil, NoteAddResult{}, fmt.Errorf("note.add: text must not be empty")
	}
	ws, err := s.store.EnsureOpenWorkSession(ctx, s.user.ID, "")
	if err != nil {
		return nil, NoteAddResult{}, fmt.Errorf("note.add: read work session: %w", err)
	}
	n := domain.Note{
		UserID:      s.user.ID,
		BodyMD:      args.Text,
		Tags:        args.Tags,
		NamespaceID: ws.Selection.NamespaceID,
		ProjectID:   ws.Selection.ProjectID,
	}
	if a, ok := id.FromContext(ctx); ok {
		n.IdempotencyKey = a.IdempotencyKey
	}
	inserted, err := s.store.AddNote(ctx, n)
	if err != nil {
		return nil, NoteAddResult{}, err
	}
	return nil, NoteAddResult{Note: noteToWire(inserted)}, nil
}

// note.list --------------------------------------------------------------

type NoteListArgs struct {
	Tag         string `json:"tag,omitempty"          jsonschema:"filter by exact tag membership"`
	NamespaceID string `json:"namespace_id,omitempty" jsonschema:"filter by capture-time namespace UUID"`
	ProjectID   string `json:"project_id,omitempty"   jsonschema:"filter by capture-time project UUID"`
	Since       string `json:"since,omitempty"        jsonschema:"only notes created at or after this RFC3339 timestamp"`
	Limit       int    `json:"limit,omitempty"        jsonschema:"max results (default 50, max 200)"`
}

type NoteListResult struct {
	Notes []NoteWire `json:"notes"`
	Count int        `json:"count"`
}

func (s *Server) handleNoteList(ctx context.Context, _ *mcp.CallToolRequest, args NoteListArgs) (*mcp.CallToolResult, NoteListResult, error) {
	f := pgstore.ListNotesFilter{Tag: args.Tag, Limit: args.Limit}
	if args.NamespaceID != "" {
		nsID, err := uuid.Parse(args.NamespaceID)
		if err != nil {
			return nil, NoteListResult{}, fmt.Errorf("note.list: namespace_id: %w", err)
		}
		f.NamespaceID = &nsID
	}
	if args.ProjectID != "" {
		pID, err := uuid.Parse(args.ProjectID)
		if err != nil {
			return nil, NoteListResult{}, fmt.Errorf("note.list: project_id: %w", err)
		}
		f.ProjectID = &pID
	}
	if args.Since != "" {
		t, err := time.Parse(time.RFC3339, args.Since)
		if err != nil {
			return nil, NoteListResult{}, fmt.Errorf("note.list: since: %w", err)
		}
		f.Since = &t
	}
	notes, err := s.store.ListNotes(ctx, s.user.ID, f)
	if err != nil {
		return nil, NoteListResult{}, err
	}
	return nil, NoteListResult{Notes: notesToWire(notes), Count: len(notes)}, nil
}

// note.search ------------------------------------------------------------

type NoteSearchArgs struct {
	Query string `json:"query" jsonschema:"substring to match against note body (case-insensitive)"`
	Limit int    `json:"limit,omitempty" jsonschema:"max results (default 50, max 200)"`
}

type NoteSearchResult struct {
	Notes []NoteWire `json:"notes"`
	Count int        `json:"count"`
}

func (s *Server) handleNoteSearch(ctx context.Context, _ *mcp.CallToolRequest, args NoteSearchArgs) (*mcp.CallToolResult, NoteSearchResult, error) {
	hits, err := s.store.SearchNotes(ctx, s.user.ID, args.Query, args.Limit)
	if err != nil {
		return nil, NoteSearchResult{}, err
	}
	return nil, NoteSearchResult{Notes: notesToWire(hits), Count: len(hits)}, nil
}

// note.get ---------------------------------------------------------------

type NoteGetArgs struct {
	ID string `json:"id" jsonschema:"note id (UUID)"`
}

type NoteGetResult struct {
	Note NoteWire `json:"note"`
}

func (s *Server) handleNoteGet(ctx context.Context, _ *mcp.CallToolRequest, args NoteGetArgs) (*mcp.CallToolResult, NoteGetResult, error) {
	id, err := uuid.Parse(args.ID)
	if err != nil {
		return nil, NoteGetResult{}, fmt.Errorf("note.get: id: %w", err)
	}
	n, err := s.store.GetNote(ctx, s.user.ID, id)
	if err != nil {
		return nil, NoteGetResult{}, err
	}
	return nil, NoteGetResult{Note: noteToWire(n)}, nil
}

// note.update ------------------------------------------------------------

type NoteUpdateArgs struct {
	ID   string   `json:"id"             jsonschema:"note id (UUID)"`
	Text string   `json:"text,omitempty" jsonschema:"new body (markdown); leave empty to keep existing"`
	Tags []string `json:"tags,omitempty" jsonschema:"new tag list; omit to keep existing, pass [] to clear"`
}

type NoteUpdateResult struct {
	Note NoteWire `json:"note"`
}

func (s *Server) handleNoteUpdate(ctx context.Context, _ *mcp.CallToolRequest, args NoteUpdateArgs) (*mcp.CallToolResult, NoteUpdateResult, error) {
	id, err := uuid.Parse(args.ID)
	if err != nil {
		return nil, NoteUpdateResult{}, fmt.Errorf("note.update: id: %w", err)
	}
	var f pgstore.UpdateNoteFields
	if args.Text != "" {
		text := args.Text
		f.BodyMD = &text
	}
	if args.Tags != nil {
		tags := args.Tags
		f.Tags = &tags
	}
	n, err := s.store.UpdateNote(ctx, s.user.ID, id, f)
	if err != nil {
		return nil, NoteUpdateResult{}, err
	}
	return nil, NoteUpdateResult{Note: noteToWire(n)}, nil
}

// note.delete ------------------------------------------------------------

type NoteDeleteArgs struct {
	ID string `json:"id" jsonschema:"note id (UUID)"`
}

type NoteDeleteResult struct {
	Deleted bool   `json:"deleted"`
	ID      string `json:"id"`
}

func (s *Server) handleNoteDelete(ctx context.Context, _ *mcp.CallToolRequest, args NoteDeleteArgs) (*mcp.CallToolResult, NoteDeleteResult, error) {
	id, err := uuid.Parse(args.ID)
	if err != nil {
		return nil, NoteDeleteResult{}, fmt.Errorf("note.delete: id: %w", err)
	}
	if err := s.store.DeleteNote(ctx, s.user.ID, id); err != nil {
		return nil, NoteDeleteResult{}, err
	}
	return nil, NoteDeleteResult{Deleted: true, ID: id.String()}, nil
}

// registerNotes registers all note.* tools on srv.
func (s *Server) registerNotes(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "note.add",
		Description: "Capture a quick markdown note. Tags are optional. Capture-time selection (namespace/project) is recorded automatically. Idempotent if the request supplies an idempotency key.",
	}, s.handleNoteAdd)

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "note.list",
		Description: "List notes most-recent-first, with optional tag/namespace/project/time filters. Returns at most 200 notes.",
	}, s.handleNoteList)

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "note.search",
		Description: "Substring (case-insensitive) search across note bodies. v0 uses ILIKE; semantic search arrives in a later phase.",
	}, s.handleNoteSearch)

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "note.get",
		Description: "Fetch a single note by id.",
	}, s.handleNoteGet)

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "note.update",
		Description: "Patch a note's body and/or tags. Omit a field to leave it unchanged; pass an empty array for tags to clear them.",
	}, s.handleNoteUpdate)

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "note.delete",
		Description: "Delete a note by id. Returns {deleted: true} on success or an error if no such note exists.",
	}, s.handleNoteDelete)
}
