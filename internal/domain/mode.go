package domain

import (
	"time"

	"github.com/google/uuid"
)

// Mode is a self-managed container nested inside a [Blueprint] version: a
// system prompt plus a declared set of capabilities (which always-on /
// scoped tools the agent should treat as visible, plus user-defined
// tools/resources/prompts the mode wants to surface).
//
// The mode is bound to a specific [Mode.BlueprintID] (i.e. one blueprint
// version). Modes are mutable only when their blueprint version is the
// latest for `(project, blueprint_name)`; otherwise a new blueprint version
// must be created first.
type Mode struct {
	ID             uuid.UUID
	BlueprintID    uuid.UUID
	Name           string
	SystemPrompt   string
	Capabilities   map[string]any
	Definition     map[string]any
	IdempotencyKey string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
