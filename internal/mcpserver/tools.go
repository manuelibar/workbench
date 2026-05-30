package mcpserver

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *Server) registerCoreTools() {
	mcp.AddTool(s.sdkServer, &mcp.Tool{Name: "refresh", Description: "Synchronize selected scope and return latest context, resources, and dynamic capability links."}, s.handleRefresh)
	mcp.AddTool(s.sdkServer, &mcp.Tool{Name: "feedback", Description: "Report problems with a delivered skill/resource so Workbench can improve future context packs."}, s.handleFeedback)
	mcp.AddTool(s.sdkServer, &mcp.Tool{Name: "query", Description: "Query Workbench knowledge captured from feedback and notes."}, s.handleQuery)
}

func (s *Server) registerScopeTools(sel selection) {
	if len(s.dynamicTools) > 0 {
		s.sdkServer.RemoveTools(s.dynamicTools...)
		s.dynamicTools = s.dynamicTools[:0]
	}

	// Discovery and selection-management tools are useful before a project is selected.
	if sel.ProjectID == nil {
		mcp.AddTool(s.sdkServer, &mcp.Tool{Name: "project.create", Description: "Create a new project."}, s.handleProjectCreate)
		mcp.AddTool(s.sdkServer, &mcp.Tool{Name: "project.list", Description: "List all projects."}, s.handleProjectList)
		mcp.AddTool(s.sdkServer, &mcp.Tool{Name: "project.delete", Description: "Delete a project by ID."}, s.handleProjectDelete)
		mcp.AddTool(s.sdkServer, &mcp.Tool{Name: "namespace.create", Description: "Create a namespace (e.g. organization)."}, s.handleNamespaceCreate)
		mcp.AddTool(s.sdkServer, &mcp.Tool{Name: "namespace.list", Description: "List namespaces."}, s.handleNamespaceList)
		mcp.AddTool(s.sdkServer, &mcp.Tool{Name: "role.create", Description: "Create a role with optional system prompt."}, s.handleRoleCreate)
		mcp.AddTool(s.sdkServer, &mcp.Tool{Name: "role.list", Description: "List roles."}, s.handleRoleList)
		s.dynamicTools = append(s.dynamicTools, "project.create", "project.list", "project.delete", "namespace.create", "namespace.list", "role.create", "role.list")
		return
	}

	// Project-scoped tools appear only after refresh(project_id=...).
	mcp.AddTool(s.sdkServer, &mcp.Tool{Name: "project.list", Description: "List all projects."}, s.handleProjectList)
	mcp.AddTool(s.sdkServer, &mcp.Tool{Name: "namespace.list", Description: "List namespaces."}, s.handleNamespaceList)
	mcp.AddTool(s.sdkServer, &mcp.Tool{Name: "role.list", Description: "List roles."}, s.handleRoleList)
	mcp.AddTool(s.sdkServer, &mcp.Tool{Name: "task.create", Description: "Create a project task in proposed state."}, s.handleTaskCreate)
	mcp.AddTool(s.sdkServer, &mcp.Tool{Name: "task.list", Description: "List tracked project tasks."}, s.handleTaskList)
	mcp.AddTool(s.sdkServer, &mcp.Tool{Name: "task.transition", Description: "Move a task through the governed state machine."}, s.handleTaskTransition)
	mcp.AddTool(s.sdkServer, &mcp.Tool{Name: "board.create", Description: "Create a project board within a namespace."}, s.handleBoardCreate)
	mcp.AddTool(s.sdkServer, &mcp.Tool{Name: "board.list", Description: "List boards for the selected or supplied project."}, s.handleBoardList)
	s.dynamicTools = append(s.dynamicTools, "project.list", "namespace.list", "role.list", "task.create", "task.list", "task.transition", "board.create", "board.list")
}

// feedback

type feedbackIn struct {
	URI     string `json:"uri,omitempty"`
	Summary string `json:"summary"`
	Details string `json:"details,omitempty"`
}

type feedbackOut struct {
	Accepted bool   `json:"accepted"`
	Message  string `json:"message"`
}

