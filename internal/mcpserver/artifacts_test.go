package mcpserver

import (
	"strings"
	"testing"
	"time"
)

func TestArtifactContractsGenerateValidateAndUpdateMarkdown(t *testing.T) {
	registry := NewContractRegistry()
	if _, ok := registry.Get("rfc"); !ok {
		t.Fatal("rfc contract missing")
	}
	if _, ok := registry.Get("charter"); !ok {
		t.Fatal("charter contract missing")
	}

	store, err := NewArtifactStore(t.TempDir(), registry)
	if err != nil {
		t.Fatal(err)
	}
	store.now = func() time.Time { return time.Date(2026, 5, 30, 12, 0, 0, 0, time.UTC) }
	artifact, err := store.Begin(BeginArtifactRequest{
		Type:  "rfc",
		Title: "Capability sync",
		Focus: "make relist deterministic",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(artifact.Markdown, "id: "+yamlString(artifact.ID)) {
		t.Fatalf("markdown missing id frontmatter:\n%s", artifact.Markdown)
	}
	if !strings.Contains(artifact.Markdown, "## Summary") {
		t.Fatalf("markdown missing contract section:\n%s", artifact.Markdown)
	}

	validation, err := store.Validate(artifact.ID)
	if err != nil {
		t.Fatal(err)
	}
	if validation.Valid {
		t.Fatalf("blank required sections validated: %+v", validation)
	}

	updated, err := store.Update(artifact.ID, UpdateArtifactRequest{
		SetSections: map[string]string{
			"summary":        "Wait until changed capability lists are observed.",
			"problem":        "Agents otherwise act on a stale tool surface.",
			"proposal":       "Track changed categories and observed list calls.",
			"tradeoffs":      "A timeout returns the full fallback capability index.",
			"rollout":        "Default to five seconds, override in tests.",
			"open_questions": "None for the kernel.",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(updated.Markdown, "Wait until changed capability lists are observed.") {
		t.Fatalf("section update missing from markdown:\n%s", updated.Markdown)
	}
	validation, err = store.Validate(artifact.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !validation.Valid {
		t.Fatalf("updated artifact did not validate: %+v", validation)
	}
}
