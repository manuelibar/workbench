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

// NamespaceWire is the JSON shape namespace tools return over MCP.
type NamespaceWire struct {
	ID          string         `json:"id"`
	ParentID    string         `json:"parent_id,omitempty"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Settings    map[string]any `json:"settings,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

func namespaceToWire(n domain.Namespace) NamespaceWire {
	w := NamespaceWire{
		ID:          n.ID.String(),
		Name:        n.Name,
		Description: n.Description,
		Settings:    n.Settings,
		CreatedAt:   n.CreatedAt,
		UpdatedAt:   n.UpdatedAt,
	}
	if n.ParentID != nil {
		w.ParentID = n.ParentID.String()
	}
	return w
}

func namespacesToWire(ns []domain.Namespace) []NamespaceWire {
	out := make([]NamespaceWire, len(ns))
	for i, n := range ns {
		out[i] = namespaceToWire(n)
	}
	return out
}

// namespace.create -------------------------------------------------------

type NamespaceCreateArgs struct {
	Name        string         `json:"name" jsonschema:"namespace name (must be unique among siblings)"`
	ParentID    string         `json:"parent_id,omitempty" jsonschema:"parent namespace UUID; omit for a root namespace"`
	Description string         `json:"description,omitempty" jsonschema:"optional human-readable description"`
	Settings    map[string]any `json:"settings,omitempty" jsonschema:"optional free-form settings JSON object"`
}

type NamespaceCreateResult struct {
	Namespace NamespaceWire `json:"namespace"`
}

func (s *Server) handleNamespaceCreate(ctx context.Context, _ *mcp.CallToolRequest, args NamespaceCreateArgs) (*mcp.CallToolResult, NamespaceCreateResult, error) {
	if args.Name == "" {
		return nil, NamespaceCreateResult{}, fmt.Errorf("namespace.create: name must not be empty")
	}
	n := domain.Namespace{
		Name:        args.Name,
		Description: args.Description,
		Settings:    args.Settings,
	}
	if args.ParentID != "" {
		pID, err := uuid.Parse(args.ParentID)
		if err != nil {
			return nil, NamespaceCreateResult{}, fmt.Errorf("namespace.create: parent_id: %w", err)
		}
		n.ParentID = &pID
	}
	if a, ok := id.FromContext(ctx); ok {
		n.IdempotencyKey = a.IdempotencyKey
	}
	created, err := s.store.CreateNamespace(ctx, n)
	if err != nil {
		return nil, NamespaceCreateResult{}, err
	}
	return nil, NamespaceCreateResult{Namespace: namespaceToWire(created)}, nil
}

// namespace.list ---------------------------------------------------------

type NamespaceListArgs struct {
	ParentID string `json:"parent_id,omitempty" jsonschema:"list children of this namespace UUID; omit for root namespaces"`
}

type NamespaceListResult struct {
	Namespaces []NamespaceWire `json:"namespaces"`
	Count      int             `json:"count"`
}

func (s *Server) handleNamespaceList(ctx context.Context, _ *mcp.CallToolRequest, args NamespaceListArgs) (*mcp.CallToolResult, NamespaceListResult, error) {
	var parent *uuid.UUID
	if args.ParentID != "" {
		pID, err := uuid.Parse(args.ParentID)
		if err != nil {
			return nil, NamespaceListResult{}, fmt.Errorf("namespace.list: parent_id: %w", err)
		}
		parent = &pID
	}
	list, err := s.store.ListNamespaces(ctx, parent)
	if err != nil {
		return nil, NamespaceListResult{}, err
	}
	return nil, NamespaceListResult{Namespaces: namespacesToWire(list), Count: len(list)}, nil
}

// namespace.get ----------------------------------------------------------

type NamespaceGetArgs struct {
	ID string `json:"id,omitempty" jsonschema:"namespace UUID; defaults to the currently-selected namespace"`
}

type NamespaceGetResult struct {
	Namespace NamespaceWire `json:"namespace"`
}

func (s *Server) handleNamespaceGet(ctx context.Context, _ *mcp.CallToolRequest, args NamespaceGetArgs) (*mcp.CallToolResult, NamespaceGetResult, error) {
	id, err := s.resolveNamespaceID(args.ID)
	if err != nil {
		return nil, NamespaceGetResult{}, err
	}
	n, err := s.store.GetNamespace(ctx, id)
	if err != nil {
		return nil, NamespaceGetResult{}, err
	}
	return nil, NamespaceGetResult{Namespace: namespaceToWire(n)}, nil
}

// namespace.update -------------------------------------------------------

type NamespaceUpdateArgs struct {
	ID          string         `json:"id,omitempty"          jsonschema:"namespace UUID; defaults to the currently-selected namespace"`
	Name        string         `json:"name,omitempty"        jsonschema:"new name; omit to keep existing"`
	Description string         `json:"description,omitempty" jsonschema:"new description; omit to keep existing"`
	Settings    map[string]any `json:"settings,omitempty"    jsonschema:"replace settings; omit to keep existing"`
}

type NamespaceUpdateResult struct {
	Namespace NamespaceWire `json:"namespace"`
}

func (s *Server) handleNamespaceUpdate(ctx context.Context, _ *mcp.CallToolRequest, args NamespaceUpdateArgs) (*mcp.CallToolResult, NamespaceUpdateResult, error) {
	id, err := s.resolveNamespaceID(args.ID)
	if err != nil {
		return nil, NamespaceUpdateResult{}, err
	}
	var f pgstore.UpdateNamespaceFields
	if args.Name != "" {
		name := args.Name
		f.Name = &name
	}
	if args.Description != "" {
		desc := args.Description
		f.Description = &desc
	}
	if args.Settings != nil {
		settings := args.Settings
		f.Settings = &settings
	}
	n, err := s.store.UpdateNamespace(ctx, id, f)
	if err != nil {
		return nil, NamespaceUpdateResult{}, err
	}
	return nil, NamespaceUpdateResult{Namespace: namespaceToWire(n)}, nil
}

// namespace.delete -------------------------------------------------------

type NamespaceDeleteArgs struct {
	ID string `json:"id,omitempty" jsonschema:"namespace UUID; defaults to the currently-selected namespace"`
}

type NamespaceDeleteResult struct {
	Deleted bool   `json:"deleted"`
	ID      string `json:"id"`
}

func (s *Server) handleNamespaceDelete(ctx context.Context, _ *mcp.CallToolRequest, args NamespaceDeleteArgs) (*mcp.CallToolResult, NamespaceDeleteResult, error) {
	id, err := s.resolveNamespaceID(args.ID)
	if err != nil {
		return nil, NamespaceDeleteResult{}, err
	}
	// If we're deleting the currently-selected namespace, also clear the
	// selection so the surface re-collapses to bootstrap-only on next refresh.
	clearSelection := false
	if sel := s.currentSelection(); sel.NamespaceID != nil && *sel.NamespaceID == id {
		clearSelection = true
	}
	if err := s.store.DeleteNamespace(ctx, id); err != nil {
		return nil, NamespaceDeleteResult{}, err
	}
	if clearSelection {
		// Best-effort: persist empty selection so refresh picks it up.
		_ = s.store.UpdateSelection(ctx, s.user.ID, domain.Selection{})
		s.selectionMu.Lock()
		s.selection = domain.Selection{}
		s.applyVisibility(s.selection)
		s.selectionMu.Unlock()
	}
	return nil, NamespaceDeleteResult{Deleted: true, ID: id.String()}, nil
}

// resolveNamespaceID parses raw, falling back to the currently-selected
// namespace when raw is empty. Returns an error if neither is available.
func (s *Server) resolveNamespaceID(raw string) (uuid.UUID, error) {
	if raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return uuid.Nil, fmt.Errorf("id: %w", err)
		}
		return id, nil
	}
	sel := s.currentSelection()
	if sel.NamespaceID == nil {
		return uuid.Nil, errors.New("no namespace selected; pass id or call refresh(namespace_id=...)")
	}
	return *sel.NamespaceID, nil
}

// registerNamespaceBootstrap registers the namespace tools that are visible
// regardless of selection: namespace.create and namespace.list.
func (s *Server) registerNamespaceBootstrap(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "namespace.create",
		Description: "Create a new namespace. Optional parent_id places it under a parent; omit for a root namespace. Idempotent if the request supplies an idempotency key.",
	}, s.handleNamespaceCreate)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "namespace.list",
		Description: "List child namespaces of parent_id. Omit parent_id to list root namespaces.",
	}, s.handleNamespaceList)
}

// registerNamespaceScoped registers the namespace tools that are visible
// only when a namespace is currently selected: namespace.get / update / delete.
func (s *Server) registerNamespaceScoped(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "namespace.get",
		Description: "Fetch a namespace. id defaults to the currently-selected namespace.",
	}, s.handleNamespaceGet)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "namespace.update",
		Description: "Patch a namespace's name / description / settings. id defaults to the currently-selected namespace.",
	}, s.handleNamespaceUpdate)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "namespace.delete",
		Description: "Delete a namespace. Children cascade. If the deleted namespace is currently selected, the selection is cleared.",
	}, s.handleNamespaceDelete)
}
