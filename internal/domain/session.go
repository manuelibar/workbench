package domain

import (
	"time"

	"github.com/google/uuid"
)

// Selection captures the workbench's currently-selected hierarchy slice.
// All fields are nil/empty when no selection is active; entries are filled in
// as the agent narrows scope: namespace → project → blueprint → mode.
//
// Selection is the single source of truth for the MCP tool surface — the
// server registers tools whose visibility depends on which fields are set.
type Selection struct {
	NamespaceID *uuid.UUID `json:"namespace_id,omitempty"`
	ProjectID   *uuid.UUID `json:"project_id,omitempty"`
	ArtifactID  *uuid.UUID `json:"artifact_id,omitempty"`
	BlueprintID *uuid.UUID `json:"blueprint_id,omitempty"`
	ModeName    string     `json:"mode_name,omitempty"`
	Focus       string     `json:"focus,omitempty"`
}

// IsEmpty reports whether no selection is active.
func (s Selection) IsEmpty() bool {
	return s.NamespaceID == nil && s.ProjectID == nil && s.ArtifactID == nil && s.BlueprintID == nil && s.ModeName == "" && s.Focus == ""
}

// WorkSession is the daily-scoped session that owns selection state and is
// shared across every connected MCP-protocol session. v0 enforces exactly one
// open WorkSession per user (partial unique index on the work_sessions table).
type WorkSession struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Name      string
	StartedAt time.Time
	LastSeen  time.Time
	ClosedAt  *time.Time
	Selection Selection
}

// IsOpen reports whether the WorkSession is still accepting work — that is,
// its closed_at column is NULL.
func (w WorkSession) IsOpen() bool { return w.ClosedAt == nil }