func (s *Server) handleFeedback(_ context.Context, _ *mcp.CallToolRequest, in feedbackIn) (*mcp.CallToolResult, feedbackOut, error) {
	if in.Summary == "" {
		return nil, feedbackOut{}, fmt.Errorf("summary is required")
	}
	// Feedback is knowledge input: keep it queryable immediately, then future
	// reconcilers can promote it into tasks, skills, or docs.
	s.ingestFeedback(in.URI, in.Summary, in.Details)
	return nil, feedbackOut{Accepted: true, Message: "feedback stored as queryable knowledge"}, nil
}

// project.create

type projectCreateIn struct {
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	SystemPrompt string `json:"system_prompt,omitempty"`
}

type projectCreateOut struct {
	Project ProjectWire `json:"project"`
}

func (s *Server) handleProjectCreate(_ context.Context, _ *mcp.CallToolRequest, in projectCreateIn) (*mcp.CallToolResult, projectCreateOut, error) {
	if in.Name == "" {
		return nil, projectCreateOut{}, fmt.Errorf("name is required")
	}
	p, err := s.store.Create(in.Name, in.Description, in.SystemPrompt)
	if err != nil {
		return nil, projectCreateOut{}, err
	}
	return nil, projectCreateOut{Project: projectToWire(p)}, nil
}

// project.list

type projectListOut struct {
	Projects []ProjectWire `json:"projects"`
}

func (s *Server) handleProjectList(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, projectListOut, error) {
	projects, err := s.store.List()
	if err != nil {
		return nil, projectListOut{}, err
	}
	wires := make([]ProjectWire, len(projects))
	for i, p := range projects {
		wires[i] = projectToWire(p)
	}
	return nil, projectListOut{Projects: wires}, nil
}

// project.delete

type projectDeleteIn struct {
	ID string `json:"id"`
}
type projectDeleteOut struct {
	Deleted bool `json:"deleted"`
}

func (s *Server) handleProjectDelete(_ context.Context, _ *mcp.CallToolRequest, in projectDeleteIn) (*mcp.CallToolResult, projectDeleteOut, error) {
	id, err := uuid.Parse(in.ID)
	if err != nil {
		return nil, projectDeleteOut{}, fmt.Errorf("invalid id: %w", err)
	}
	deleted, err := s.store.Delete(id)
	if err != nil {
		return nil, projectDeleteOut{}, err
	}
	return nil, projectDeleteOut{Deleted: deleted}, nil
}

// namespace tools

type namespaceCreateIn struct {
	Name string `json:"name"`
}
type namespaceWire struct{ ID, Name, CreatedAt string }
type namespaceCreateOut struct {
	Namespace namespaceWire `json:"namespace"`
}
type namespaceListOut struct {
	Namespaces []namespaceWire `json:"namespaces"`
}

func (s *Server) handleNamespaceCreate(_ context.Context, _ *mcp.CallToolRequest, in namespaceCreateIn) (*mcp.CallToolResult, namespaceCreateOut, error) {
	if in.Name == "" {
		return nil, namespaceCreateOut{}, fmt.Errorf("name is required")
	}
	ns := Namespace{ID: uuid.New(), Name: in.Name, CreatedAt: time.Now().UTC()}
	s.mu.Lock()
	s.namespaces[ns.ID] = ns
	s.mu.Unlock()
	return nil, namespaceCreateOut{Namespace: namespaceWire{ID: ns.ID.String(), Name: ns.Name, CreatedAt: ns.CreatedAt.Format(time.RFC3339)}}, nil
}

func (s *Server) handleNamespaceList(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, namespaceListOut, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]namespaceWire, 0, len(s.namespaces))
	for _, ns := range s.namespaces {
		out = append(out, namespaceWire{ID: ns.ID.String(), Name: ns.Name, CreatedAt: ns.CreatedAt.Format(time.RFC3339)})
	}
	return nil, namespaceListOut{Namespaces: out}, nil
}

// role tools

type roleCreateIn struct{ Name, Description, SystemPrompt string }
type roleWire struct{ ID, Name, Description, SystemPrompt, CreatedAt string }
type roleCreateOut struct {
	Role roleWire `json:"role"`
}
type roleListOut struct {
	Roles []roleWire `json:"roles"`
}

