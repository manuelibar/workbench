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

const projectSelectColumns = `id, namespace_id, name, description,
    settings_jsonb, idempotency_key, created_at, updated_at`

// CreateProject inserts p. If p.ID is the zero UUID a fresh UUIDv7 is
// assigned. Idempotency-key replays return the original row.
func (s *Store) CreateProject(ctx context.Context, p domain.Project) (domain.Project, error) {
	if p.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return domain.Project{}, fmt.Errorf("pgstore: generate project id: %w", err)
		}
		p.ID = id
	}
	settings, err := marshalSettings(p.Settings)
	if err != nil {
		return domain.Project{}, fmt.Errorf("pgstore: marshal project settings: %w", err)
	}
	var idempCol any
	if p.IdempotencyKey != "" {
		idempCol = p.IdempotencyKey
	}

	row := s.pool.QueryRow(ctx, `
		INSERT INTO projects (id, namespace_id, name, description, settings_jsonb, idempotency_key)
		VALUES ($1, $2, $3, $4, $5::jsonb, $6)
		RETURNING `+projectSelectColumns,
		p.ID, p.NamespaceID, p.Name, p.Description, settings, idempCol)
	out, err := scanProject(row)
	if err == nil {
		return out, nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		if pgErr.ConstraintName == "projects_idempotency" && p.IdempotencyKey != "" {
			return s.GetProjectByIdempotency(ctx, p.IdempotencyKey)
		}
		if pgErr.ConstraintName == "projects_unique_name_per_namespace" {
			return domain.Project{}, fmt.Errorf("%w: project name %q already exists in this namespace", ErrConflict, p.Name)
		}
	}
	return domain.Project{}, fmt.Errorf("pgstore: insert project: %w", err)
}

// GetProject returns the named project.
func (s *Store) GetProject(ctx context.Context, id uuid.UUID) (domain.Project, error) {
	row := s.pool.QueryRow(ctx, `SELECT `+projectSelectColumns+` FROM projects WHERE id = $1`, id)
	out, err := scanProject(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Project{}, ErrNotFound
	}
	if err != nil {
		return domain.Project{}, fmt.Errorf("pgstore: get project: %w", err)
	}
	return out, nil
}

// GetProjectByIdempotency returns the project previously written with the
// supplied idempotency key, or [ErrNotFound] if none exists.
func (s *Store) GetProjectByIdempotency(ctx context.Context, key string) (domain.Project, error) {
	row := s.pool.QueryRow(ctx, `SELECT `+projectSelectColumns+` FROM projects WHERE idempotency_key = $1`, key)
	out, err := scanProject(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Project{}, ErrNotFound
	}
	if err != nil {
		return domain.Project{}, fmt.Errorf("pgstore: get project by idempotency: %w", err)
	}
	return out, nil
}

// ListProjects returns the projects in namespaceID (nil = standalone projects).
func (s *Store) ListProjects(ctx context.Context, namespaceID *uuid.UUID) ([]domain.Project, error) {
	q := `SELECT ` + projectSelectColumns + ` FROM projects`
	args := []any{}
	if namespaceID == nil {
		q += ` WHERE namespace_id IS NULL`
	} else {
		q += ` WHERE namespace_id = $1`
		args = append(args, *namespaceID)
	}
	q += ` ORDER BY name ASC`
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("pgstore: list projects: %w", err)
	}
	defer rows.Close()
	var out []domain.Project
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, fmt.Errorf("pgstore: scan project: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("pgstore: iterate projects: %w", err)
	}
	return out, nil
}

// UpdateProjectFields holds the optional patches applied by [Store.UpdateProject].
type UpdateProjectFields struct {
	Name        *string
	Description *string
	Settings    *map[string]any
}

// UpdateProject patches the named project. Returns the updated row.
func (s *Store) UpdateProject(ctx context.Context, id uuid.UUID, f UpdateProjectFields) (domain.Project, error) {
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
			return domain.Project{}, fmt.Errorf("pgstore: marshal project settings: %w", err)
		}
		sets = append(sets, fmt.Sprintf("settings_jsonb = $%d::jsonb", len(args)+1))
		args = append(args, raw)
	}
	if len(sets) == 1 {
		return s.GetProject(ctx, id)
	}
	q := `UPDATE projects SET ` + strings.Join(sets, ", ") + ` WHERE id = $1 RETURNING ` + projectSelectColumns
	row := s.pool.QueryRow(ctx, q, args...)
	out, err := scanProject(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Project{}, ErrNotFound
	}
	if err != nil {
		return domain.Project{}, fmt.Errorf("pgstore: update project: %w", err)
	}
	return out, nil
}

// DeleteProject removes the project.
func (s *Store) DeleteProject(ctx context.Context, id uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM projects WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("pgstore: delete project: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func scanProject(r scannable) (domain.Project, error) {
	var (
		p        domain.Project
		settings []byte
		key      *string
	)
	if err := r.Scan(
		&p.ID, &p.NamespaceID, &p.Name, &p.Description,
		&settings, &key, &p.CreatedAt, &p.UpdatedAt,
	); err != nil {
		return domain.Project{}, err
	}
	if len(settings) > 0 {
		_ = json.Unmarshal(settings, &p.Settings)
	}
	if key != nil {
		p.IdempotencyKey = *key
	}
	return p, nil
}
