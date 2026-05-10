package domain

import (
	"time"

	"github.com/google/uuid"
)

// Blueprint is the versioned, immutable composition root for a Run. v0
// stores the definition as a JSON blob; the manifesto's full Blueprint YAML
// schema is honoured purely as content (the workbench doesn't validate the
// shape).
//
// Updates write a new row with `version = MAX(version) + 1`; existing rows
// are never modified in place.
type Blueprint struct {
	ID             uuid.UUID
	ProjectID      uuid.UUID
	Name           string
	Version        int
	Definition     map[string]any
	IdempotencyKey string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
