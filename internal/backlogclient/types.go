// Package backlogclient is the workbench-side HTTP client for the
// standalone backlog-service. Its wire DTOs mirror
// backlog-service/internal/httpapi byte-for-byte; the two stay in lockstep
// by convention.
package backlogclient

import "time"

// Issue is the wire shape of an issue. UUIDs travel as strings.
type Issue struct {
	ID             string    `json:"id"`
	ProjectID      string    `json:"project_id"`
	Type           string    `json:"type"`
	Status         string    `json:"status"`
	Priority       string    `json:"priority"`
	Title          string    `json:"title"`
	Body           string    `json:"body"`
	Labels         []string  `json:"labels"`
	ParentID       string    `json:"parent_id,omitempty"`
	Assignee       string    `json:"assignee,omitempty"`
	SourceRefs     []string  `json:"source_refs"`
	Version        int       `json:"version"`
	RequestID      string    `json:"request_id,omitempty"`
	CorrelationID  string    `json:"correlation_id,omitempty"`
	CausationID    string    `json:"causation_id,omitempty"`
	IdempotencyKey string    `json:"idempotency_key,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// IssueCreateRequest is the body of `POST /v1/issues`.
type IssueCreateRequest struct {
	ProjectID  string   `json:"project_id"`
	Title      string   `json:"title"`
	Body       string   `json:"body,omitempty"`
	Type       string   `json:"type,omitempty"`
	Status     string   `json:"status,omitempty"`
	Priority   string   `json:"priority,omitempty"`
	Labels     []string `json:"labels,omitempty"`
	ParentID   string   `json:"parent_id,omitempty"`
	Assignee   string   `json:"assignee,omitempty"`
	SourceRefs []string `json:"source_refs,omitempty"`
}

// IssueUpdateRequest is the body of `PATCH /v1/issues/{id}`. Pointer fields
// distinguish "leave unchanged" from "set to this value".
type IssueUpdateRequest struct {
	ExpectedVersion int       `json:"expected_version"`
	Title           *string   `json:"title,omitempty"`
	Body            *string   `json:"body,omitempty"`
	Status          *string   `json:"status,omitempty"`
	Type            *string   `json:"type,omitempty"`
	Priority        *string   `json:"priority,omitempty"`
	Labels          *[]string `json:"labels,omitempty"`
	ParentID        *string   `json:"parent_id,omitempty"` // "" = clear
	Assignee        *string   `json:"assignee,omitempty"`  // "" = clear
	SourceRefs      *[]string `json:"source_refs,omitempty"`
}

// IssueListResponse is the body of `GET /v1/issues`.
type IssueListResponse struct {
	Items      []Issue `json:"items"`
	NextOffset int     `json:"next_offset,omitempty"`
}

// TakeNextRequest is the body of `POST /v1/issues:take_next`.
type TakeNextRequest struct {
	ProjectID string `json:"project_id,omitempty"`
	Assignee  string `json:"assignee,omitempty"`
}

// Event mirrors backlog-service's EventDTO.
type Event struct {
	ID             string         `json:"id"`
	OccurredAt     time.Time      `json:"occurred_at"`
	Type           string         `json:"type"`
	SubjectKind    string         `json:"subject_kind"`
	SubjectID      string         `json:"subject_id"`
	Actor          string         `json:"actor,omitempty"`
	ProjectID      string         `json:"project_id,omitempty"`
	Payload        map[string]any `json:"payload"`
	RequestID      string         `json:"request_id,omitempty"`
	CorrelationID  string         `json:"correlation_id,omitempty"`
	CausationID    string         `json:"causation_id,omitempty"`
	IdempotencyKey string         `json:"idempotency_key,omitempty"`
}

// EventListResponse is the body of `GET /v1/events`.
type EventListResponse struct {
	Items      []Event `json:"items"`
	NextOffset int     `json:"next_offset,omitempty"`
}
