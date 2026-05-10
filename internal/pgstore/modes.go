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

const modeSelectColumns = `id, blueprint_id, name, system_prompt,
    capabilities_jsonb, definition_jsonb, idempotency_key, created_at, updated_at`

// CreateMode inserts a new mode under blueprintID. The blueprint must be
// the latest version for its (project, name); otherwise [ErrConflict] is
// returned. (Modes are only mutable on the latest version.)
func (s *Store) CreateMode(ctx context.Context, m domain.Mode) (domain.Mode, error) {
	if m.BlueprintID == uuid.Nil {
		return domain.Mode{}, fmt.Errorf("pgstore: mode: blueprint_id required")
	}
	if m.Name == "" {
		return domain.Mode{}, fmt.Errorf("pgstore: mode: name required")
	}
	if m.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return domain.Mode{}, err
		}
		m.ID = id
	}
	caps := []byte(`{}`)
	if len(m.Capabilities) > 0 {
		raw, err := json.Marshal(m.Capabilities)
		if err != nil {
			return domain.Mode{}, fmt.Errorf("pgstore: marshal capabilities: %w", err)
		}
		caps = raw
	}
	def := []byte(`{}`)
	if len(m.Definition) > 0 {
		raw, err := json.Marshal(m.Definition)
		if err != nil {
			return domain.Mode{}, fmt.Errorf("pgstore: marshal mode definition: %w", err)
		}
		def = raw
	}
	var idempCol any
	if m.IdempotencyKey != "" {
		idempCol = m.IdempotencyKey
	}

	var out domain.Mode
	err := s.WithTx(ctx, func(tx pgx.Tx) error {
		if err := assertBlueprintIsLatestTx(ctx, tx, m.BlueprintID); err != nil {
			return err
		}
		row := tx.QueryRow(ctx, `
			INSERT INTO modes (id, blueprint_id, name, system_prompt, capabilities_jsonb, definition_jsonb, idempotency_key)
			VALUES ($1, $2, $3, $4, $5::jsonb, $6::jsonb, $7)
			RETURNING `+modeSelectColumns,
			m.ID, m.BlueprintID, m.Name, m.SystemPrompt, caps, def, idempCol)
		got, err := scanMode(row)
		if err != nil {
			return err
		}
		out = got
		return nil
	})
	if err == nil {
		return out, nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		switch pgErr.ConstraintName {
		case "modes_idempotency":
			if m.IdempotencyKey != "" {
				return s.GetModeByIdempotency(ctx, m.IdempotencyKey)
			}
		case "modes_blueprint_id_name_key":
			return domain.Mode{}, fmt.Errorf("%w: mode %q already exists in this blueprint version", ErrConflict, m.Name)
		}
	}
	return domain.Mode{}, fmt.Errorf("pgstore: insert mode: %w", err)
}

// GetMode by id.
func (s *Store) GetMode(ctx context.Context, id uuid.UUID) (domain.Mode, error) {
	row := s.pool.QueryRow(ctx, `SELECT `+modeSelectColumns+` FROM modes WHERE id = $1`, id)
	out, err := scanMode(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Mode{}, ErrNotFound
	}
	if err != nil {
		return domain.Mode{}, fmt.Errorf("pgstore: get mode: %w", err)
	}
	return out, nil
}

