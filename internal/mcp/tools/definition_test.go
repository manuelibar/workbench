package tools

import (
	"encoding/json"
	"testing"
)

func TestAllDefinitionsAreUsable(t *testing.T) {
	seenNames := map[string]bool{}
	for _, def := range All() {
		if def.Name() == "" {
			t.Fatalf("%T has empty name", def)
		}
		if seenNames[def.Name()] {
			t.Fatalf("duplicate tool name %q", def.Name())
		}
		seenNames[def.Name()] = true

		if def.Description() == nil || *def.Description() == "" {
			t.Fatalf("%s has empty description", def.Name())
		}
		if def.Group() == "" {
			t.Fatalf("%s has empty group", def.Name())
		}
		if def.Visibility() == "" {
			t.Fatalf("%s has empty visibility", def.Name())
		}
		if def.InputSchema() == nil {
			t.Fatalf("%s has nil input schema", def.Name())
		}
		if _, err := json.Marshal(def.InputSchema()); err != nil {
			t.Fatalf("%s input schema does not marshal: %v", def.Name(), err)
		}
	}
}

func TestByName(t *testing.T) {
	def, ok := ByName(ContextName)
	if !ok {
		t.Fatal("context tool missing")
	}
	if Key(def) != ContextName {
		t.Fatalf("context key = %q", Key(def))
	}
	if _, ok := ByName("missing"); ok {
		t.Fatal("missing tool found")
	}
}
