package postgres

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// DB wraps a Postgres database handle used by Workbench infrastructure.
type DB struct {
	SQL *sql.DB
}

// Open connects to Postgres using pgx's database/sql driver and verifies the
// connection. Call Close when the server shuts down.
func Open(ctx context.Context, databaseURL string) (*DB, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("database URL is required")
	}
	sqldb, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	if err := sqldb.PingContext(ctx); err != nil {
		sqldb.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &DB{SQL: sqldb}, nil
}

// Close releases the database handle.
func (db *DB) Close() error {
	if db == nil || db.SQL == nil {
		return nil
	}
	return db.SQL.Close()
}
