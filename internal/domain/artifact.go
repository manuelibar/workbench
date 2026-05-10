package domain

import (
	"time"

	"github.com/google/uuid"
)

// Artifact lifecycle status. Free-form string at the schema level so future
// types can extend without a migration; the v0 enum is enforced at the SQL
// layer via a CHECK constraint.
const (
	ArtifactStatusDraft     = "draft"
	ArtifactStatusReviewing = "reviewing"
	ArtifactStatusSignedOff = "signed_off"
	ArtifactStatusArchived  = "archived"
)

// Artifact is the typed, versioned, project-scoped output of a piece of
// work. The body lives in [ArtifactVersion] rows (one per version);
// [Artifact.LatestVersion] points at the current head.
//
// `Type` is intentionally free-form — the user defines what types matter
// (note, prd, spec, task, …). v0 places no constraints beyond "non-empty".
type Artifact struct {
	ID             uuid.UUID
	ProjectID      uuid.UUID
	Type           string
	Status         string
	Parents        []uuid.UUID
	LatestVersion  int
	IdempotencyKey string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// ArtifactVersion is one immutable revision of an [Artifact]. ContentJSON is
// the structured body; ContentText is an optional plain-text projection
// (typically used for full-text search / recall in later phases).
type ArtifactVersion struct {
	ArtifactID  uuid.UUID
	Version     int
	Content     map[string]any
	ContentText string
	CreatedAt   time.Time
}
