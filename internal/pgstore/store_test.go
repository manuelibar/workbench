package pgstore_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/manuelibar/workbench/internal/domain"
	"github.com/manuelibar/workbench/internal/pgstore"
)

// integrationDSN returns the DSN used by integration tests; falls back to the
// docker-compose default if WORKBENCH_DB_URL is unset.
func integrationDSN() string {
	if dsn := os.Getenv("WORKBENCH_DB_URL"); dsn != "" {
		return dsn
	}
	return "postgres://workbench:workbench@127.0.0.1:5432/workbench?sslmode=disable"
}

// openForIntegration is the boilerplate skip-and-connect dance used by every
// integration test in this file.
func openForIntegration(t *testing.T) (*pgstore.Store, context.Context, context.CancelFunc) {
	t.Helper()
	if testing.Short() {
		t.Skip("integration test (requires running Postgres on " + integrationDSN() + "); pass -short to skip")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	s, err := pgstore.Open(ctx, integrationDSN())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(s.Close)

	if err := s.Migrate(ctx); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	return s, ctx, cancel
}

func TestStore_Migrate_Idempotent(t *testing.T) {
	s, ctx, _ := openForIntegration(t)
	// Migrate was already called once via openForIntegration; calling again
	// must be a no-op.
	if err := s.Migrate(ctx); err != nil {
		t.Fatalf("second Migrate: %v", err)
	}
}

func TestStore_EnsureSingletonUser(t *testing.T) {
	s, ctx, _ := openForIntegration(t)
	u1, err := s.EnsureSingletonUser(ctx, "test-user")
	if err != nil {
		t.Fatalf("EnsureSingletonUser: %v", err)
	}
	if u1.ID == uuid.Nil {
		t.Fatal("user id is zero")
	}
	u2, err := s.EnsureSingletonUser(ctx, "different-name-should-be-ignored")
	if err != nil {
		t.Fatalf("EnsureSingletonUser (second call): %v", err)
	}
	if u1.ID != u2.ID {
		t.Errorf("expected singleton; got id1=%s id2=%s", u1.ID, u2.ID)
	}
}

func TestStore_WorkSessionLifecycle(t *testing.T) {
	s, ctx, _ := openForIntegration(t)
	u, err := s.EnsureSingletonUser(ctx, "")
	if err != nil {
		t.Fatalf("EnsureSingletonUser: %v", err)
	}

	w1, err := s.EnsureOpenWorkSession(ctx, u.ID, "")
	if err != nil {
		t.Fatalf("EnsureOpenWorkSession: %v", err)
	}
	if w1.UserID != u.ID {
		t.Errorf("user id mismatch")
	}
	if !w1.IsOpen() {
		t.Error("expected open")
	}
	if !w1.Selection.IsEmpty() {
		t.Errorf("expected empty selection, got %+v", w1.Selection)
	}

	w2, err := s.EnsureOpenWorkSession(ctx, u.ID, "")
	if err != nil {
		t.Fatalf("EnsureOpenWorkSession (second call): %v", err)
	}
	if w1.ID != w2.ID {
		t.Errorf("expected singleton open session: %s vs %s", w1.ID, w2.ID)
	}

	nsID := uuid.Must(uuid.NewV7())
	if err := s.UpdateSelection(ctx, u.ID, domain.Selection{NamespaceID: &nsID}); err != nil {
		t.Fatalf("UpdateSelection: %v", err)
	}
	w3, err := s.EnsureOpenWorkSession(ctx, u.ID, "")
	if err != nil {
		t.Fatalf("EnsureOpenWorkSession (after update): %v", err)
	}
	if w3.Selection.NamespaceID == nil || *w3.Selection.NamespaceID != nsID {
		t.Errorf("selection not persisted: %+v", w3.Selection)
	}

	if err := s.CloseWorkSession(ctx, u.ID); err != nil {
		t.Fatalf("CloseWorkSession: %v", err)
	}
	// After close, EnsureOpenWorkSession should create a brand-new session.
	w4, err := s.EnsureOpenWorkSession(ctx, u.ID, "")
	if err != nil {
		t.Fatalf("EnsureOpenWorkSession (post-close): %v", err)
	}
	if w4.ID == w1.ID {
		t.Errorf("expected new session id after close")
	}
}
