package domain

import (
	"time"

	"github.com/google/uuid"
)

// Namespace is a tree-shaped organisational container. The root nodes have
// nil [Namespace.ParentID]; arbitrary depth is allowed. (name, parent_id)
// is unique within siblings, with NULL parent treated as a single
// equivalence class.
type Namespace struct {
	ID             uuid.UUID
	ParentID       *uuid.UUID
	Name           string
	Description    string
	Settings       map[string]any
	IdempotencyKey string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
