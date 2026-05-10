package domain

import (
	"time"

	"github.com/google/uuid"
)

// Prompt is a project-scoped prompt template. Body may contain `{{var}}`
// placeholders that map to entries in [Prompt.Args]; v0 stores both
// verbatim without any rendering — interpolation is the agent's
// responsibility.
type Prompt struct {
	ID             uuid.UUID
	ProjectID      uuid.UUID
	Name           string
	Description    string
	Body           string
	Args           []PromptArg
	IdempotencyKey string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// PromptArg matches the MCP wire shape for prompt arguments.
type PromptArg struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}
