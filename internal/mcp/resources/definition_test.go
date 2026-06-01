package resources

import "testing"

func TestResourceDefinitions(t *testing.T) {
	seen := map[string]bool{}
	for _, def := range All() {
		key := Key(def)
		if key == "" {
			t.Fatalf("%T has empty key", def)
		}
		if seen[key] {
			t.Fatalf("duplicate resource key %q", key)
		}
		seen[key] = true
		if def.Name() == "" || def.Title() == "" || def.Description() == "" || def.MIMEType() == "" || def.Group() == "" || def.Visibility() == "" {
			t.Fatalf("%s has incomplete metadata", key)
		}
	}
	for _, def := range Templates() {
		if TemplateKey(def) == "" || def.URITemplate() == "" || def.Name() == "" || def.Description() == "" {
			t.Fatalf("%T has incomplete template metadata", def)
		}
	}
}

func TestArtifactIDFromURI(t *testing.T) {
	if got := ArtifactIDFromURI("workbench:///artifacts/abc-123"); got != "abc-123" {
		t.Fatalf("artifact id = %q", got)
	}
	if got := ArtifactIDFromURI("workbench:///artifacts/a/b"); got != "" {
		t.Fatalf("nested artifact id = %q, want empty", got)
	}
	if got := ArtifactIDFromURI("workbench:///context"); got != "" {
		t.Fatalf("context artifact id = %q, want empty", got)
	}
}

func TestSelectedArtifactResourceUsesArtifactMetadata(t *testing.T) {
	resource := NewSelectedArtifactResource(SelectedArtifact{
		ID:     "abc-123",
		Type:   "rfc",
		Title:  "Capability Sync",
		Status: "draft",
	})
	if got := resource.URI(); got != "workbench:///artifacts/abc-123" {
		t.Fatalf("uri = %q", got)
	}
	if got := resource.Name(); got != "Capability Sync" {
		t.Fatalf("name = %q", got)
	}
	if got := resource.Title(); got != "Capability Sync" {
		t.Fatalf("title = %q", got)
	}
	if got := resource.Description(); got != "Read the selected rfc draft artifact Markdown resource." {
		t.Fatalf("description = %q", got)
	}
}
