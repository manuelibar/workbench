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

// ProjectWire is the JSON shape project tools return over MCP.
type ProjectWire struct {
	ID          string         `json:"id"`
	NamespaceID string         `json:"namespace_id,omitempty"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Settings    map[string]any `json:"settings,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

func projectToWire(p domain.Project) ProjectWire {
	w := ProjectWire{
		ID:          p.ID.String(),
		Name:        p.Name,
		Description: p.Description,
		Settings:    p.Settings,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
	if p.NamespaceID != nil {
		w.NamespaceID = p.NamespaceID.String()
	}
	return w
}

func projectsToWire(ps []domain.Project) []ProjectWire {
	out := make([]ProjectWire, len(ps))
	for i, p := range ps {
		out[i] = projectToWire(p)
	}
	return out
}

// project.create ---------------------------------------------------------

type ProjectCreateArgs struct {
	Name        string         `json:"name" jsonschema:"project name (unique within its namespace)"`
	NamespaceID string         `json:"namespace_id,omitempty" jsonschema:"namespace UUID; defaults to the currently-selected namespace"`
	Description string         `json:"description,omitempty"`
	Settings    map[string]any `json:"settings,omitempty"`
}

type ProjectCreateResult struct {
	Project ProjectWire `json:"project"`
}

func (s *Server) handleProjectCreate(ctx context.Context, _ *mcp.CallToolRequest, args ProjectCreateArgs) (*mcp.CallToolResult, ProjectCreateResult, error) {
	if args.Name == "" {
		return nil, ProjectCreateResult{}, fmt.Errorf("project.create: name must not be empty")
	}
	nsID, err := s.resolveProjectNamespaceID(args.NamespaceID)
	if err != nil {
		return nil, ProjectCreateResult{}, err
	}
	p := domain.Project{
		NamespaceID: &nsID,
		Name:        args.Name,
		Description: args.Description,
		Settings:    args.Settings,
	}
	if a, ok := id.FromContext(ctx); ok {
		p.IdempotencyKey = a.IdempotencyKey
	}
	created, err := s.store.CreateProject(ctx, p)
	if err != nil {
		return nil, ProjectCreateResult{}, err
	}
	return nil, ProjectCreateResult{Project: projectToWire(created)}, nil
}

// project.list -----------------------------------------------------------

type ProjectListArgs struct {
	NamespaceID string `json:"namespace_id,omitempty" jsonschema:"namespace UUID; defaults to the currently-selected namespace"`
}

type ProjectListResult struct {
	Projects []ProjectWire `json:"projects"`
	Count    int           `json:"count"`
}

func (s *Server) handleProjectList(ctx context.Context, _ *mcp.CallToolRequest, args ProjectListArgs) (*mcp.CallToolResult, ProjectListResult, error) {
	nsID, err := s.resolveProjectNamespaceID(args.NamespaceID)
	if err != nil {
		return nil, ProjectListResult{}, err
	}
	list, err := s.store.ListProjects(ctx, &nsID)
	if err != nil {
		return nil, ProjectListResult{}, err
	}
	return nil, ProjectListResult{Projects: projectsToWire(list), Count: len(list)}, nil
}

// project.get / update / delete (require a project to be selected) -------

type ProjectGetArgs struct {
	ID string `json:"id,omitempty" jsonschema:"project UUID; defaults to the currently-selected project"`
}

type ProjectGetResult struct {
	Project ProjectWire `json:"project"`
}

func (s *Server) handleProjectGet(ctx context.Context, _ *mcp.CallToolRequest, args ProjectGetArgs) (*mcp.CallToolResult, ProjectGetResult, error) {
	id, err := s.resolveProjectID(args.ID)
	if err != nil {
		return nil, ProjectGetResult{}, err
	}
	p, err := s.store.GetProject(ctx, id)
	if err != nil {
		return nil, ProjectGetResult{}, err
	}
	return nil, ProjectGetResult{Project: projectToWire(p)}, nil
}

type ProjectUpdateArgs struct {
	ID          string         `json:"id,omitempty"`
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
	Settings    map[string]any `json:"settings,omitempty"`
}

type ProjectUpdateResult struct {
	Project ProjectWire `json:"project"`
}

func (s *Server) handleProjectUpdate(ctx context.Context, _ *mcp.CallToolRequest, args ProjectUpdateArgs) (*mcp.CallToolResult, ProjectUpdateResult, error) {
	id, err := s.resolveProjectID(args.ID)
	if err != nil {
		return nil, ProjectUpdateResult{}, err
	}
	var f pgstore.UpdateProjectFields
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
	p, err := s.store.UpdateProject(ctx, id, f)
	if err != nil {
		return nil, ProjectUpdateResult{}, err
	}
	return nil, ProjectUpdateResult{Project: projectToWire(p)}, nil
}

type ProjectDeleteArgs struct {
	ID string `json:"id,omitempty"`
}

type ProjectDeleteResult struct {
	Deleted bool   `json:"deleted"`
	ID      string `json:"id"`
}

func (s *Server) handleProjectDelete(ctx context.Context, _ *mcp.CallToolRequest, args ProjectDeleteArgs) (*mcp.CallToolResult, ProjectDeleteResult, error) {
	id, err := s.resolveProjectID(args.ID)
	if err != nil {
		return nil, ProjectDeleteResult{}, err
	}
	clearProject := false
	if sel := s.currentSelection(); sel.ProjectID != nil && *sel.ProjectID == id {
		clearProject = true
	}
	if err := s.store.DeleteProject(ctx, id); err != nil {
		return nil, ProjectDeleteResult{}, err
	}
	if clearProject {
		s.selectionMu.Lock()
		s.selection.ProjectID = nil
		s.selection.BlueprintID = nil
		s.selection.ModeName = ""
		_ = s.store.UpdateSelection(ctx, s.user.ID, s.selection)
		s.applyVisibility(s.selection)
		s.selectionMu.Unlock()
	}
	return nil, ProjectDeleteResult{Deleted: true, ID: id.String()}, nil
}

// resolveProjectNamespaceID: namespace UUID for project bootstrap tools.
func (s *Server) resolveProjectNamespaceID(raw string) (uuid.UUID, error) {
	if raw != "" {
		nsID, err := uuid.Parse(raw)
		if err != nil {
			return uuid.Nil, fmt.Errorf("namespace_id: %w", err)
		}
		return nsID, nil
	}
	sel := s.currentSelection()
	if sel.NamespaceID == nil {
		return uuid.Nil, errors.New("no namespace selected; pass namespace_id or call refresh(namespace_id=...)")
	}
	return *sel.NamespaceID, nil
}

// resolveProjectID: project UUID for project-scoped tools.
func (s *Server) resolveProjectID(raw string) (uuid.UUID, error) {
	if raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return uuid.Nil, fmt.Errorf("id: %w", err)
		}
		return id, nil
	}
	sel := s.currentSelection()
	if sel.ProjectID == nil {
		return uuid.Nil, errors.New("no project selected; pass id or call refresh(project_id=...)")
	}
	return *sel.ProjectID, nil
}

// registerProjectBootstrap registers project.create + project.list (visible
// when a namespace is selected).
func (s *Server) registerProjectBootstrap(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "project.create",
		Description: "Create a project under a namespace. namespace_id defaults to the currently-selected namespace.",
	}, s.handleProjectCreate)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "project.list",
		Description: "List projects in a namespace. namespace_id defaults to the currently-selected namespace.",
	}, s.handleProjectList)
}

// registerProjectFull registers project.get / update / delete (visible only
// when a project is selected).
func (s *Server) registerProjectFull(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "project.get",
		Description: "Fetch a project. id defaults to the currently-selected project.",
	}, s.handleProjectGet)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "project.update",
		Description: "Patch a project's name / description / settings. id defaults to the currently-selected project.",
	}, s.handleProjectUpdate)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "project.delete",
		Description: "Delete a project. Cascades to artifacts/skills/prompts. If the deleted project is currently selected, the project / blueprint / mode parts of the selection are cleared.",
	}, s.handleProjectDelete)
}
