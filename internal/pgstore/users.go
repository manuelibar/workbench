package pgstore

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/manuelibar/workbench/internal/domain"
)

// DefaultUserName is the display name used by [Store.EnsureSingletonUser]
// when no name is supplied and no user row exists yet.
const DefaultUserName = "default"

// EnsureSingletonUser returns the single user row for this workbench
// instance, creating it with displayName (or [DefaultUserName] if empty) if
// no row exists yet. v0 enforces no real "single user" invariant at the
// schema level — it just looks up the first row — but no path in v0 ever
// inserts more than one.
func (s *Store) EnsureSingletonUser(ctx context.Context, displayName string) (domain.User, error) {
	if displayName == "" {
		displayName = DefaultUserName
	}
	var u domain.User
	err := s.pool.QueryRow(ctx, `SELECT id, display_name FROM users LIMIT 1`).
		Scan(&u.ID, &u.DisplayName)
	switch {
	case err == nil:
		return u, nil
	case errors.Is(err, pgx.ErrNoRows):
		// fall through to insert
	default:
		return domain.User{}, fmt.Errorf("pgstore: read user: %w", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return domain.User{}, fmt.Errorf("pgstore: generate user id: %w", err)
	}
	u = domain.User{ID: id, DisplayName: displayName}
	if _, err := s.pool.Exec(ctx,
		`INSERT INTO users (id, display_name) VALUES ($1, $2)`,
		u.ID, u.DisplayName,
	); err != nil {
		return domain.User{}, fmt.Errorf("pgstore: insert user: %w", err)
	}
	return u, nil
}
