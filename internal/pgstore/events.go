package pgstore

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/manuelibar/workbench/internal/domain"
)

const (
	// DefaultRecentEventsLimit is the default page size for [Store.RecentEvents].
	DefaultRecentEventsLimit = 20
	// MaxRecentEventsLimit is the cap applied to caller-supplied limits.
	MaxRecentEventsLimit = 200
)

// RecordEvent appends e to the events log. If e.ID is the zero UUID, a fresh
// UUIDv7 is assigned. The supplied [domain.Event.OccurredAt] is honoured if
// non-zero; otherwise the database default (now()) is used.
func (s *Store) RecordEvent(ctx context.Context, e domain.Event) (domain.Event, error) {
	if e.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return domain.Event{}, fmt.Errorf("pgstore: generate event id: %w", err)
		}
		e.ID = id
	}
	if e.WorkSessionID == uuid.Nil {
		return domain.Event{}, fmt.Errorf("pgstore: record event: WorkSessionID required")
	}
	if e.Type == "" {
		return domain.Event{}, fmt.Errorf("pgstore: record event: Type required")
	}

	payload, err := json.Marshal(e.Payload)
	if err != nil {
		return domain.Event{}, fmt.Errorf("pgstore: marshal event payload: %w", err)
	}
	if len(payload) == 0 || string(payload) == "null" {
		payload = []byte(`{}`)
	}

	var occurredAtArg any
	if !e.OccurredAt.IsZero() {
		occurredAtArg = e.OccurredAt
	}
	var mcpSessArg, subjectKindArg, subjectIDArg, idempArg any
	if e.MCPSessionID != "" {
		mcpSessArg = e.MCPSessionID
	}
	if e.SubjectKind != "" {
		subjectKindArg = e.SubjectKind
	}
	if e.SubjectID != "" {
		subjectIDArg = e.SubjectID
	}
	if e.IdempotencyKey != "" {
		idempArg = e.IdempotencyKey
	}
	var requestArg, correlationArg, causationArg any
	if e.RequestID != uuid.Nil {
		requestArg = e.RequestID
	}
	if e.CorrelationID != uuid.Nil {
		correlationArg = e.CorrelationID
	}
	if e.CausationID != uuid.Nil {
		causationArg = e.CausationID
	}

	row := s.pool.QueryRow(ctx, `
		INSERT INTO events (
			id, work_session_id, mcp_session_id,
			occurred_at, type, subject_kind, subject_id, payload_jsonb,
			request_id, correlation_id, causation_id, idempotency_key
		)
		VALUES (
			$1, $2, $3,
			COALESCE($4, now()), $5, $6, $7, $8::jsonb,
			$9, $10, $11, $12
		)
		RETURNING id, work_session_id, mcp_session_id,
			occurred_at, type, subject_kind, subject_id, payload_jsonb,
			request_id, correlation_id, causation_id, idempotency_key`,
		e.ID, e.WorkSessionID, mcpSessArg,
		occurredAtArg, e.Type, subjectKindArg, subjectIDArg, payload,
		requestArg, correlationArg, causationArg, idempArg,
	)
	out, err := scanEvent(row)
	if err != nil {
		return domain.Event{}, fmt.Errorf("pgstore: insert event: %w", err)
	}
	return out, nil
}

// RecentEvents returns up to limit events for workSessionID, most-recent-first.
func (s *Store) RecentEvents(ctx context.Context, workSessionID uuid.UUID, limit int) ([]domain.Event, error) {
	if limit <= 0 {
		limit = DefaultRecentEventsLimit
	}
	if limit > MaxRecentEventsLimit {
		limit = MaxRecentEventsLimit
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, work_session_id, mcp_session_id,
		       occurred_at, type, subject_kind, subject_id, payload_jsonb,
		       request_id, correlation_id, causation_id, idempotency_key
		FROM events
		WHERE work_session_id = $1
		ORDER BY occurred_at DESC
		LIMIT $2`, workSessionID, limit)
	if err != nil {
		return nil, fmt.Errorf("pgstore: recent events: %w", err)
	}
	defer rows.Close()
	var out []domain.Event
	for rows.Next() {
		ev, err := scanEvent(rows)
		if err != nil {
			return nil, fmt.Errorf("pgstore: scan event: %w", err)
		}
		out = append(out, ev)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("pgstore: iterate events: %w", err)
	}
	return out, nil
}

func scanEvent(r scannable) (domain.Event, error) {
	var (
		ev         domain.Event
		mcpSess    *string
		subjKind   *string
		subjID     *string
		idemp      *string
		reqID      *uuid.UUID
		corrID     *uuid.UUID
		causID     *uuid.UUID
		payloadRaw []byte
	)
	if err := r.Scan(
		&ev.ID, &ev.WorkSessionID, &mcpSess,
		&ev.OccurredAt, &ev.Type, &subjKind, &subjID, &payloadRaw,
		&reqID, &corrID, &causID, &idemp,
	); err != nil {
		return domain.Event{}, err
	}
	if mcpSess != nil {
		ev.MCPSessionID = *mcpSess
	}
	if subjKind != nil {
		ev.SubjectKind = *subjKind
	}
	if subjID != nil {
		ev.SubjectID = *subjID
	}
	if idemp != nil {
		ev.IdempotencyKey = *idemp
	}
	if reqID != nil {
		ev.RequestID = *reqID
	}
	if corrID != nil {
		ev.CorrelationID = *corrID
	}
	if causID != nil {
		ev.CausationID = *causID
	}
	if len(payloadRaw) > 0 {
		_ = json.Unmarshal(payloadRaw, &ev.Payload)
	}
	return ev, nil
}
