package pgstore

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/manuelibar/workbench/internal/domain"
)

const (
	pgUniqueViolation = "23505"

	// DefaultNoteListLimit is the default page size for [Store.ListNotes] /
	// [Store.SearchNotes] when the caller does not specify one.
	DefaultNoteListLimit = 50
	// MaxNoteListLimit is the cap applied to caller-supplied limits.
	MaxNoteListLimit = 200
)

// noteSelectColumns is the canonical column projection used by every read
// path so scan order stays in sync.
const noteSelectColumns = `id, user_id, body_md, tags,
    namespace_id, project_id, promoted_to, idempotency_key,
    created_at, updated_at`

// AddNote inserts n. If n.ID is the zero UUID, a fresh UUIDv7 is generated.
//
// When n.IdempotencyKey is non-empty and a row already exists for
// (user_id, idempotency_key), the previously inserted row is returned
// (no new row is created). Otherwise the freshly inserted row is returned.
func (s *Store) AddNote(ctx context.Context, n domain.Note) (domain.Note, error) {
	if n.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return domain.Note{}, fmt.Errorf("pgstore: generate note id: %w", err)
		}
		n.ID = id
	}
	// The notes.tags column is NOT NULL DEFAULT '{}'. The DEFAULT only kicks
	// in when the column is omitted from the INSERT, not when it's passed
	// explicitly as NULL — so normalise a nil Go slice to an empty array.
	if n.Tags == nil {
		n.Tags = []string{}
	}
	var idempCol any
	if n.IdempotencyKey != "" {
		idempCol = n.IdempotencyKey
	}

	row := s.pool.QueryRow(ctx, `
		INSERT INTO notes (id, user_id, body_md, tags, namespace_id, project_id, idempotency_key)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING `+noteSelectColumns,
		n.ID, n.UserID, n.BodyMD, n.Tags, n.NamespaceID, n.ProjectID, idempCol)

	out, err := scanNote(row)
	if err == nil {
		return out, nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation && pgErr.ConstraintName == "notes_idempotency" {
		return s.GetNoteByIdempotency(ctx, n.UserID, n.IdempotencyKey)
	}
	return domain.Note{}, fmt.Errorf("pgstore: insert note: %w", err)
}

// GetNote returns a single note by id, scoped to userID for safety.
func (s *Store) GetNote(ctx context.Context, userID, id uuid.UUID) (domain.Note, error) {
	row := s.pool.QueryRow(ctx, `SELECT `+noteSelectColumns+` FROM notes WHERE id = $1 AND user_id = $2`, id, userID)
	out, err := scanNote(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Note{}, ErrNotFound
	}
	if err != nil {
		return domain.Note{}, fmt.Errorf("pgstore: get note: %w", err)
	}
	return out, nil
}

// GetNoteByIdempotency returns the note previously written with the supplied
// idempotency key, or [ErrNotFound] if none exists.
func (s *Store) GetNoteByIdempotency(ctx context.Context, userID uuid.UUID, key string) (domain.Note, error) {
	row := s.pool.QueryRow(ctx, `SELECT `+noteSelectColumns+` FROM notes WHERE user_id = $1 AND idempotency_key = $2`, userID, key)
	out, err := scanNote(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Note{}, ErrNotFound
	}
	if err != nil {
		return domain.Note{}, fmt.Errorf("pgstore: get note by idempotency: %w", err)
	}
	return out, nil
}

// ListNotesFilter narrows the list returned by [Store.ListNotes].
type ListNotesFilter struct {
	Tag         string     // optional, filter by exact tag membership
	NamespaceID *uuid.UUID // optional, filter by capture-time namespace
	ProjectID   *uuid.UUID // optional, filter by capture-time project
	Since       *time.Time // optional, only notes created at-or-after Since
	Limit       int        // 0 → DefaultNoteListLimit; capped at MaxNoteListLimit
}

