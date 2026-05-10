package domain

import (
	"time"

	"github.com/google/uuid"
)

// Skill is a project-scoped procedural-knowledge document, surfaced to the
// agent as MCP resources. v0 stores the markdown body verbatim; later
// phases may add tagging, vector embeddings, or multi-version history.
type Skill struct {
	ID             uuid.UUID
	ProjectID      uuid.UUID
	Name           string
	BodyMD         string
	IdempotencyKey string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
