package pgstore

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/manuelibar/workbench/internal/domain"
)

const skillSelectColumns = `id, project_id, name, body_md,
    idempotency_key, created_at, updated_at`

// CreateSkill writes a new skill. Idempotency-key replays return the original.
func (s *Store) CreateSkill(ctx context.Context, sk domain.Skill) (domain.Skill, error) {
	if sk.ProjectID == uuid.Nil {
		return domain.Skill{}, fmt.Errorf("pgstore: skill: project_id required")
	}
	if sk.Name == "" {
		return domain.Skill{}, fmt.Errorf("pgstore: skill: name required")
	}
	if sk.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return domain.Skill{}, err
		}
		sk.ID = id
	}
	var idempCol any
	if sk.IdempotencyKey != "" {
		idempCol = sk.IdempotencyKey
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO skills (id, project_id, name, body_md, idempotency_key)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING `+skillSelectColumns,
		sk.ID, sk.ProjectID, sk.Name, sk.BodyMD, idempCol)
	out, err := scanSkill(row)
	if err == nil {
		return out, nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		switch pgErr.ConstraintName {
		case "skills_idempotency":
			if sk.IdempotencyKey != "" {
				return s.GetSkillByIdempotency(ctx, sk.IdempotencyKey)
			}
		case "skills_unique_name_per_project":
			return domain.Skill{}, fmt.Errorf("%w: skill %q already exists in project", ErrConflict, sk.Name)
		}
	}
	return domain.Skill{}, fmt.Errorf("pgstore: insert skill: %w", err)
}

// GetSkill by id.
func (s *Store) GetSkill(ctx context.Context, id uuid.UUID) (domain.Skill, error) {
	row := s.pool.QueryRow(ctx, `SELECT `+skillSelectColumns+` FROM skills WHERE id = $1`, id)
	out, err := scanSkill(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Skill{}, ErrNotFound
	}
	if err != nil {
		return domain.Skill{}, fmt.Errorf("pgstore: get skill: %w", err)
	}
	return out, nil
}

// GetSkillByName fetches a skill by its (project, name) tuple.
func (s *Store) GetSkillByName(ctx context.Context, projectID uuid.UUID, name string) (domain.Skill, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT `+skillSelectColumns+`
		FROM skills WHERE project_id = $1 AND name = $2`,
		projectID, name)
	out, err := scanSkill(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Skill{}, ErrNotFound
	}
	if err != nil {
		return domain.Skill{}, fmt.Errorf("pgstore: get skill by name: %w", err)
	}
	return out, nil
}

// GetSkillByIdempotency.
func (s *Store) GetSkillByIdempotency(ctx context.Context, key string) (domain.Skill, error) {
	row := s.pool.QueryRow(ctx, `SELECT `+skillSelectColumns+` FROM skills WHERE idempotency_key = $1`, key)
	out, err := scanSkill(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Skill{}, ErrNotFound
	}
	if err != nil {
		return domain.Skill{}, fmt.Errorf("pgstore: get skill by idempotency: %w", err)
	}
	return out, nil
}

// ListSkills in projectID, alphabetical by name.
func (s *Store) ListSkills(ctx context.Context, projectID uuid.UUID) ([]domain.Skill, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT `+skillSelectColumns+`
		FROM skills WHERE project_id = $1 ORDER BY name ASC`,
		projectID)
	if err != nil {
		return nil, fmt.Errorf("pgstore: list skills: %w", err)
	}
	defer rows.Close()
	var out []domain.Skill
	for rows.Next() {
		sk, err := scanSkill(rows)
		if err != nil {
			return nil, fmt.Errorf("pgstore: scan skill: %w", err)
		}
		out = append(out, sk)
	}
	return out, rows.Err()
}

// UpdateSkillFields holds the optional patches applied by [Store.UpdateSkill].
type UpdateSkillFields struct {
	Name   *string
	BodyMD *string
}

// UpdateSkill patches the named skill. Returns the updated row.
func (s *Store) UpdateSkill(ctx context.Context, id uuid.UUID, f UpdateSkillFields) (domain.Skill, error) {
	sets := []string{"updated_at = now()"}
	args := []any{id}
	if f.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", len(args)+1))
		args = append(args, *f.Name)
	}
	if f.BodyMD != nil {
		sets = append(sets, fmt.Sprintf("body_md = $%d", len(args)+1))
		args = append(args, *f.BodyMD)
	}
	if len(sets) == 1 {
		return s.GetSkill(ctx, id)
	}
	q := `UPDATE skills SET ` + strings.Join(sets, ", ") + ` WHERE id = $1 RETURNING ` + skillSelectColumns
	row := s.pool.QueryRow(ctx, q, args...)
	out, err := scanSkill(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Skill{}, ErrNotFound
	}
	if err != nil {
		return domain.Skill{}, fmt.Errorf("pgstore: update skill: %w", err)
	}
	return out, nil
}

// DeleteSkill.
func (s *Store) DeleteSkill(ctx context.Context, id uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM skills WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("pgstore: delete skill: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func scanSkill(r scannable) (domain.Skill, error) {
	var (
		sk  domain.Skill
		key *string
	)
	if err := r.Scan(
		&sk.ID, &sk.ProjectID, &sk.Name, &sk.BodyMD,
		&key, &sk.CreatedAt, &sk.UpdatedAt,
	); err != nil {
		return domain.Skill{}, err
	}
	if key != nil {
		sk.IdempotencyKey = *key
	}
	return sk, nil
}