func (s *Server) handleRoleCreate(_ context.Context, _ *mcp.CallToolRequest, in roleCreateIn) (*mcp.CallToolResult, roleCreateOut, error) {
	if in.Name == "" {
		return nil, roleCreateOut{}, fmt.Errorf("name is required")
	}
	r := Role{ID: uuid.New(), Name: in.Name, Description: in.Description, SystemPrompt: in.SystemPrompt, CreatedAt: time.Now().UTC()}
	s.mu.Lock()
	s.roles[r.ID] = r
	s.mu.Unlock()
	return nil, roleCreateOut{Role: roleWire{ID: r.ID.String(), Name: r.Name, Description: r.Description, SystemPrompt: r.SystemPrompt, CreatedAt: r.CreatedAt.Format(time.RFC3339)}}, nil
}

func (s *Server) handleRoleList(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, roleListOut, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]roleWire, 0, len(s.roles))
	for _, r := range s.roles {
		out = append(out, roleWire{ID: r.ID.String(), Name: r.Name, Description: r.Description, SystemPrompt: r.SystemPrompt, CreatedAt: r.CreatedAt.Format(time.RFC3339)})
	}
	return nil, roleListOut{Roles: out}, nil
}

// board tools

type boardCreateIn struct{ ProjectID, NamespaceID, Name, Description string }
type boardWire struct{ ID, ProjectID, NamespaceID, Name, Description, CreatedAt string }
type boardCreateOut struct {
	Board boardWire `json:"board"`
}
type boardListIn struct {
	ProjectID string `json:"project_id,omitempty"`
}
type boardListOut struct {
	Boards []boardWire `json:"boards"`
}

func (s *Server) handleBoardCreate(_ context.Context, _ *mcp.CallToolRequest, in boardCreateIn) (*mcp.CallToolResult, boardCreateOut, error) {
	if in.ProjectID == "" || in.NamespaceID == "" || in.Name == "" {
		return nil, boardCreateOut{}, fmt.Errorf("project_id, namespace_id, and name are required")
	}
	pid, err := uuid.Parse(in.ProjectID)
	if err != nil {
		return nil, boardCreateOut{}, fmt.Errorf("invalid project_id: %w", err)
	}
	nsid, err := uuid.Parse(in.NamespaceID)
	if err != nil {
		return nil, boardCreateOut{}, fmt.Errorf("invalid namespace_id: %w", err)
	}
	if _, err := s.store.Get(pid); err != nil {
		return nil, boardCreateOut{}, fmt.Errorf("project not found")
	}
	s.mu.Lock()
	_, ok := s.namespaces[nsid]
	s.mu.Unlock()
	if !ok {
		return nil, boardCreateOut{}, fmt.Errorf("namespace not found")
	}
	b := Board{ID: uuid.New(), ProjectID: pid, NamespaceID: nsid, Name: in.Name, Description: in.Description, CreatedAt: time.Now().UTC()}
	s.mu.Lock()
	s.boards[b.ID] = b
	s.mu.Unlock()
	return nil, boardCreateOut{Board: boardWire{ID: b.ID.String(), ProjectID: b.ProjectID.String(), NamespaceID: b.NamespaceID.String(), Name: b.Name, Description: b.Description, CreatedAt: b.CreatedAt.Format(time.RFC3339)}}, nil
}

func (s *Server) handleBoardList(_ context.Context, _ *mcp.CallToolRequest, in boardListIn) (*mcp.CallToolResult, boardListOut, error) {
	var filter *uuid.UUID
	if in.ProjectID != "" {
		pid, err := uuid.Parse(in.ProjectID)
		if err != nil {
			return nil, boardListOut{}, fmt.Errorf("invalid project_id: %w", err)
		}
		filter = &pid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []boardWire{}
	for _, b := range s.boards {
		if filter != nil && b.ProjectID != *filter {
			continue
		}
		out = append(out, boardWire{ID: b.ID.String(), ProjectID: b.ProjectID.String(), NamespaceID: b.NamespaceID.String(), Name: b.Name, Description: b.Description, CreatedAt: b.CreatedAt.Format(time.RFC3339)})
	}
	return nil, boardListOut{Boards: out}, nil
}
