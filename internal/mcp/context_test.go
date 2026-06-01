package mcp

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/manuelibar/workbench/internal/errs"
)

func TestContextPatchTriState(t *testing.T) {
	store := NewContextStore()

	state := store.Apply(ContextPatch{
		Focus: PatchString{Present: true, Value: "ship kernel"},
	})
	if state.Focus == nil || *state.Focus != "ship kernel" {
		t.Fatalf("focus set = %v, want ship kernel", state.Focus)
	}
	if state.Version != 1 {
		t.Fatalf("version = %d, want 1", state.Version)
	}

	state = store.Apply(ContextPatch{})
	if state.Focus == nil || *state.Focus != "ship kernel" {
		t.Fatalf("omitted focus did not preserve value: %v", state.Focus)
	}
	if state.Version != 1 {
		t.Fatalf("omitted patch changed version to %d", state.Version)
	}

	state = store.Apply(ContextPatch{
		ArtifactID: PatchString{Present: true, Value: "artifact-1"},
	})
	if state.Focus == nil || *state.Focus != "ship kernel" {
		t.Fatalf("artifact patch changed focus: %v", state.Focus)
	}
	if state.ArtifactID == nil || *state.ArtifactID != "artifact-1" {
		t.Fatalf("artifact_id set = %v, want artifact-1", state.ArtifactID)
	}

	state = store.Apply(ContextPatch{
		Focus: PatchString{Present: true, Null: true},
	})
	if state.Focus != nil {
		t.Fatalf("null focus did not clear: %v", *state.Focus)
	}
	if state.ArtifactID == nil || *state.ArtifactID != "artifact-1" {
		t.Fatalf("focus clear changed artifact_id: %v", state.ArtifactID)
	}

	patch, err := ParseContextPatch(map[string]any{"artifact_id": nil})
	if err != nil {
		t.Fatal(err)
	}
	state = store.Apply(patch)
	if state.ArtifactID != nil {
		t.Fatalf("null artifact_id did not clear: %v", *state.ArtifactID)
	}
}

func TestContextPatchRejectsUnknownAndWrongTypes(t *testing.T) {
	if _, err := ParseContextPatch(map[string]any{"namespace_id": "later"}); err == nil {
		t.Fatal("unknown field accepted")
	} else if !errors.Is(err, errs.ErrInvalid) {
		t.Fatalf("unknown field error = %v, want ErrInvalid", err)
	} else if got := errs.CodeOf(err); got != errCodeContextPatchInvalid {
		t.Fatalf("unknown field code = %q", got)
	}
	if _, err := ParseContextPatch(map[string]any{"focus": 42}); err == nil {
		t.Fatal("non-string focus accepted")
	} else if !errors.Is(err, errs.ErrInvalid) {
		t.Fatalf("non-string focus error = %v, want ErrInvalid", err)
	} else if got := errs.CodeOf(err); got != errCodeContextPatchInvalid {
		t.Fatalf("non-string focus code = %q", got)
	}
}

func TestContextStoreConcurrentReadsWrites(t *testing.T) {
	store := NewContextStore()
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				_ = store.Snapshot()
				patch := ContextPatch{
					Focus: PatchString{Present: true, Value: fmt.Sprintf("focus-%d-%d", i, j)},
				}
				if j%3 == 0 {
					patch.ArtifactID = PatchString{Present: true, Value: fmt.Sprintf("artifact-%d", i)}
				}
				if j%5 == 0 {
					patch.Focus = PatchString{Present: true, Null: true}
				}
				_ = store.Apply(patch)
			}
		}()
	}
	wg.Wait()
	_ = store.Snapshot()
}
