package mcpserver

import (
	"time"

	"github.com/google/uuid"
)

type Namespace struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Role struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	SystemPrompt string    `json:"system_prompt"`
	CreatedAt    time.Time `json:"created_at"`
}

type Board struct {
	ID          uuid.UUID `json:"id"`
	ProjectID   uuid.UUID `json:"project_id"`
	NamespaceID uuid.UUID `json:"namespace_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type GitHubConfig struct {
	Organization string `json:"organization"`
	Token        string `json:"token,omitempty"`
}

type TaskState string

const (
	TaskProposed   TaskState = "proposed"
	TaskReady      TaskState = "ready"
	TaskInProgress TaskState = "in_progress"
	TaskBlocked    TaskState = "blocked"
	TaskReview     TaskState = "review"
	TaskDone       TaskState = "done"
	TaskCancelled  TaskState = "cancelled"
)

type Task struct {
	ID          uuid.UUID `json:"id"`
	ProjectID   uuid.UUID `json:"project_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	State       TaskState `json:"state"`
	Evidence    []string  `json:"evidence"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type KnowledgeItem struct {
	ID        uuid.UUID `json:"id"`
	Kind      string    `json:"kind"`
	URI       string    `json:"uri,omitempty"`
	Summary   string    `json:"summary"`
	Details   string    `json:"details,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func CanTransitionTask(from, to TaskState) bool {
	allowed := map[TaskState][]TaskState{
		TaskProposed:   {TaskReady, TaskCancelled},
		TaskReady:      {TaskInProgress, TaskBlocked, TaskCancelled},
		TaskInProgress: {TaskBlocked, TaskReview, TaskCancelled},
		TaskBlocked:    {TaskReady, TaskInProgress, TaskCancelled},
		TaskReview:     {TaskInProgress, TaskDone, TaskCancelled},
		TaskDone:       {},
		TaskCancelled:  {},
	}
	for _, candidate := range allowed[from] {
		if candidate == to {
			return true
		}
	}
	return false
}
