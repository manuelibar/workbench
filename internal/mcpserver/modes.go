package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/manuelibar/workbench/internal/domain"
	"github.com/manuelibar/workbench/internal/id"
	"github.com/manuelibar/workbench/internal/pgstore"
)

// ModeWire is the JSON shape mode tools return.
type ModeWire struct {
	ID           string         `json:"id"`
	BlueprintID  string         `json:"blueprint_id"`
	Name         string         `json:"name"`
	SystemPrompt string         `json:"system_prompt"`
	Capabilities map[string]any `json:"capabilities,omitempty"`
	Definition   map[string]any `json:"definition,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

func modeToWire(m domain.Mode) ModeWire {
	return ModeWire{
		ID:           m.ID.String(),
		BlueprintID:  m.BlueprintID.String(),
		Name:         m.Name,
		SystemPrompt: m.SystemPrompt,
		Capabilities: m.Capabilities,
		Definition:   m.Definition,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func modesToWire(in []domain.Mode) []ModeWire {
	out := make([]ModeWire, len(in))
	for i, m := range in {
		out[i] = modeToWire(m)
	}
	return out
}

// resolveBlueprintID returns the blueprint UUID for mode tools — either
// caller-supplied or the currently-selected blueprint.
func (s *Server) resolveBlueprintID(raw string) (uuid.UUID, error) {
	if raw != "" {
		bID, err := uuid.Parse(raw)
		if err != nil {
			return uuid.Nil, fmt.Errorf("blueprint_id: %w", err)
		}
		return bID, nil
	}
	sel := s.currentSelection()
	if sel.BlueprintID == nil {
		return uuid.Nil, errors.New("no blueprint selected; pass blueprint_id or call refresh(blueprint_id=...)")
	}
	return *sel.BlueprintID, nil
}

// mode.create / list / get / update / delete -----------------------------

type ModeCreateArgs struct {
	Name         string         `json:"name" jsonschema:"unique-within-blueprint mode name"`
	SystemPrompt string         `json:"system_prompt,omitempty"`
	Capabilities map[string]any `json:"capabilities,omitempty" jsonschema:"declared capability set (free-form JSON)"`
	Definition   map[string]any `json:"definition,omitempty" jsonschema:"additional mode metadata"`
	BlueprintID  string         `json:"blueprint_id,omitempty" jsonschema:"defaults to the currently-selected blueprint"`
}

type ModeCreateResult struct {
	Mode ModeWire `json:"mode"`
}

func (s *Server) handleModeCreate(ctx context.Context, _ *mcp.CallToolRequest, args ModeCreateArgs) (*mcp.CallToolResult, ModeCreateResult, error) {
	bID, err := s.resolveBlueprintID(args.BlueprintID)
	if err != nil {
		return nil, ModeCreateResult{}, err
	}
	m := domain.Mode{
		BlueprintID:  bID,
		Name:         args.Name,
		SystemPrompt: args.SystemPrompt,
		Capabilities: args.Capabilities,
		Definition:   args.Definition,
	}
	if a, ok := id.FromContext(ctx); ok {
		m.IdempotencyKey = a.IdempotencyKey
	}
	created, err := s.store.CreateMode(ctx, m)
	if err != nil {
		return nil, ModeCreateResult{}, err
	}
	return nil, ModeCreateResult{Mode: modeToWire(created)}, nil
}

type ModeListArgs struct {
	BlueprintID string `json:"blueprint_id,omitempty"`
}

type ModeListResult struct {
	Modes []ModeWire `json:"modes"`
	Count int        `json:"count"`
}

func (s *Server) handleModeList(ctx context.Context, _ *mcp.CallToolRequest, args ModeListArgs) (*mcp.CallToolResult, ModeListResult, error) {
	bID, err := s.resolveBlueprintID(args.BlueprintID)
	if err != nil {
		return nil, ModeListResult{}, err
	}
	list, err := s.store.ListModes(ctx, bID)
	if err != nil {
		return nil, ModeListResult{}, err
	}
	return nil, ModeListResult{Modes: modesToWire(list), Count: len(list)}, nil
}

type ModeGetArgs struct {
	ID          string `json:"id,omitempty"`
	BlueprintID string `json:"blueprint_id,omitempty"`
	Name        string `json:"name,omitempty"`
}

type ModeGetResult struct {
	Mode ModeWire `json:"mode"`
}

func (s *Server) handleModeGet(ctx context.Context, _ *mcp.CallToolRequest, args ModeGetArgs) (*mcp.CallToolResult, ModeGetResult, error) {
	if args.ID != "" {
		id, err := uuid.Parse(args.ID)
		if err != nil {
			return nil, ModeGetResult{}, fmt.Errorf("mode.get: id: %w", err)
		}
		m, err := s.store.GetMode(ctx, id)
		if err != nil {
			return nil, ModeGetResult{}, err
		}
		return nil, ModeGetResult{Mode: modeToWire(m)}, nil
	}
	if args.Name == "" {
		return nil, ModeGetResult{}, errors.New("mode.get: id or name required")
	}
	bID, err := s.resolveBlueprintID(args.BlueprintID)
	if err != nil {
		return nil, ModeGetResult{}, err
	}
	m, err := s.store.GetModeByName(ctx, bID, args.Name)
	if err != nil {
		return nil, ModeGetResult{}, err
	}
	return nil, ModeGetResult{Mode: modeToWire(m)}, nil
}

type ModeUpdateArgs struct {
	ID           string         `json:"id" jsonschema:"mode UUID"`
	Name         string         `json:"name,omitempty"`
	SystemPrompt string         `json:"system_prompt,omitempty"`
	Capabilities map[string]any `json:"capabilities,omitempty"`
	Definition   map[string]any `json:"definition,omitempty"`
}

type ModeUpdateResult struct {
	Mode ModeWire `json:"mode"`
}

func (s *Server) handleModeUpdate(ctx context.Context, _ *mcp.CallToolRequest, args ModeUpdateArgs) (*mcp.CallToolResult, ModeUpdateResult, error) {
	id, err := uuid.Parse(args.ID)
	if err != nil {
		return nil, ModeUpdateResult{}, fmt.Errorf("mode.update: id: %w", err)
	}
	var f pgstore.UpdateModeFields
	if args.Name != "" {
		name := args.Name
		f.Name = &name
	}
	if args.SystemPrompt != "" {
		sp := args.SystemPrompt
		f.SystemPrompt = &sp
	}
	if args.Capabilities != nil {
		caps := args.Capabilities
		f.Capabilities = &caps
	}
	if args.Definition != nil {
		def := args.Definition
		f.Definition = &def
	}
	m, err := s.store.UpdateMode(ctx, id, f)
	if err != nil {
		return nil, ModeUpdateResult{}, err
	}
	return nil, ModeUpdateResult{Mode: modeToWire(m)}, nil
}

type ModeDeleteArgs struct {
	ID string `json:"id"`
}

type ModeDeleteResult struct {
	Deleted bool   `json:"deleted"`
	ID      string `json:"id"`
}

func (s *Server) handleModeDelete(ctx context.Context, _ *mcp.CallToolRequest, args ModeDeleteArgs) (*mcp.CallToolResult, ModeDeleteResult, error) {
	id, err := uuid.Parse(args.ID)
	if err != nil {
		return nil, ModeDeleteResult{}, fmt.Errorf("mode.delete: id: %w", err)
	}
	if err := s.store.DeleteMode(ctx, id); err != nil {
		return nil, ModeDeleteResult{}, err
	}
	return nil, ModeDeleteResult{Deleted: true, ID: id.String()}, nil
}

// registerModes wires mode CRUD on srv (visible while a blueprint is selected).
func (s *Server) registerModes(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "mode.create",
		Description: "Create a mode inside the currently-selected blueprint version. The blueprint must be the latest version for its (project, name) — otherwise create a new version via blueprint.update first.",
	}, s.handleModeCreate)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "mode.list",
		Description: "List the modes inside the currently-selected blueprint version.",
	}, s.handleModeList)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "mode.get",
		Description: "Fetch a mode by id, or by name within the currently-selected blueprint.",
	}, s.handleModeGet)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "mode.update",
		Description: "Patch a mode's metadata. Only allowed on the latest blueprint version; otherwise call blueprint.update first.",
	}, s.handleModeUpdate)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "mode.delete",
		Description: "Delete a mode by id. Only allowed on the latest blueprint version.",
	}, s.handleModeDelete)
}
