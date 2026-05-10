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

// BlueprintWire is the JSON shape blueprint tools return.
type BlueprintWire struct {
	ID         string         `json:"id"`
	ProjectID  string         `json:"project_id"`
	Name       string         `json:"name"`
	Version    int            `json:"version"`
	Definition map[string]any `json:"definition,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

func blueprintToWire(b domain.Blueprint) BlueprintWire {
	return BlueprintWire{
		ID:         b.ID.String(),
		ProjectID:  b.ProjectID.String(),
		Name:       b.Name,
		Version:    b.Version,
		Definition: b.Definition,
		CreatedAt:  b.CreatedAt,
		UpdatedAt:  b.UpdatedAt,
	}
}

func blueprintsToWire(in []domain.Blueprint) []BlueprintWire {
	out := make([]BlueprintWire, len(in))
	for i, b := range in {
		out[i] = blueprintToWire(b)
	}
	return out
}

// blueprint.create -------------------------------------------------------

type BlueprintCreateArgs struct {
	Name       string         `json:"name" jsonschema:"blueprint name (unique within project; first call creates v1)"`
	Definition map[string]any `json:"definition,omitempty" jsonschema:"YAML-shaped blueprint definition (any JSON-marshallable object)"`
}

type BlueprintCreateResult struct {
	Blueprint BlueprintWire `json:"blueprint"`
}

func (s *Server) handleBlueprintCreate(ctx context.Context, _ *mcp.CallToolRequest, args BlueprintCreateArgs) (*mcp.CallToolResult, BlueprintCreateResult, error) {
	pID, err := s.resolveProjectID("")
	if err != nil {
		return nil, BlueprintCreateResult{}, err
	}
	b := domain.Blueprint{ProjectID: pID, Name: args.Name, Definition: args.Definition}
	if a, ok := id.FromContext(ctx); ok {
		b.IdempotencyKey = a.IdempotencyKey
	}
	created, err := s.store.CreateBlueprint(ctx, b)
	if err != nil {
		return nil, BlueprintCreateResult{}, err
	}
	return nil, BlueprintCreateResult{Blueprint: blueprintToWire(created)}, nil
}

// blueprint.list ---------------------------------------------------------

type BlueprintListArgs struct {
	LatestOnly bool `json:"latest_only,omitempty" jsonschema:"return only the highest-version row per (project, name)"`
}

type BlueprintListResult struct {
	Blueprints []BlueprintWire `json:"blueprints"`
	Count      int             `json:"count"`
}

func (s *Server) handleBlueprintList(ctx context.Context, _ *mcp.CallToolRequest, args BlueprintListArgs) (*mcp.CallToolResult, BlueprintListResult, error) {
	pID, err := s.resolveProjectID("")
	if err != nil {
		return nil, BlueprintListResult{}, err
	}
	list, err := s.store.ListBlueprints(ctx, pID, pgstore.ListBlueprintsFilter{LatestOnly: args.LatestOnly})
	if err != nil {
		return nil, BlueprintListResult{}, err
	}
	return nil, BlueprintListResult{Blueprints: blueprintsToWire(list), Count: len(list)}, nil
}

// blueprint.get ----------------------------------------------------------

type BlueprintGetArgs struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Version int    `json:"version,omitempty" jsonschema:"specific version; 0 or omitted = latest"`
}

type BlueprintGetResult struct {
	Blueprint BlueprintWire `json:"blueprint"`
}

func (s *Server) handleBlueprintGet(ctx context.Context, _ *mcp.CallToolRequest, args BlueprintGetArgs) (*mcp.CallToolResult, BlueprintGetResult, error) {
	if args.ID != "" {
		id, err := uuid.Parse(args.ID)
		if err != nil {
			return nil, BlueprintGetResult{}, fmt.Errorf("blueprint.get: id: %w", err)
		}
		b, err := s.store.GetBlueprint(ctx, id)
		if err != nil {
			return nil, BlueprintGetResult{}, err
		}
		return nil, BlueprintGetResult{Blueprint: blueprintToWire(b)}, nil
	}
	if args.Name == "" {
		return nil, BlueprintGetResult{}, errors.New("blueprint.get: id or name required")
	}
	pID, err := s.resolveProjectID("")
	if err != nil {
		return nil, BlueprintGetResult{}, err
	}
	var b domain.Blueprint
	if args.Version > 0 {
		b, err = s.store.GetBlueprintByVersion(ctx, pID, args.Name, args.Version)
	} else {
		b, err = s.store.GetLatestBlueprint(ctx, pID, args.Name)
	}
	if err != nil {
		return nil, BlueprintGetResult{}, err
	}
	return nil, BlueprintGetResult{Blueprint: blueprintToWire(b)}, nil
}

// blueprint.update -------------------------------------------------------

type BlueprintUpdateArgs struct {
	Name       string         `json:"name" jsonschema:"blueprint name (server reads MAX(version) and appends a new row)"`
	Definition map[string]any `json:"definition,omitempty"`
}

type BlueprintUpdateResult struct {
	Blueprint BlueprintWire `json:"blueprint"`
}

func (s *Server) handleBlueprintUpdate(ctx context.Context, _ *mcp.CallToolRequest, args BlueprintUpdateArgs) (*mcp.CallToolResult, BlueprintUpdateResult, error) {
	if args.Name == "" {
		return nil, BlueprintUpdateResult{}, errors.New("blueprint.update: name required")
	}
	pID, err := s.resolveProjectID("")
	if err != nil {
		return nil, BlueprintUpdateResult{}, err
	}
	b, err := s.store.AppendBlueprintVersion(ctx, pID, args.Name, args.Definition)
	if err != nil {
		return nil, BlueprintUpdateResult{}, err
	}
	return nil, BlueprintUpdateResult{Blueprint: blueprintToWire(b)}, nil
}

// blueprint.delete -------------------------------------------------------

type BlueprintDeleteArgs struct {
	ID string `json:"id"`
}

type BlueprintDeleteResult struct {
	Deleted bool   `json:"deleted"`
	ID      string `json:"id"`
}

func (s *Server) handleBlueprintDelete(ctx context.Context, _ *mcp.CallToolRequest, args BlueprintDeleteArgs) (*mcp.CallToolResult, BlueprintDeleteResult, error) {
	id, err := uuid.Parse(args.ID)
	if err != nil {
		return nil, BlueprintDeleteResult{}, fmt.Errorf("blueprint.delete: id: %w", err)
	}
	clearBlueprint := false
	if sel := s.currentSelection(); sel.BlueprintID != nil && *sel.BlueprintID == id {
		clearBlueprint = true
	}
	if err := s.store.DeleteBlueprint(ctx, id); err != nil {
		return nil, BlueprintDeleteResult{}, err
	}
	if clearBlueprint {
		s.selectionMu.Lock()
		s.selection.BlueprintID = nil
		s.selection.ModeName = ""
		_ = s.store.UpdateSelection(ctx, s.user.ID, s.selection)
		s.applyVisibility(s.selection)
		s.selectionMu.Unlock()
	}
	return nil, BlueprintDeleteResult{Deleted: true, ID: id.String()}, nil
}

// registerBlueprints wires blueprint CRUD on srv (visible while a project
// is selected).
func (s *Server) registerBlueprints(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "blueprint.create",
		Description: "Create a brand-new blueprint at version 1 in the currently-selected project.",
	}, s.handleBlueprintCreate)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "blueprint.list",
		Description: "List blueprint rows in the currently-selected project. latest_only=true collapses to one row per (name).",
	}, s.handleBlueprintList)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "blueprint.get",
		Description: "Fetch a blueprint by id, or by name (latest version unless version is supplied).",
	}, s.handleBlueprintGet)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "blueprint.update",
		Description: "Append a new immutable version of an existing blueprint name (server-monotonic version).",
	}, s.handleBlueprintUpdate)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "blueprint.delete",
		Description: "Delete a single blueprint version (and its modes). Other versions are untouched.",
	}, s.handleBlueprintDelete)
}
