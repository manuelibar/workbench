package domain

import (
	"time"

	"github.com/google/uuid"
)

// Note is the workbench's universal Zettelkasten capture primitive: a
// scope-agnostic markdown blob with optional tags, recorded with whatever
// selection was active at capture time. Promotion (note → typed artifact)
// is a separate workflow that points [Note.PromotedTo] at the resulting
// artifact id; v0 only models the field.
type Note struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	BodyMD         string
	Tags           []string
	NamespaceID    *uuid.UUID // capture-time namespace selection, if any
	ProjectID      *uuid.UUID // capture-time project selection, if any
	PromotedTo     *uuid.UUID // artifact id once the note has been promoted
	IdempotencyKey string     // empty == not deduplicated
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
