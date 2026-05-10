package pgstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/manuelibar/workbench/internal/domain"
)

const namespaceSelectColumns = `id, parent_id, name, description,
    settings_jsonb, idempotency_key, created_at, updated_at`

// CreateNamespace inserts n. If n.ID is the zero UUID a fresh UUIDv7 is
// assigned. If n.IdempotencyKey is non-empty and a row already exists with
// that key, the existing row is returned (no new row is created).
func (s *Store) CreateNamespace(ctx context.Context, n domain.Namespace) (domain.Namespace, error) {
	if n.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return domain.Namespace{}, fmt.Errorf("pgstore: generate namespace id: %w", err)
		}
		n.ID = id
	}
	settings, err := marshalSettings(n.Settings)
	if err != nil {
		return domain.Namespace{}, fmt.Errorf("pgstore: marshal namespace settings: %w", err)
	}
	var idempCol any
	if n.IdempotencyKey != "" {
		idempCol = n.IdempotencyKey
	}

	row := s.pool.QueryRow(ctx, `
		INSERT INTO namespaces (id, parent_id, name, description, settings_jsonb, idempotency_key)
		VALUES ($1, $2, $3, $4, $5::jsonb, $6)
		RETURNING `+namespaceSelectColumns,
		n.ID, n.ParentID, n.Name, n.Description, settings, idempCol)
	out, err := scanNamespace(row)
	if err == nil {
		return out, nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		if pgErr.ConstraintName == "namespaces_idempotency" && n.IdempotencyKey != "" {
			return s.GetNamespaceByIdempotency(ctx, n.IdempotencyKey)
		}
		if pgErr.ConstraintName == "namespaces_unique_name_per_parent" {
			return domain.Namespace{}, fmt.Errorf("%w: namespace name %q already exists under this parent", ErrConflict, n.Name)
		}
	}
	return domain.Namespace{}, fmt.Errorf("pgstore: insert namespace: %w", err)
}

// GetNamespace returns the named namespace.
func (s *Store) GetNamespace(ctx context.Context, id uuid.UUID) (domain.Namespace, error) {
	row := s.pool.QueryRow(ctx, `SELECT `+namespaceSelectColumns+` FROM namespaces WHERE id = $1`, id)
	out, err := scanNamespace(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Namespace{}, ErrNotFound
	}
	if err != nil {
		return domain.Namespace{}, fmt.Errorf("pgstore: get namespace: %w", err)
	}
	return out, nil
}

// GetNamespaceByIdempotency returns the namespace previously written with
// the supplied idempotency key, or [ErrNotFound] if none exists.
func (s *Store) GetNamespaceByIdempotency(ctx context.Context, key string) (domain.Namespace, error) {
	row := s.pool.QueryRow(ctx, `SELECT `+namespaceSelectColumns+` FROM namespaces WHERE idempotency_key = $1`, key)
	out, err := scanNamespace(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Namespace{}, ErrNotFound
	}
	if err != nil {
		return domain.Namespace{}, fmt.Errorf("pgstore: get namespace by idempotency: %w", err)
	}
	return out, nil
}

// ListNamespaces returns immediate children of parentID (nil = roots).
func (s *Store) ListNamespaces(ctx context.Context, parentID *uuid.UUID) ([]domain.Namespace, error) {
	q := `SELECT ` + namespaceSelectColumns + ` FROM namespaces`
	args := []any{}
	if parentID == nil {
		q += ` WHERE parent_id IS NULL`
	} else {
		q += ` WHERE parent_id = $1`
		args = append(args, *parentID)
	}
	q += ` ORDER BY name ASC`
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("pgstore: list namespaces: %w", err)
	}
	defer rows.Close()
	var out []domain.Namespace
	for rows.Next() {
		n, err := scanNamespace(rows)
		if err != nil {
			return nil, fmt.Errorf("pgstore: scan namespace: %w", err)
		}
		out = append(out, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("pgstore: iterate namespaces: %w", err)
	}
	return out, nil
}

// UpdateNamespaceFields holds the optional patches applied by [Store.UpdateNamespace].
type UpdateNamespaceFields struct {
	Name        *string
	Description *string
	Settings    *map[string]any
}

// UpdateNamespace patches the named namespace. Returns the updated row.
func (s *Store) UpdateNamespace(ctx context.Context, id uuid.UUID, f UpdateNamespaceFields) (domain.Namespace, error) {
	sets := []string{"updated_at = now()"}
	args := []any{id}
	if f.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", len(args)+1))
		args = append(args, *f.Name)
	}
	if f.Description != nil {
		sets = append(sets, fmt.Sprintf("description = $%d", len(args)+1))
		args = append(args, *f.Description)
	}
	if f.Settings != nil {
		raw, err := marshalSettings(*f.Settings)
		if err != nil {
			return domain.Namespace{}, fmt.Errorf("pgstore: marshal namespace settings: %w", err)
		}
		sets = append(sets, fmt.Sprintf("settings_jsonb = $%d::jsonb", len(args)+1))
		args = append(args, raw)
	}
	if len(sets) == 1 {
		return s.GetNamespace(ctx, id)
	}
	q := `UPDATE namespaces SET ` + strings.Join(sets, ", ") + ` WHERE id = $1 RETURNING ` + namespaceSelectColumns
	row := s.pool.QueryRow(ctx, q, args...)
	out, err := scanNamespace(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Namespace{}, ErrNotFound
	}
	if err != nil {
		return domain.Namespace{}, fmt.Errorf("pgstore: update namespace: %w", err)
	}
	return out, nil
}

// DeleteNamespace removes the namespace. Children are cascaded.
func (s *Store) DeleteNamespace(ctx context.Context, id uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM namespaces WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("pgstore: delete namespace: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func scanNamespace(r scannable) (domain.Namespace, error) {
	var (
		n        domain.Namespace
		settings []byte
		key      *string
	)
	if err := r.Scan(
		&n.ID, &n.ParentID, &n.Name, &n.Description,
		&settings, &key, &n.CreatedAt, &n.UpdatedAt,
	); err != nil {
		return domain.Namespace{}, err
	}
	if len(settings) > 0 {
		_ = json.Unmarshal(settings, &n.Settings)
	}
	if key != nil {
		n.IdempotencyKey = *key
	}
	return n, nil
}

// marshalSettings is the shared helper used by namespace and project rows
// to encode JSONB columns.
func marshalSettings(m map[string]any) ([]byte, error) {
	if len(m) == 0 {
		return []byte(`{}`), nil
	}
	return json.Marshal(m)
}
