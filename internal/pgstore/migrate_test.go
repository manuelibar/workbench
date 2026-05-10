package pgstore

// White-box tests for the embedded migration manifest. These don't need a
// live Postgres — they only validate the static set of bundled SQL files.

import (
	"strings"
	"testing"
)

func TestLoadMigrations_BundledManifest(t *testing.T) {
	t.Parallel()
	ms, err := loadMigrations()
	if err != nil {
		t.Fatalf("loadMigrations: %v", err)
	}
	if len(ms) == 0 {
		t.Fatal("no migrations bundled")
	}
	// Versions must be 1..N, ascending, no gaps.
	for i, m := range ms {
		if int64(i+1) != m.Version {
			t.Fatalf("migration %d has version %d, want %d", i, m.Version, i+1)
		}
		if m.Name == "" {
			t.Errorf("migration %04d has empty name", m.Version)
		}
		if strings.TrimSpace(m.SQL) == "" {
			t.Errorf("migration %04d_%s has empty SQL body", m.Version, m.Name)
		}
	}
}

func TestLoadMigrations_KnownVersions(t *testing.T) {
	t.Parallel()
	ms, err := loadMigrations()
	if err != nil {
		t.Fatalf("loadMigrations: %v", err)
	}
	want := map[int64]string{
		1: "init",
		2: "pgvector",
	}
	got := make(map[int64]string, len(ms))
	for _, m := range ms {
		got[m.Version] = m.Name
	}
	for v, name := range want {
		if got[v] != name {
			t.Errorf("migration %d = %q, want %q", v, got[v], name)
		}
	}
}
