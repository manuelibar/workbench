package mcpserver

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type taskWire struct {
	ID          string   `json:"id"`
	ProjectID   string   `json:"project_id"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	State       string   `json:"state"`
	Evidence    []string `json:"evidence,omitempty"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

type taskCreateIn struct{ ProjectID, Title, Description string }
type taskCreateOut struct {
	Task taskWire `json:"task"`
}

type taskTransitionIn struct {
	ID       string   `json:"id"`
	State    string   `json:"state"`
	Evidence []string `json:"evidence,omitempty"`
}
type taskTransitionOut struct {
	Task taskWire `json:"task"`
}

type taskListOut struct {
	Tasks []taskWire `json:"tasks"`
}

func (s *Server) handleTaskCreate(_ context.Context, _ *mcp.CallToolRequest, in taskCreateIn) (*mcp.CallToolResult, taskCreateOut, error) {
	if in.ProjectID == "" || in.Title == "" {
		return nil, taskCreateOut{}, fmt.Errorf("project_id and title are required")
	}
	pid, err := uuid.Parse(in.ProjectID)
	if err != nil {
		return nil, taskCreateOut{}, fmt.Errorf("invalid project_id: %w", err)
	}
	if _, err := s.store.Get(pid); err != nil {
		return nil, taskCreateOut{}, fmt.Errorf("project not found")
	}
	now := time.Now().UTC()
	task := Task{ID: uuid.New(), ProjectID: pid, Title: in.Title, Description: in.Description, State: TaskProposed, Evidence: []string{}, CreatedAt: now, UpdatedAt: now}
	s.mu.Lock()
	s.tasks[task.ID] = task
	s.mu.Unlock()
	return nil, taskCreateOut{Task: taskToWire(task)}, nil
}

func (s *Server) handleTaskTransition(_ context.Context, _ *mcp.CallToolRequest, in taskTransitionIn) (*mcp.CallToolResult, taskTransitionOut, error) {
	id, err := uuid.Parse(in.ID)
	if err != nil {
		return nil, taskTransitionOut{}, fmt.Errorf("invalid id: %w", err)
	}
	to := TaskState(in.State)
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasks[id]
	if !ok {
		return nil, taskTransitionOut{}, fmt.Errorf("task not found")
	}
	if !CanTransitionTask(task.State, to) {
		return nil, taskTransitionOut{}, fmt.Errorf("invalid task transition: %s -> %s", task.State, to)
	}
	task.State = to
	task.Evidence = append(task.Evidence, in.Evidence...)
	task.UpdatedAt = time.Now().UTC()
	s.tasks[id] = task
	return nil, taskTransitionOut{Task: taskToWire(task)}, nil
}

func (s *Server) handleTaskList(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, taskListOut, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []taskWire{}
	for _, task := range s.tasks {
		out = append(out, taskToWire(task))
	}
	return nil, taskListOut{Tasks: out}, nil
}

func (s *Server) handleTasksResource(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []taskWire{}
	for _, task := range s.tasks {
		out = append(out, taskToWire(task))
	}
	return jsonResource(req.Params.URI, map[string]any{"tasks": out})
}

func taskToWire(t Task) taskWire {
	return taskWire{ID: t.ID.String(), ProjectID: t.ProjectID.String(), Title: t.Title, Description: t.Description, State: string(t.State), Evidence: t.Evidence, CreatedAt: t.CreatedAt.Format(time.RFC3339), UpdatedAt: t.UpdatedAt.Format(time.RFC3339)}
}
