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

const artifactSelectColumns = `id, project_id, type, status, parents,
    latest_version, idempotency_key, created_at, updated_at`

// CreateArtifactInput bundles the optional initial-version content with the
// artifact-row metadata. `Content` and `ContentText` are written as
// version 1 if either is non-empty.
type CreateArtifactInput struct {
	Artifact    domain.Artifact
	Content     map[string]any
	ContentText string
}

// CreateArtifact writes a new artifact row plus (if any content is supplied)
// its initial version, in a single transaction. Idempotency-key replays
// return the original artifact (no new version is appended).
func (s *Store) CreateArtifact(ctx context.Context, in CreateArtifactInput) (domain.Artifact, error) {
	a := in.Artifact
	if a.ProjectID == uuid.Nil {
		return domain.Artifact{}, fmt.Errorf("pgstore: artifact: project_id required")
	}
	if a.Type == "" {
		return domain.Artifact{}, fmt.Errorf("pgstore: artifact: type required")
	}
	if a.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return domain.Artifact{}, fmt.Errorf("pgstore: generate artifact id: %w", err)
		}
		a.ID = id
	}
	if a.Status == "" {
		a.Status = domain.ArtifactStatusDraft
	}
	if a.Parents == nil {
		a.Parents = []uuid.UUID{}
	}

	hasContent := len(in.Content) > 0 || in.ContentText != ""
	if hasContent {
		a.LatestVersion = 1
	}

	contentJSON := []byte(`{}`)
	if len(in.Content) > 0 {
		raw, err := json.Marshal(in.Content)
		if err != nil {
			return domain.Artifact{}, fmt.Errorf("pgstore: marshal artifact content: %w", err)
		}
		contentJSON = raw
	}
	var idempCol any
	if a.IdempotencyKey != "" {
		idempCol = a.IdempotencyKey
	}

	var out domain.Artifact
	err := s.WithTx(ctx, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, `
			INSERT INTO artifacts (id, project_id, type, status, parents, latest_version, idempotency_key)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING `+artifactSelectColumns,
			a.ID, a.ProjectID, a.Type, a.Status, a.Parents, a.LatestVersion, idempCol)
		got, err := scanArtifact(row)
		if err != nil {
			return err
		}
		if hasContent {
			if _, err := tx.Exec(ctx, `
				INSERT INTO artifact_versions (artifact_id, version, content_jsonb, content_text)
				VALUES ($1, $2, $3::jsonb, $4)
			`, got.ID, 1, contentJSON, in.ContentText); err != nil {
				return fmt.Errorf("insert artifact version: %w", err)
			}
		}
		out = got
		return nil
	})
	if err == nil {
		return out, nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		if pgErr.ConstraintName == "artifacts_idempotency" && a.IdempotencyKey != "" {
			return s.GetArtifactByIdempotency(ctx, a.IdempotencyKey)
		}
	}
	return domain.Artifact{}, fmt.Errorf("pgstore: insert artifact: %w", err)
}

// GetArtifact returns the artifact row.
func (s *Store) GetArtifact(ctx context.Context, id uuid.UUID) (domain.Artifact, error) {
	row := s.pool.QueryRow(ctx, `SELECT `+artifactSelectColumns+` FROM artifacts WHERE id = $1`, id)
	a, err := scanArtifact(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Artifact{}, ErrNotFound
	}
	if err != nil {
		return domain.Artifact{}, fmt.Errorf("pgstore: get artifact: %w", err)
	}
	return a, nil
}

// GetArtifactByIdempotency returns the artifact previously written with key.
func (s *Store) GetArtifactByIdempotency(ctx context.Context, key string) (domain.Artifact, error) {
	row := s.pool.QueryRow(ctx, `SELECT `+artifactSelectColumns+` FROM artifacts WHERE idempotency_key = $1`, key)
	a, err := scanArtifact(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Artifact{}, ErrNotFound
	}
	if err != nil {
		return domain.Artifact{}, fmt.Errorf("pgstore: get artifact by idempotency: %w", err)
	}
	return a, nil
}

// GetArtifactVersion returns a specific version's content. version=0 means
// "latest".
func (s *Store) GetArtifactVersion(ctx context.Context, artifactID uuid.UUID, version int) (domain.ArtifactVersion, error) {
	a, err := s.GetArtifact(ctx, artifactID)
	if err != nil {
		return domain.ArtifactVersion{}, err
	}
	if version <= 0 {
		version = a.LatestVersion
	}
	if version <= 0 {
		return domain.ArtifactVersion{}, ErrNotFound
	}
	row := s.pool.QueryRow(ctx, `
		SELECT artifact_id, version, content_jsonb, content_text, created_at
		FROM artifact_versions
		WHERE artifact_id = $1 AND version = $2`, artifactID, version)
	var (
		v       domain.ArtifactVersion
		rawJSON []byte
	)
	if err := row.Scan(&v.ArtifactID, &v.Version, &rawJSON, &v.ContentText, &v.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ArtifactVersion{}, ErrNotFound
		}
		return domain.ArtifactVersion{}, fmt.Errorf("pgstore: get artifact version: %w", err)
	}
	if len(rawJSON) > 0 {
		_ = json.Unmarshal(rawJSON, &v.Content)
	}
	return v, nil
}

