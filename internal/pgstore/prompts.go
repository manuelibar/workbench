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

const promptSelectColumns = `id, project_id, name, description, body, args_jsonb,
    idempotency_key, created_at, updated_at`

// CreatePrompt writes a new prompt. Idempotency-key replays return the original.
func (s *Store) CreatePrompt(ctx context.Context, p domain.Prompt) (domain.Prompt, error) {
	if p.ProjectID == uuid.Nil {
		return domain.Prompt{}, fmt.Errorf("pgstore: prompt: project_id required")
	}
	if p.Name == "" {
		return domain.Prompt{}, fmt.Errorf("pgstore: prompt: name required")
	}
	if p.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return domain.Prompt{}, err
		}
		p.ID = id
	}
	if p.Args == nil {
		p.Args = []domain.PromptArg{}
	}
	argsJSON, err := json.Marshal(p.Args)
	if err != nil {
		return domain.Prompt{}, fmt.Errorf("pgstore: marshal prompt args: %w", err)
	}
	var idempCol any
	if p.IdempotencyKey != "" {
		idempCol = p.IdempotencyKey
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO prompts (id, project_id, name, description, body, args_jsonb, idempotency_key)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7)
		RETURNING `+promptSelectColumns,
		p.ID, p.ProjectID, p.Name, p.Description, p.Body, argsJSON, idempCol)
	out, scanErr := scanPrompt(row)
	if scanErr == nil {
		return out, nil
	}
	var pgErr *pgconn.PgError
	if errors.As(scanErr, &pgErr) && pgErr.Code == pgUniqueViolation {
		switch pgErr.ConstraintName {
		case "prompts_idempotency":
			if p.IdempotencyKey != "" {
				return s.GetPromptByIdempotency(ctx, p.IdempotencyKey)
			}
		case "prompts_unique_name_per_project":
			return domain.Prompt{}, fmt.Errorf("%w: prompt %q already exists in project", ErrConflict, p.Name)
		}
	}
	return domain.Prompt{}, fmt.Errorf("pgstore: insert prompt: %w", scanErr)
}

// GetPrompt by id.
func (s *Store) GetPrompt(ctx context.Context, id uuid.UUID) (domain.Prompt, error) {
	row := s.pool.QueryRow(ctx, `SELECT `+promptSelectColumns+` FROM prompts WHERE id = $1`, id)
	out, err := scanPrompt(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Prompt{}, ErrNotFound
	}
	if err != nil {
		return domain.Prompt{}, fmt.Errorf("pgstore: get prompt: %w", err)
	}
	return out, nil
}

// GetPromptByName fetches a prompt by its (project, name) tuple.
func (s *Store) GetPromptByName(ctx context.Context, projectID uuid.UUID, name string) (domain.Prompt, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT `+promptSelectColumns+`
		FROM prompts WHERE project_id = $1 AND name = $2`,
		projectID, name)
	out, err := scanPrompt(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Prompt{}, ErrNotFound
	}
	if err != nil {
		return domain.Prompt{}, fmt.Errorf("pgstore: get prompt by name: %w", err)
	}
	return out, nil
}

// GetPromptByIdempotency.
func (s *Store) GetPromptByIdempotency(ctx context.Context, key string) (domain.Prompt, error) {
	row := s.pool.QueryRow(ctx, `SELECT `+promptSelectColumns+` FROM prompts WHERE idempotency_key = $1`, key)
	out, err := scanPrompt(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Prompt{}, ErrNotFound
	}
	if err != nil {
		return domain.Prompt{}, fmt.Errorf("pgstore: get prompt by idempotency: %w", err)
	}
	return out, nil
}

// ListPrompts in projectID, alphabetical by name.
func (s *Store) ListPrompts(ctx context.Context, projectID uuid.UUID) ([]domain.Prompt, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT `+promptSelectColumns+`
		FROM prompts WHERE project_id = $1 ORDER BY name ASC`,
		projectID)
	if err != nil {
		return nil, fmt.Errorf("pgstore: list prompts: %w", err)
	}
	defer rows.Close()
	var out []domain.Prompt
	for rows.Next() {
		p, err := scanPrompt(rows)
		if err != nil {
			return nil, fmt.Errorf("pgstore: scan prompt: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// UpdatePromptFields holds the optional patches applied by [Store.UpdatePrompt].
type UpdatePromptFields struct {
	Name        *string
	Description *string
	Body        *string
	Args        *[]domain.PromptArg
}

// UpdatePrompt patches the named prompt. Returns the updated row.
func (s *Store) UpdatePrompt(ctx context.Context, id uuid.UUID, f UpdatePromptFields) (domain.Prompt, error) {
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
	if f.Body != nil {
		sets = append(sets, fmt.Sprintf("body = $%d", len(args)+1))
		args = append(args, *f.Body)
	}
	if f.Args != nil {
		raw, err := json.Marshal(*f.Args)
		if err != nil {
			return domain.Prompt{}, fmt.Errorf("pgstore: marshal prompt args: %w", err)
		}
		sets = append(sets, fmt.Sprintf("args_jsonb = $%d::jsonb", len(args)+1))
		args = append(args, raw)
	}
	if len(sets) == 1 {
		return s.GetPrompt(ctx, id)
	}
	q := `UPDATE prompts SET ` + strings.Join(sets, ", ") + ` WHERE id = $1 RETURNING ` + promptSelectColumns
	row := s.pool.QueryRow(ctx, q, args...)
	out, err := scanPrompt(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Prompt{}, ErrNotFound
	}
	if err != nil {
		return domain.Prompt{}, fmt.Errorf("pgstore: update prompt: %w", err)
	}
	return out, nil
}

// DeletePrompt.
func (s *Store) DeletePrompt(ctx context.Context, id uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM prompts WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("pgstore: delete prompt: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func scanPrompt(r scannable) (domain.Prompt, error) {
	var (
		p       domain.Prompt
		key     *string
		argsRaw []byte
	)
	if err := r.Scan(
		&p.ID, &p.ProjectID, &p.Name, &p.Description, &p.Body, &argsRaw,
		&key, &p.CreatedAt, &p.UpdatedAt,
	); err != nil {
		return domain.Prompt{}, err
	}
	if len(argsRaw) > 0 {
		_ = json.Unmarshal(argsRaw, &p.Args)
	}
	if key != nil {
		p.IdempotencyKey = *key
	}
	return p, nil
}