// GetModeByName fetches a mode by its (blueprint, name) tuple.
func (s *Store) GetModeByName(ctx context.Context, blueprintID uuid.UUID, name string) (domain.Mode, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT `+modeSelectColumns+`
		FROM modes WHERE blueprint_id = $1 AND name = $2`,
		blueprintID, name)
	out, err := scanMode(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Mode{}, ErrNotFound
	}
	if err != nil {
		return domain.Mode{}, fmt.Errorf("pgstore: get mode by name: %w", err)
	}
	return out, nil
}

// GetModeByIdempotency.
func (s *Store) GetModeByIdempotency(ctx context.Context, key string) (domain.Mode, error) {
	row := s.pool.QueryRow(ctx, `SELECT `+modeSelectColumns+` FROM modes WHERE idempotency_key = $1`, key)
	out, err := scanMode(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Mode{}, ErrNotFound
	}
	if err != nil {
		return domain.Mode{}, fmt.Errorf("pgstore: get mode by idempotency: %w", err)
	}
	return out, nil
}

// ListModes returns the modes inside blueprintID, alphabetical by name.
func (s *Store) ListModes(ctx context.Context, blueprintID uuid.UUID) ([]domain.Mode, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT `+modeSelectColumns+`
		FROM modes WHERE blueprint_id = $1 ORDER BY name ASC`,
		blueprintID)
	if err != nil {
		return nil, fmt.Errorf("pgstore: list modes: %w", err)
	}
	defer rows.Close()
	var out []domain.Mode
	for rows.Next() {
		m, err := scanMode(rows)
		if err != nil {
			return nil, fmt.Errorf("pgstore: scan mode: %w", err)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// UpdateModeFields holds the optional patches applied by [Store.UpdateMode].
type UpdateModeFields struct {
	Name         *string
	SystemPrompt *string
	Capabilities *map[string]any
	Definition   *map[string]any
}

// UpdateMode patches the named mode. Only allowed on a mode whose blueprint
// is the latest version for its (project, name) — otherwise returns
// [ErrConflict] with a "create a new blueprint version first" message.
func (s *Store) UpdateMode(ctx context.Context, id uuid.UUID, f UpdateModeFields) (domain.Mode, error) {
	var out domain.Mode
	err := s.WithTx(ctx, func(tx pgx.Tx) error {
		var bpID uuid.UUID
		if err := tx.QueryRow(ctx, `SELECT blueprint_id FROM modes WHERE id = $1`, id).Scan(&bpID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("read mode: %w", err)
		}
		if err := assertBlueprintIsLatestTx(ctx, tx, bpID); err != nil {
			return err
		}

		sets := []string{"updated_at = now()"}
		args := []any{id}
		if f.Name != nil {
			sets = append(sets, fmt.Sprintf("name = $%d", len(args)+1))
			args = append(args, *f.Name)
		}
		if f.SystemPrompt != nil {
			sets = append(sets, fmt.Sprintf("system_prompt = $%d", len(args)+1))
			args = append(args, *f.SystemPrompt)
		}
		if f.Capabilities != nil {
			raw, err := json.Marshal(*f.Capabilities)
			if err != nil {
				return fmt.Errorf("marshal capabilities: %w", err)
			}
			sets = append(sets, fmt.Sprintf("capabilities_jsonb = $%d::jsonb", len(args)+1))
			args = append(args, raw)
		}
		if f.Definition != nil {
			raw, err := json.Marshal(*f.Definition)
			if err != nil {
				return fmt.Errorf("marshal definition: %w", err)
			}
			sets = append(sets, fmt.Sprintf("definition_jsonb = $%d::jsonb", len(args)+1))
			args = append(args, raw)
		}
		if len(sets) == 1 {
			row := tx.QueryRow(ctx, `SELECT `+modeSelectColumns+` FROM modes WHERE id = $1`, id)
			got, err := scanMode(row)
			if err != nil {
				return err
			}
			out = got
			return nil
		}
		q := `UPDATE modes SET ` + strings.Join(sets, ", ") + ` WHERE id = $1 RETURNING ` + modeSelectColumns
		row := tx.QueryRow(ctx, q, args...)
		got, err := scanMode(row)
		if err != nil {
			return err
		}
		out = got
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return domain.Mode{}, ErrNotFound
		}
		return domain.Mode{}, fmt.Errorf("pgstore: update mode: %w", err)
	}
	return out, nil
}

// DeleteMode removes a mode. Like update, only allowed on the latest
// blueprint version.
func (s *Store) DeleteMode(ctx context.Context, id uuid.UUID) error {
	return s.WithTx(ctx, func(tx pgx.Tx) error {
		var bpID uuid.UUID
		if err := tx.QueryRow(ctx, `SELECT blueprint_id FROM modes WHERE id = $1`, id).Scan(&bpID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("read mode: %w", err)
		}
		if err := assertBlueprintIsLatestTx(ctx, tx, bpID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `DELETE FROM modes WHERE id = $1`, id); err != nil {
			return fmt.Errorf("delete mode: %w", err)
		}
		return nil
	})
}

func assertBlueprintIsLatestTx(ctx context.Context, tx pgx.Tx, blueprintID uuid.UUID) error {
	var (
		projectID uuid.UUID
		name      string
		version   int
	)
	if err := tx.QueryRow(ctx, `
		SELECT project_id, name, version FROM blueprints WHERE id = $1`,
		blueprintID).Scan(&projectID, &name, &version); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("read blueprint: %w", err)
	}
	var maxV int
	if err := tx.QueryRow(ctx, `
		SELECT MAX(version) FROM blueprints
		WHERE project_id = $1 AND name = $2`, projectID, name).Scan(&maxV); err != nil {
		return fmt.Errorf("read max version: %w", err)
	}
	if version != maxV {
		return fmt.Errorf("%w: mode change rejected — blueprint %q is at v%d but the latest is v%d; create a new blueprint version first via blueprint.update",
			ErrConflict, name, version, maxV)
	}
	return nil
}

func scanMode(r scannable) (domain.Mode, error) {
	var (
		m    domain.Mode
		caps []byte
		def  []byte
		key  *string
	)
	if err := r.Scan(
		&m.ID, &m.BlueprintID, &m.Name, &m.SystemPrompt,
		&caps, &def, &key, &m.CreatedAt, &m.UpdatedAt,
	); err != nil {
		return domain.Mode{}, err
	}
	if len(caps) > 0 {
		_ = json.Unmarshal(caps, &m.Capabilities)
	}
	if len(def) > 0 {
		_ = json.Unmarshal(def, &m.Definition)
	}
	if key != nil {
		m.IdempotencyKey = *key
	}
	return m, nil
}
