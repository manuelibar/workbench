package domain

import (
	"time"

	"github.com/google/uuid"
)

// Event is one entry in the append-only events log: an observation about
// something the workbench did or saw, attributed to a WorkSession (and
// optionally the MCP-protocol session that triggered it).
//
// Events are the workbench's episodic memory and the source of
// `recent_events` in the refresh tool's result.
type Event struct {
	ID             uuid.UUID
	WorkSessionID  uuid.UUID
	MCPSessionID   string // optional; empty for system-originated events
	OccurredAt     time.Time
	Type           string         // e.g. "tool.call", "tool.failed"
	SubjectKind    string         // e.g. "tool", "note", "artifact"
	SubjectID      string         // free-form identifier of the subject
	Payload        map[string]any // arbitrary structured detail
	RequestID      uuid.UUID
	CorrelationID  uuid.UUID
	CausationID    uuid.UUID
	IdempotencyKey string
}
