package pgstore

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Sentinel errors returned by Store methods. Callers can match with [errors.Is].
var (
	// ErrNotFound is returned when a row is expected but not present.
	ErrNotFound = errors.New("pgstore: not found")
	// ErrConflict is returned on uniqueness or constraint violations that the
	// caller is expected to handle (e.g. duplicate idempotency key with a
	// different payload).
	ErrConflict = errors.New("pgstore: conflict")
)

// Store is a typed wrapper around a [pgxpool.Pool] exposing the workbench
// persistence API. Construct with [Open]; close with [Store.Close].
type Store struct {
	pool *pgxpool.Pool
}

// Open dials the supplied DSN, validates connectivity with a Ping, and
// returns a ready-to-use [Store]. The caller must defer [Store.Close].
func Open(ctx context.Context, dsn string) (*Store, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("pgstore: parse dsn: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pgstore: ping: %w", err)
	}
	return &Store{pool: pool}, nil
}

// Close releases all connections held by the underlying pool.
func (s *Store) Close() {
	if s == nil || s.pool == nil {
		return
	}
	s.pool.Close()
}

// Ping verifies the database is reachable.
func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

// Pool returns the underlying [pgxpool.Pool] for callers that need direct
// access (e.g. for advanced query types not yet exposed on Store).
func (s *Store) Pool() *pgxpool.Pool { return s.pool }

// WithTx runs fn inside a serializable-by-default transaction. The
// transaction commits on a nil return; any non-nil error rolls back.
func (s *Store) WithTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("pgstore: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
