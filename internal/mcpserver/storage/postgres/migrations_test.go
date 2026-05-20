package postgres

import (
	"strings"
	"testing"
)

func TestEmbeddedInitialMigrationDefinesDurableWorkbenchState(t *testing.T) {
	files, err := migrationFiles()
	if err != nil {
		t.Fatalf("migrationFiles: %v", err)
	}
	if len(files) != 1 || files[0] != "0001_initial.sql" {
		t.Fatalf("expected one initial migration, got %v", files)
	}

	sqlText, err := readMigration(files[0])
	if err != nil {
		t.Fatalf("readMigration: %v", err)
	}
	for _, want := range []string{
		"CREATE TABLE IF NOT EXISTS namespaces",
		"CREATE TABLE IF NOT EXISTS namespace_edges",
		"CREATE TABLE IF NOT EXISTS projects",
		"CREATE TABLE IF NOT EXISTS roles",
		"CREATE TABLE IF NOT EXISTS prompt_templates",
		"CREATE TABLE IF NOT EXISTS tasks",
		"CREATE TABLE IF NOT EXISTS knowledge_items",
		"CREATE TABLE IF NOT EXISTS sessions",
		"slug, display_name, description, kind",
		"'default', 'Default', 'Default Workbench namespace', 'namespace'",
	} {
		if !strings.Contains(sqlText, want) {
			t.Fatalf("initial migration missing %q\n%s", want, sqlText)
		}
	}
}

func TestMigrationVersionFromFilename(t *testing.T) {
	version, err := migrationVersion("0001_initial.sql")
	if err != nil {
		t.Fatalf("migrationVersion: %v", err)
	}
	if version != "0001" {
		t.Fatalf("expected version 0001, got %q", version)
	}
	if _, err := migrationVersion("initial.sql"); err == nil {
		t.Fatalf("expected invalid migration filename to fail")
	}
}