// ListArtifactsFilter narrows the list returned by [Store.ListArtifacts].
type ListArtifactsFilter struct {
	Type   string
	Status string
	Limit  int
}

// ListArtifacts returns artifacts in projectID, most-recent-first.
func (s *Store) ListArtifacts(ctx context.Context, projectID uuid.UUID, f ListArtifactsFilter) ([]domain.Artifact, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	q := `SELECT ` + artifactSelectColumns + ` FROM artifacts WHERE project_id = $1`
	args := []any{projectID}
	if f.Type != "" {
		q += fmt.Sprintf(" AND type = $%d", len(args)+1)
		args = append(args, f.Type)
	}
	if f.Status != "" {
		q += fmt.Sprintf(" AND status = $%d", len(args)+1)
		args = append(args, f.Status)
	}
	q += fmt.Sprintf(" ORDER BY updated_at DESC LIMIT $%d", len(args)+1)
	args = append(args, limit)

	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("pgstore: list artifacts: %w", err)
	}
	defer rows.Close()
	var out []domain.Artifact
	for rows.Next() {
		a, err := scanArtifact(rows)
		if err != nil {
			return nil, fmt.Errorf("pgstore: scan artifact: %w", err)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// AppendArtifactVersion writes a new version with content/contentText and
// bumps the artifact's latest_version pointer in a single transaction.
// Returns the updated artifact.
func (s *Store) AppendArtifactVersion(ctx context.Context, artifactID uuid.UUID, content map[string]any, contentText string) (domain.Artifact, int, error) {
	contentJSON := []byte(`{}`)
	if len(content) > 0 {
		raw, err := json.Marshal(content)
		if err != nil {
			return domain.Artifact{}, 0, fmt.Errorf("pgstore: marshal version content: %w", err)
		}
		contentJSON = raw
	}
	var (
		updated    domain.Artifact
		newVersion int
	)
	err := s.WithTx(ctx, func(tx pgx.Tx) error {
		var current int
		if err := tx.QueryRow(ctx, `SELECT latest_version FROM artifacts WHERE id = $1 FOR UPDATE`, artifactID).Scan(&current); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("read latest_version: %w", err)
		}
		newVersion = current + 1
		if _, err := tx.Exec(ctx, `
			INSERT INTO artifact_versions (artifact_id, version, content_jsonb, content_text)
			VALUES ($1, $2, $3::jsonb, $4)
		`, artifactID, newVersion, contentJSON, contentText); err != nil {
			return fmt.Errorf("insert version: %w", err)
		}
		row := tx.QueryRow(ctx, `
			UPDATE artifacts
			SET latest_version = $2, updated_at = now()
			WHERE id = $1
			RETURNING `+artifactSelectColumns,
			artifactID, newVersion)
		got, err := scanArtifact(row)
		if err != nil {
			return err
		}
		updated = got
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return domain.Artifact{}, 0, ErrNotFound
		}
		return domain.Artifact{}, 0, fmt.Errorf("pgstore: append artifact version: %w", err)
	}
	return updated, newVersion, nil
}

// SetArtifactStatus updates only the status field. Status must be one of the
// values accepted by the schema CHECK constraint.
func (s *Store) SetArtifactStatus(ctx context.Context, id uuid.UUID, status string) (domain.Artifact, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE artifacts
		SET status = $2, updated_at = now()
		WHERE id = $1
		RETURNING `+artifactSelectColumns,
		id, status)
	a, err := scanArtifact(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Artifact{}, ErrNotFound
	}
	if err != nil {
		return domain.Artifact{}, fmt.Errorf("pgstore: set artifact status: %w", err)
	}
	return a, nil
}

// AttachArtifactParent appends parentID to the artifact's parents array
// (idempotently — duplicates are ignored).
func (s *Store) AttachArtifactParent(ctx context.Context, id, parentID uuid.UUID) (domain.Artifact, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE artifacts
		SET parents = (
		  CASE WHEN $2 = ANY (parents) THEN parents
		       ELSE array_append(parents, $2) END
		),
		    updated_at = now()
		WHERE id = $1
		RETURNING `+artifactSelectColumns,
		id, parentID)
	a, err := scanArtifact(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Artifact{}, ErrNotFound
	}
	if err != nil {
		return domain.Artifact{}, fmt.Errorf("pgstore: attach artifact parent: %w", err)
	}
	return a, nil
}

// DeleteArtifact removes the artifact and all its versions.
func (s *Store) DeleteArtifact(ctx context.Context, id uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM artifacts WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("pgstore: delete artifact: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func scanArtifact(r scannable) (domain.Artifact, error) {
	var (
		a   domain.Artifact
		key *string
	)
	if err := r.Scan(
		&a.ID, &a.ProjectID, &a.Type, &a.Status, &a.Parents,
		&a.LatestVersion, &key, &a.CreatedAt, &a.UpdatedAt,
	); err != nil {
		return domain.Artifact{}, err
	}
	if key != nil {
		a.IdempotencyKey = *key
	}
	return a, nil
}
