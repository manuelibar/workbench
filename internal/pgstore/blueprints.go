package pgstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/manuelibar/workbench/internal/domain"
)

const blueprintSelectColumns = `id, project_id, name, version, definition_jsonb,
    idempotency_key, created_at, updated_at`

// CreateBlueprint inserts a brand-new blueprint at version 1 for a fresh
// (project, name) pair. To update an existing blueprint's content, use
// [Store.AppendBlueprintVersion].
func (s *Store) CreateBlueprint(ctx context.Context, b domain.Blueprint) (domain.Blueprint, error) {
	if b.ProjectID == uuid.Nil {
		return domain.Blueprint{}, fmt.Errorf("pgstore: blueprint: project_id required")
	}
	if b.Name == "" {
		return domain.Blueprint{}, fmt.Errorf("pgstore: blueprint: name required")
	}
	if b.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return domain.Blueprint{}, err
		}
		b.ID = id
	}
	if b.Version == 0 {
		b.Version = 1
	}
	def := []byte(`{}`)
	if len(b.Definition) > 0 {
		raw, err := json.Marshal(b.Definition)
		if err != nil {
			return domain.Blueprint{}, fmt.Errorf("pgstore: marshal blueprint definition: %w", err)
		}
		def = raw
	}
	var idempCol any
	if b.IdempotencyKey != "" {
		idempCol = b.IdempotencyKey
	}

	row := s.pool.QueryRow(ctx, `
		INSERT INTO blueprints (id, project_id, name, version, definition_jsonb, idempotency_key)
		VALUES ($1, $2, $3, $4, $5::jsonb, $6)
		RETURNING `+blueprintSelectColumns,
		b.ID, b.ProjectID, b.Name, b.Version, def, idempCol)
	out, err := scanBlueprint(row)
	if err == nil {
		return out, nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		switch pgErr.ConstraintName {
		case "blueprints_idempotency":
			if b.IdempotencyKey != "" {
				return s.GetBlueprintByIdempotency(ctx, b.IdempotencyKey)
			}
		case "blueprints_project_id_name_version_key":
			return domain.Blueprint{}, fmt.Errorf("%w: blueprint %q v%d already exists", ErrConflict, b.Name, b.Version)
		}
	}
	return domain.Blueprint{}, fmt.Errorf("pgstore: insert blueprint: %w", err)
}

// AppendBlueprintVersion writes a new version of (project, name) with the
// supplied definition. Version is server-monotonic: MAX(version) + 1.
func (s *Store) AppendBlueprintVersion(ctx context.Context, projectID uuid.UUID, name string, definition map[string]any) (domain.Blueprint, error) {
	if projectID == uuid.Nil || name == "" {
		return domain.Blueprint{}, fmt.Errorf("pgstore: blueprint: project_id and name required")
	}
	def := []byte(`{}`)
	if len(definition) > 0 {
		raw, err := json.Marshal(definition)
		if err != nil {
			return domain.Blueprint{}, fmt.Errorf("pgstore: marshal blueprint definition: %w", err)
		}
		def = raw
	}
	var out domain.Blueprint
	err := s.WithTx(ctx, func(tx pgx.Tx) error {
		var maxV int
		if err := tx.QueryRow(ctx, `
			SELECT COALESCE(MAX(version), 0)
			FROM blueprints
			WHERE project_id = $1 AND name = $2`, projectID, name).Scan(&maxV); err != nil {
			return fmt.Errorf("read max version: %w", err)
		}
		if maxV == 0 {
			return fmt.Errorf("%w: no blueprint named %q in project; call blueprint.create first", ErrNotFound, name)
		}
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		row := tx.QueryRow(ctx, `
			INSERT INTO blueprints (id, project_id, name, version, definition_jsonb)
			VALUES ($1, $2, $3, $4, $5::jsonb)
			RETURNING `+blueprintSelectColumns,
			id, projectID, name, maxV+1, def)
		got, err := scanBlueprint(row)
		if err != nil {
			return err
		}
		out = got
		return nil
	})
	if err != nil {
		return domain.Blueprint{}, fmt.Errorf("pgstore: append blueprint version: %w", err)
	}
	return out, nil
}

// GetBlueprint by id.
func (s *Store) GetBlueprint(ctx context.Context, id uuid.UUID) (domain.Blueprint, error) {
	row := s.pool.QueryRow(ctx, `SELECT `+blueprintSelectColumns+` FROM blueprints WHERE id = $1`, id)
	out, err := scanBlueprint(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Blueprint{}, ErrNotFound
	}
	if err != nil {
		return domain.Blueprint{}, fmt.Errorf("pgstore: get blueprint: %w", err)
	}
	return out, nil
}

