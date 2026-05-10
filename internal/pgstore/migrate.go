package pgstore

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// migration is one parsed migration file: a numeric version, a human name,
// and the raw SQL body.
type migration struct {
	Version int64
	Name    string
	SQL     string
}

// migrationLockKey is the Postgres advisory-lock key acquired around the
// migrate path. The value is arbitrary but must stay stable across releases
// so concurrent processes serialise consistently.
const migrationLockKey int64 = 0x776B626D6967726E // "wkbmigrn" big-endian

// Migrate applies all pending migrations to the connected database in
// numeric order, recording successful applications in the schema_migrations
// table. Each migration runs in its own transaction.
//
// Migrate is safe to call repeatedly and concurrently: it acquires a session
// advisory lock so two callers (multiple workbench-mcp processes, or the
// test runner using parallel packages) serialise on the same lock key.
// Already-applied versions are skipped.
func (s *Store) Migrate(ctx context.Context) error {
	ms, err := loadMigrations()
	if err != nil {
		return fmt.Errorf("migrate: load: %w", err)
	}

	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("migrate: acquire conn: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, `SELECT pg_advisory_lock($1)`, migrationLockKey); err != nil {
		return fmt.Errorf("migrate: acquire advisory lock: %w", err)
	}
	defer func() {
		// Use a fresh background context so the unlock fires even if ctx was
		// cancelled mid-migration.
		unlockCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = conn.Exec(unlockCtx, `SELECT pg_advisory_unlock($1)`, migrationLockKey)
	}()

	if _, err := conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version    BIGINT      PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
	)`); err != nil {
		return fmt.Errorf("migrate: create schema_migrations: %w", err)
	}

	rows, err := conn.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return fmt.Errorf("migrate: read state: %w", err)
	}
	applied := map[int64]struct{}{}
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			rows.Close()
			return fmt.Errorf("migrate: scan version: %w", err)
		}
		applied[v] = struct{}{}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return fmt.Errorf("migrate: iterate state: %w", err)
	}

	for _, m := range ms {
		if _, ok := applied[m.Version]; ok {
			continue
		}
		if err := applyMigrationOnConn(ctx, conn.Conn(), m); err != nil {
			return fmt.Errorf("migrate: apply %04d_%s: %w", m.Version, m.Name, err)
		}
	}
	return nil
}

// applyMigrationOnConn runs one migration inside its own transaction on the
// supplied connection (so migrations stay on the connection that holds the
// advisory lock).
func applyMigrationOnConn(ctx context.Context, conn *pgx.Conn, m migration) error {
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, m.SQL); err != nil {
		return fmt.Errorf("exec sql: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, m.Version); err != nil {
		return fmt.Errorf("record version: %w", err)
	}
	return tx.Commit(ctx)
}

// loadMigrations reads every NNNN_name.sql file under migrations/, parses the
// version prefix, sorts ascending, and verifies versions are 1, 2, 3, ...
// (no gaps, no duplicates).
func loadMigrations() ([]migration, error) {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}
	ms := make([]migration, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		name := e.Name()
		stem := strings.TrimSuffix(name, ".sql")
		parts := strings.SplitN(stem, "_", 2)
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid migration filename %q (want NNNN_name.sql)", name)
		}
		v, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse version in %q: %w", name, err)
		}
		body, err := fs.ReadFile(migrationsFS, "migrations/"+name)
		if err != nil {
			return nil, fmt.Errorf("read %q: %w", name, err)
		}
		ms = append(ms, migration{Version: v, Name: parts[1], SQL: string(body)})
	}
	sort.Slice(ms, func(i, j int) bool { return ms[i].Version < ms[j].Version })
	for i, m := range ms {
		if int64(i+1) != m.Version {
			return nil, fmt.Errorf("non-monotonic migrations: expected version %d, got %d (%s)", i+1, m.Version, m.Name)
		}
	}
	return ms, nil
}
