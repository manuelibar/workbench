package tools

import "testing"

func TestRegisteredTools(t *testing.T) {
	registry := DefaultRegistry()
	seen := map[string]bool{}
	for _, tool := range registry.Tools() {
		name := tool.FullName()
		if name == "" {
			t.Fatalf("%T has empty full name", tool)
		}
		if seen[name] {
			t.Fatalf("duplicate tool %q", name)
		}
		seen[name] = true
		if tool.Description() == "" {
			t.Fatalf("%s has empty description", name)
		}
	}
	for _, name := range []string{
		"contextualize",
		"artifact.begin",
		"artifact.get",
		"artifact.guidance",
		"artifact.list",
		"artifact.update",
		"artifact.validate",
	} {
		if !seen[name] {
			t.Fatalf("registered tools missing %q", name)
		}
		if _, ok := registry.ByName(name); !ok {
			t.Fatalf("tool lookup missing %q", name)
		}
	}
}

func TestToolGroupComposesArtifactNames(t *testing.T) {
	tool, ok := DefaultRegistry().ByName("artifact.begin")
	if !ok {
		t.Fatal("artifact.begin missing")
	}
	if tool.Name() != "begin" {
		t.Fatalf("short name = %q, want begin", tool.Name())
	}
	if tool.Group() != "artifact" {
		t.Fatalf("group = %q, want artifact", tool.Group())
	}
}