// GetLatestBlueprint returns the highest-version blueprint for (project, name).
func (s *Store) GetLatestBlueprint(ctx context.Context, projectID uuid.UUID, name string) (domain.Blueprint, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT `+blueprintSelectColumns+`
		FROM blueprints
		WHERE project_id = $1 AND name = $2
		ORDER BY version DESC
		LIMIT 1`, projectID, name)
	out, err := scanBlueprint(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Blueprint{}, ErrNotFound
	}
	if err != nil {
		return domain.Blueprint{}, fmt.Errorf("pgstore: get latest blueprint: %w", err)
	}
	return out, nil
}

// GetBlueprintByVersion returns the specific (project, name, version) row.
func (s *Store) GetBlueprintByVersion(ctx context.Context, projectID uuid.UUID, name string, version int) (domain.Blueprint, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT `+blueprintSelectColumns+`
		FROM blueprints
		WHERE project_id = $1 AND name = $2 AND version = $3`, projectID, name, version)
	out, err := scanBlueprint(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Blueprint{}, ErrNotFound
	}
	if err != nil {
		return domain.Blueprint{}, fmt.Errorf("pgstore: get blueprint by version: %w", err)
	}
	return out, nil
}

// GetBlueprintByIdempotency returns the blueprint previously written with key.
func (s *Store) GetBlueprintByIdempotency(ctx context.Context, key string) (domain.Blueprint, error) {
	row := s.pool.QueryRow(ctx, `SELECT `+blueprintSelectColumns+` FROM blueprints WHERE idempotency_key = $1`, key)
	out, err := scanBlueprint(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Blueprint{}, ErrNotFound
	}
	if err != nil {
		return domain.Blueprint{}, fmt.Errorf("pgstore: get blueprint by idempotency: %w", err)
	}
	return out, nil
}

// ListBlueprintsFilter narrows the list returned by [Store.ListBlueprints].
type ListBlueprintsFilter struct {
	// LatestOnly returns only the highest-version row for each unique
	// (project, name); useful for the agent-facing "what blueprints exist
	// in this project?" query. Default behaviour returns every row.
	LatestOnly bool
}

// ListBlueprints in projectID, alphabetical by name then version-asc.
func (s *Store) ListBlueprints(ctx context.Context, projectID uuid.UUID, f ListBlueprintsFilter) ([]domain.Blueprint, error) {
	var q string
	if f.LatestOnly {
		q = `
			SELECT ` + blueprintSelectColumns + `
			FROM blueprints b
			WHERE project_id = $1 AND version = (
			    SELECT MAX(version) FROM blueprints
			    WHERE project_id = b.project_id AND name = b.name
			)
			ORDER BY name ASC`
	} else {
		q = `
			SELECT ` + blueprintSelectColumns + `
			FROM blueprints
			WHERE project_id = $1
			ORDER BY name ASC, version ASC`
	}
	rows, err := s.pool.Query(ctx, q, projectID)
	if err != nil {
		return nil, fmt.Errorf("pgstore: list blueprints: %w", err)
	}
	defer rows.Close()
	var out []domain.Blueprint
	for rows.Next() {
		b, err := scanBlueprint(rows)
		if err != nil {
			return nil, fmt.Errorf("pgstore: scan blueprint: %w", err)
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// DeleteBlueprint removes a single blueprint version (and its modes).
// Other versions of the same name are untouched.
func (s *Store) DeleteBlueprint(ctx context.Context, id uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM blueprints WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("pgstore: delete blueprint: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// IsLatestBlueprint reports whether the given blueprint is the highest-version
// row for its (project, name) tuple.
func (s *Store) IsLatestBlueprint(ctx context.Context, b domain.Blueprint) (bool, error) {
	var maxV int
	err := s.pool.QueryRow(ctx, `
		SELECT MAX(version) FROM blueprints
		WHERE project_id = $1 AND name = $2`, b.ProjectID, b.Name).Scan(&maxV)
	if err != nil {
		return false, fmt.Errorf("pgstore: max blueprint version: %w", err)
	}
	return b.Version == maxV, nil
}

func scanBlueprint(r scannable) (domain.Blueprint, error) {
	var (
		b   domain.Blueprint
		def []byte
		key *string
	)
	if err := r.Scan(
		&b.ID, &b.ProjectID, &b.Name, &b.Version, &def,
		&key, &b.CreatedAt, &b.UpdatedAt,
	); err != nil {
		return domain.Blueprint{}, err
	}
	if len(def) > 0 {
		_ = json.Unmarshal(def, &b.Definition)
	}
	if key != nil {
		b.IdempotencyKey = *key
	}
	return b, nil
}
