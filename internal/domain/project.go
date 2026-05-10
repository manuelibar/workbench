package domain

import (
	"time"

	"github.com/google/uuid"
)

// Project is the leaf of the namespace hierarchy — the unit that maps to a
// concrete codebase, repo, or scope of work. A project may live under a
// namespace ([Project.NamespaceID] non-nil) or be standalone (nil).
type Project struct {
	ID             uuid.UUID
	NamespaceID    *uuid.UUID
	Name           string
	Description    string
	Settings       map[string]any
	IdempotencyKey string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
