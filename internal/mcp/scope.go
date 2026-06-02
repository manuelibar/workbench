package mcp

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/manuelibar/workbench/internal/artifacts"
	"github.com/manuelibar/workbench/internal/errs"
	mcpresources "github.com/manuelibar/workbench/internal/mcp/resources"
)

// scopeState is the in-memory scope state for one stdio Workbench process.
type scopeState struct {
	Focus             *string `json:"focus,omitempty"`
	ArtifactID        *string `json:"artifact_id,omitempty"`
	LocalArtifactPath *string `json:"local_artifact_path,omitempty"`
	Version           uint64  `json:"version"`
}

// scopePatch carries tri-state patch fields:
// omitted preserves, null clears, and string sets.
type scopePatch struct {
	Focus             patchString
	ArtifactID        patchString
	LocalArtifactPath patchString
}

// patchString represents one tri-state string field in a JSON patch.
type patchString struct {
	Present bool
	Null    bool
	Value   string
}

// scopeStore is a concurrency-safe in-memory scope store.
type scopeStore struct {
	mu    sync.RWMutex
	state scopeState
}

func newScopeStore() *scopeStore {
	return &scopeStore{}
}

func (s *scopeStore) Snapshot() scopeState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneScopeState(s.state)
}

func (s *scopeStore) Apply(patch scopePatch) scopeState {
	s.mu.Lock()
	defer s.mu.Unlock()

	changed := false
	if patch.Focus.Present {
		changed = applyPatchString(&s.state.Focus, patch.Focus) || changed
	}
	if patch.ArtifactID.Present {
		changed = applyPatchString(&s.state.ArtifactID, patch.ArtifactID) || changed
	}
	if patch.LocalArtifactPath.Present {
		changed = applyPatchString(&s.state.LocalArtifactPath, patch.LocalArtifactPath) || changed
	}
	if changed {
		s.state.Version++
	}
	return cloneScopeState(s.state)
}

func applyPatchString(dst **string, patch patchString) bool {
	before := ""
	beforeSet := *dst != nil
	if beforeSet {
		before = **dst
	}
	if patch.Null {
		*dst = nil
		return beforeSet
	}
	next := strings.TrimSpace(patch.Value)
	if next == "" {
		*dst = nil
		return beforeSet
	}
	*dst = ptr(next)
	return !beforeSet || before != next
}

func cloneScopeState(in scopeState) scopeState {
	out := scopeState{Version: in.Version}
	if in.Focus != nil {
		out.Focus = ptr(*in.Focus)
	}
	if in.ArtifactID != nil {
		out.ArtifactID = ptr(*in.ArtifactID)
	}
	if in.LocalArtifactPath != nil {
		out.LocalArtifactPath = ptr(*in.LocalArtifactPath)
	}
	return out
}

func ptr[T any](v T) *T {
	return &v
}

func parseScopePatch(raw map[string]any) (scopePatch, error) {
	attrs := map[string]any{"operation": "scope.patch.parse"}
	var patch scopePatch
	for key, value := range raw {
		attrs["field"] = key
		switch key {
		case "focus":
			field, err := parsePatchString("focus", value, attrs)
			if err != nil {
				return scopePatch{}, err
			}
			patch.Focus = field
		case "artifact_id":
			field, err := parsePatchString("artifact_id", value, attrs)
			if err != nil {
				return scopePatch{}, err
			}
			patch.ArtifactID = field
		default:
			attrs["reason"] = "unknown_field"
			return scopePatch{}, errs.New(
				"Scope patch is invalid",
				errs.WithSentinel(errs.ErrInvalid),
				errs.WithCode(errCodeScopePatchInvalid),
				errs.WithSeverity(errs.SeverityWarning),
				errs.WithAttrs(attrs),
				errs.WithRetryable(false),
			)
		}
	}
	return patch, nil
}

func parsePatchString(name string, value any, attrs map[string]any) (patchString, error) {
	if value == nil {
		return patchString{Present: true, Null: true}, nil
	}
	switch v := value.(type) {
	case string:
		return patchString{Present: true, Value: v}, nil
	default:
		attrs["field"] = name
		attrs["reason"] = "wrong_type"
		attrs["type"] = reflect.TypeOf(value).String()
		return patchString{}, errs.New(
			"Scope patch is invalid",
			errs.WithSentinel(errs.ErrInvalid),
			errs.WithCode(errCodeScopePatchInvalid),
			errs.WithSeverity(errs.SeverityWarning),
			errs.WithAttrs(attrs),
			errs.WithRetryable(false),
		)
	}
}

func scopeDocument(state scopeState, _ capabilityPlan, artifact *artifacts.Summary) string {
	var b strings.Builder
	b.WriteString(strings.TrimSpace(mcpresources.ScopeMarkdown()))
	b.WriteString("\n\n")
	b.WriteString("## State\n\n")
	b.WriteString(fmt.Sprintf("- version: %d\n", state.Version))
	if state.Focus != nil && *state.Focus != "" {
		b.WriteString("- focus: " + *state.Focus + "\n")
	} else {
		b.WriteString("- focus: none\n")
	}
	if state.ArtifactID != nil && *state.ArtifactID != "" {
		b.WriteString("- artifact_id: " + *state.ArtifactID + "\n")
	} else {
		b.WriteString("- artifact_id: none\n")
	}
	if artifact != nil {
		b.WriteString("\n## Artifact\n\n")
		b.WriteString("- id: " + artifact.ID + "\n")
		b.WriteString("- title: " + artifact.Title + "\n")
		b.WriteString("- type: " + artifact.Type + "\n")
		b.WriteString("- status: " + artifact.Status + "\n")
		b.WriteString("- resource: `workbench:///artifacts/" + artifact.ID + "`\n")
		if state.LocalArtifactPath != nil && *state.LocalArtifactPath != "" {
			b.WriteString("- local_file: `" + *state.LocalArtifactPath + "`\n")
		}
	}
	return b.String()
}
