package pgstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/manuelibar/workbench/internal/domain"
)

// EnsureOpenWorkSession returns the user's open WorkSession, creating one if
// none exists. The schema's partial unique index `WHERE closed_at IS NULL`
// guarantees at most one open row per user, so concurrent calls converge on
// the same session.
//
// If name is empty, the new session is named after the current UTC date
// (YYYY-MM-DD). The session's selection_jsonb defaults to an empty object
// (no selection).
func (s *Store) EnsureOpenWorkSession(ctx context.Context, userID uuid.UUID, name string) (domain.WorkSession, error) {
	ws, err := s.readOpenWorkSession(ctx, userID)
	switch {
	case err == nil:
		return ws, nil
	case errors.Is(err, ErrNotFound):
		// fall through to create
	default:
		return domain.WorkSession{}, err
	}

	if name == "" {
		name = time.Now().UTC().Format("2006-01-02")
	}
	id, err := uuid.NewV7()
	if err != nil {
		return domain.WorkSession{}, fmt.Errorf("pgstore: generate work_session id: %w", err)
	}
	if _, err := s.pool.Exec(ctx,
		`INSERT INTO work_sessions (id, user_id, name) VALUES ($1, $2, $3)`,
		id, userID, name,
	); err != nil {
		return domain.WorkSession{}, fmt.Errorf("pgstore: insert work_session: %w", err)
	}
	// Re-read to capture DB-supplied timestamps and default selection_jsonb.
	return s.readOpenWorkSession(ctx, userID)
}

// UpdateSelection persists the supplied selection on the open WorkSession
// for userID and bumps last_seen. Returns [ErrNotFound] if no open session
// exists.
func (s *Store) UpdateSelection(ctx context.Context, userID uuid.UUID, sel domain.Selection) error {
	raw, err := json.Marshal(sel)
	if err != nil {
		return fmt.Errorf("pgstore: encode selection: %w", err)
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE work_sessions
		SET selection_jsonb = $1::jsonb,
		    last_seen       = now(),
		    updated_at      = now()
		WHERE user_id = $2 AND closed_at IS NULL
	`, raw, userID)
	if err != nil {
		return fmt.Errorf("pgstore: update selection: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// CloseWorkSession marks the open work session for userID as closed. Returns
// [ErrNotFound] if there is no open session.
func (s *Store) CloseWorkSession(ctx context.Context, userID uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE work_sessions
		SET closed_at  = now(),
		    updated_at = now()
		WHERE user_id = $1 AND closed_at IS NULL
	`, userID)
	if err != nil {
		return fmt.Errorf("pgstore: close work_session: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) readOpenWorkSession(ctx context.Context, userID uuid.UUID) (domain.WorkSession, error) {
	var (
		ws     domain.WorkSession
		rawSel []byte
	)
	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id, name, started_at, last_seen, closed_at, selection_jsonb
		FROM work_sessions
		WHERE user_id = $1 AND closed_at IS NULL
		LIMIT 1
	`, userID).Scan(&ws.ID, &ws.UserID, &ws.Name, &ws.StartedAt, &ws.LastSeen, &ws.ClosedAt, &rawSel)
	switch {
	case err == nil:
		// proceed
	case errors.Is(err, pgx.ErrNoRows):
		return domain.WorkSession{}, ErrNotFound
	default:
		return domain.WorkSession{}, fmt.Errorf("pgstore: read work_session: %w", err)
	}
	if len(rawSel) > 0 {
		if err := json.Unmarshal(rawSel, &ws.Selection); err != nil {
			return domain.WorkSession{}, fmt.Errorf("pgstore: decode selection: %w", err)
		}
	}
	return ws, nil
}
