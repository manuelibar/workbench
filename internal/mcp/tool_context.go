package mcp

import (
	"context"
	"strings"

	"github.com/manuelibar/workbench/internal/errs"
)

type contextTool struct{}

func init() {
	registerTool[map[string]any, ContextResult](contextTool{})
}

func (contextTool) Name() string {
	return "context"
}

func (contextTool) Group() string {
	return ""
}

func (contextTool) Description() string {
	return "Read or patch focus/artifact context. Optional inputs: omit focus or artifact_id to preserve it, set a string to update it, or set null to clear it."
}

func (contextTool) Handle(ctx context.Context, s *Server, args map[string]any) (ContextResult, error) {
	attrs := map[string]any{"tool": "context"}
	patch, err := ParseContextPatch(args)
	if err != nil {
		return ContextResult{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	if patch.ArtifactID.Present && !patch.ArtifactID.Null && strings.TrimSpace(patch.ArtifactID.Value) != "" {
		id := strings.TrimSpace(patch.ArtifactID.Value)
		attrs["artifact_id"] = id
		if err := s.artifacts.CheckExists(id); err != nil {
			return ContextResult{}, errs.Decorate(err, errs.WithAttrs(attrs))
		}
	}
	state := s.context.Apply(patch)
	result, err := s.contextResult(ctx, state)
	if err != nil {
		return ContextResult{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	return result, nil
}