// ListNotes returns the user's notes most-recent-first.
func (s *Store) ListNotes(ctx context.Context, userID uuid.UUID, f ListNotesFilter) ([]domain.Note, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = DefaultNoteListLimit
	}
	if limit > MaxNoteListLimit {
		limit = MaxNoteListLimit
	}

	q := `SELECT ` + noteSelectColumns + ` FROM notes WHERE user_id = $1`
	args := []any{userID}
	if f.Tag != "" {
		q += fmt.Sprintf(" AND $%d = ANY (tags)", len(args)+1)
		args = append(args, f.Tag)
	}
	if f.NamespaceID != nil {
		q += fmt.Sprintf(" AND namespace_id = $%d", len(args)+1)
		args = append(args, *f.NamespaceID)
	}
	if f.ProjectID != nil {
		q += fmt.Sprintf(" AND project_id = $%d", len(args)+1)
		args = append(args, *f.ProjectID)
	}
	if f.Since != nil {
		q += fmt.Sprintf(" AND created_at >= $%d", len(args)+1)
		args = append(args, *f.Since)
	}
	q += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", len(args)+1)
	args = append(args, limit)

	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("pgstore: list notes: %w", err)
	}
	defer rows.Close()
	return scanNotes(rows)
}

// SearchNotes performs a substring (case-insensitive) match on body_md.
// v0 uses ILIKE; semantic search arrives in a later phase.
func (s *Store) SearchNotes(ctx context.Context, userID uuid.UUID, query string, limit int) ([]domain.Note, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("pgstore: search notes: empty query")
	}
	if limit <= 0 {
		limit = DefaultNoteListLimit
	}
	if limit > MaxNoteListLimit {
		limit = MaxNoteListLimit
	}
	rows, err := s.pool.Query(ctx, `
		SELECT `+noteSelectColumns+`
		FROM notes
		WHERE user_id = $1 AND body_md ILIKE '%' || $2 || '%'
		ORDER BY created_at DESC
		LIMIT $3
	`, userID, query, limit)
	if err != nil {
		return nil, fmt.Errorf("pgstore: search notes: %w", err)
	}
	defer rows.Close()
	return scanNotes(rows)
}

// UpdateNoteFields holds the optional patches applied by [Store.UpdateNote].
// A nil pointer means "leave this field as is"; a non-nil pointer to an empty
// value means "set to empty".
type UpdateNoteFields struct {
	BodyMD *string
	Tags   *[]string
}

// UpdateNote patches the named note. Returns the updated row.
func (s *Store) UpdateNote(ctx context.Context, userID, id uuid.UUID, f UpdateNoteFields) (domain.Note, error) {
	sets := []string{"updated_at = now()"}
	args := []any{id, userID}
	if f.BodyMD != nil {
		sets = append(sets, fmt.Sprintf("body_md = $%d", len(args)+1))
		args = append(args, *f.BodyMD)
	}
	if f.Tags != nil {
		sets = append(sets, fmt.Sprintf("tags = $%d", len(args)+1))
		args = append(args, *f.Tags)
	}
	if len(sets) == 1 {
		// no-op patch — just bump updated_at by re-reading
		return s.GetNote(ctx, userID, id)
	}
	q := `UPDATE notes SET ` + strings.Join(sets, ", ") + ` WHERE id = $1 AND user_id = $2 RETURNING ` + noteSelectColumns
	row := s.pool.QueryRow(ctx, q, args...)
	out, err := scanNote(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Note{}, ErrNotFound
	}
	if err != nil {
		return domain.Note{}, fmt.Errorf("pgstore: update note: %w", err)
	}
	return out, nil
}

// DeleteNote removes the named note. Returns [ErrNotFound] if no such row.
func (s *Store) DeleteNote(ctx context.Context, userID, id uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM notes WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("pgstore: delete note: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// scannable is the minimal interface implemented by both [pgx.Row] and the
// rows returned by [pgx.Rows.Next] (via the .Scan method on pgx.Rows).
type scannable interface {
	Scan(dest ...any) error
}

func scanNote(r scannable) (domain.Note, error) {
	var (
		n   domain.Note
		key *string
	)
	if err := r.Scan(
		&n.ID, &n.UserID, &n.BodyMD, &n.Tags,
		&n.NamespaceID, &n.ProjectID, &n.PromotedTo, &key,
		&n.CreatedAt, &n.UpdatedAt,
	); err != nil {
		return domain.Note{}, err
	}
	if key != nil {
		n.IdempotencyKey = *key
	}
	return n, nil
}

func scanNotes(rows pgx.Rows) ([]domain.Note, error) {
	var out []domain.Note
	for rows.Next() {
		n, err := scanNote(rows)
		if err != nil {
			return nil, fmt.Errorf("pgstore: scan note: %w", err)
		}
		out = append(out, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("pgstore: iterate notes: %w", err)
	}
	return out, nil
}
