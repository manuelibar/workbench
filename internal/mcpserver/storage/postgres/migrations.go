package postgres

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrate applies embedded SQL migrations exactly once, ordered by filename.
func Migrate(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("database handle is required")
	}
	if _, err := db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS workbench_schema_migrations (
    version text PRIMARY KEY,
    applied_at timestamptz NOT NULL DEFAULT now()
)`); err != nil {
		return fmt.Errorf("ensure schema migrations table: %w", err)
	}

	if _, err := db.ExecContext(ctx, `SELECT pg_advisory_lock(hashtext('workbench_migrations'))`); err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}
	defer db.ExecContext(context.Background(), `SELECT pg_advisory_unlock(hashtext('workbench_migrations'))`)

	applied, err := appliedMigrations(ctx, db)
	if err != nil {
		return err
	}
	files, err := migrationFiles()
	if err != nil {
		return err
	}
	for _, file := range files {
		version, err := migrationVersion(file)
		if err != nil {
			return err
		}
		if applied[version] {
			continue
		}
		sqlText, err := readMigration(file)
		if err != nil {
			return err
		}
		if err := applyMigration(ctx, db, version, sqlText); err != nil {
			return err
		}
	}
	return nil
}

func appliedMigrations(ctx context.Context, db *sql.DB) (map[string]bool, error) {
	rows, err := db.QueryContext(ctx, `SELECT version FROM workbench_schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("list applied migrations: %w", err)
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("scan applied migration: %w", err)
		}
		out[version] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate applied migrations: %w", err)
	}
	return out, nil
}

func applyMigration(ctx context.Context, db *sql.DB, version, sqlText string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", version, err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, sqlText); err != nil {
		return fmt.Errorf("apply migration %s: %w", version, err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO workbench_schema_migrations(version) VALUES ($1)`, version); err != nil {
		return fmt.Errorf("record migration %s: %w", version, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", version, err)
	}
	return nil
}

func migrationFiles() ([]string, error) {
	matches, err := fs.Glob(migrationsFS, "migrations/*.sql")
	if err != nil {
		return nil, fmt.Errorf("glob migrations: %w", err)
	}
	files := make([]string, 0, len(matches))
	for _, match := range matches {
		files = append(files, filepath.Base(match))
	}
	sort.Strings(files)
	return files, nil
}

func readMigration(file string) (string, error) {
	data, err := migrationsFS.ReadFile("migrations/" + file)
	if err != nil {
		return "", fmt.Errorf("read migration %s: %w", file, err)
	}
	return string(data), nil
}

func migrationVersion(file string) (string, error) {
	base := filepath.Base(file)
	parts := strings.SplitN(base, "_", 2)
	if len(parts) != 2 || parts[0] == "" || !strings.HasSuffix(base, ".sql") {
		return "", fmt.Errorf("invalid migration filename %q", file)
	}
	return parts[0], nil
}
