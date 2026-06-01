package mcp

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/manuelibar/workbench/internal/errs"
)

// ContextState is the in-memory state for one stdio Workbench process.
type ContextState struct {
	Focus      *string `json:"focus,omitempty"`
	ArtifactID *string `json:"artifact_id,omitempty"`
	Version    uint64  `json:"version"`
}

// ContextPatch carries tri-state patch fields:
// omitted preserves, null clears, and string sets.
type ContextPatch struct {
	Focus      PatchString
	ArtifactID PatchString
}

// PatchString represents one tri-state string field in a JSON patch.
type PatchString struct {
	Present bool
	Null    bool
	Value   string
}

// ContextStore is a concurrency-safe in-memory context store.
type ContextStore struct {
	mu    sync.RWMutex
	state ContextState
}

func NewContextStore() *ContextStore {
	return &ContextStore{}
}

func (s *ContextStore) Snapshot() ContextState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneContextState(s.state)
}

func (s *ContextStore) Apply(patch ContextPatch) ContextState {
	s.mu.Lock()
	defer s.mu.Unlock()

	changed := false
	if patch.Focus.Present {
		changed = applyPatchString(&s.state.Focus, patch.Focus) || changed
	}
	if patch.ArtifactID.Present {
		changed = applyPatchString(&s.state.ArtifactID, patch.ArtifactID) || changed
	}
	if changed {
		s.state.Version++
	}
	return cloneContextState(s.state)
}

func applyPatchString(dst **string, patch PatchString) bool {
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

func cloneContextState(in ContextState) ContextState {
	out := ContextState{Version: in.Version}
	if in.Focus != nil {
		out.Focus = ptr(*in.Focus)
	}
	if in.ArtifactID != nil {
		out.ArtifactID = ptr(*in.ArtifactID)
	}
	return out
}

func ptr[T any](v T) *T {
	return &v
}

func ParseContextPatch(raw map[string]any) (ContextPatch, error) {
	attrs := map[string]any{"operation": "context.patch.parse"}
	var patch ContextPatch
	for key, value := range raw {
		attrs["field"] = key
		switch key {
		case "focus":
			field, err := parsePatchString("focus", value, attrs)
			if err != nil {
				return ContextPatch{}, err
			}
			patch.Focus = field
		case "artifact_id":
			field, err := parsePatchString("artifact_id", value, attrs)
			if err != nil {
				return ContextPatch{}, err
			}
			patch.ArtifactID = field
		default:
			attrs["reason"] = "unknown_field"
			return ContextPatch{}, errs.New(
				"Context patch is invalid",
				errs.WithSentinel(errs.ErrInvalid),
				errs.WithCode(errCodeContextPatchInvalid),
				errs.WithSeverity(errs.SeverityWarning),
				errs.WithAttrs(attrs),
				errs.WithRetryable(false),
			)
		}
	}
	return patch, nil
}

func parsePatchString(name string, value any, attrs map[string]any) (PatchString, error) {
	if value == nil {
		return PatchString{Present: true, Null: true}, nil
	}
	switch v := value.(type) {
	case string:
		return PatchString{Present: true, Value: v}, nil
	default:
		attrs["field"] = name
		attrs["reason"] = "wrong_type"
		attrs["type"] = reflect.TypeOf(value).String()
		return PatchString{}, errs.New(
			"Context patch is invalid",
			errs.WithSentinel(errs.ErrInvalid),
			errs.WithCode(errCodeContextPatchInvalid),
			errs.WithSeverity(errs.SeverityWarning),
			errs.WithAttrs(attrs),
			errs.WithRetryable(false),
		)
	}
}

func contextDocument(state ContextState, plan CapabilityPlan, artifact *ArtifactSummary) string {
	var b strings.Builder
	b.WriteString("# Workbench Context\n\n")
	b.WriteString("This is the exact context document served by `workbench:///context`.\n\n")
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
	b.WriteString("\n## Active Capabilities\n\n")
	b.WriteString("Tools:\n")
	for _, cap := range plan.Index.Tools {
		b.WriteString("- `" + cap.Name + "`: " + cap.Description + "\n")
	}
	if len(plan.Index.Resources) > 0 {
		b.WriteString("\nResources:\n")
		for _, cap := range plan.Index.Resources {
			b.WriteString("- `" + cap.URI + "`: " + cap.Description + "\n")
		}
	}
	if len(plan.Index.ResourceTemplates) > 0 {
		b.WriteString("\nResource templates:\n")
		for _, cap := range plan.Index.ResourceTemplates {
			b.WriteString("- `" + cap.URITemplate + "`: " + cap.Description + "\n")
		}
	}
	if artifact != nil {
		b.WriteString("\n## Selected Artifact\n\n")
		b.WriteString("- title: " + artifact.Title + "\n")
		b.WriteString("- type: " + artifact.Type + "\n")
		b.WriteString("- status: " + artifact.Status + "\n")
		b.WriteString("- resource: `workbench:///artifacts/" + artifact.ID + "`\n")
	}
	return b.String()
}
