package mcp

import (
	"context"
	"os"
	"strings"

	"github.com/manuelibar/workbench/internal/artifacts"
	"github.com/manuelibar/workbench/internal/errs"
	"github.com/manuelibar/workbench/internal/mcp/tools"
)

func (s *Server) ApplyScopePatch(ctx context.Context, args map[string]any) (tools.ContextualizeResult, error) {
	attrs := map[string]any{"operation": "scope.apply"}
	patch, err := parseScopePatch(args)
	if err != nil {
		return tools.ContextualizeResult{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	if patch.ArtifactID.Present && !patch.ArtifactID.Null && strings.TrimSpace(patch.ArtifactID.Value) != "" {
		id := strings.TrimSpace(patch.ArtifactID.Value)
		attrs["artifact_id"] = id
		path, err := s.materializeScopedArtifact(ctx, id)
		if err != nil {
			return tools.ContextualizeResult{}, errs.Decorate(err, errs.WithAttrs(attrs))
		}
		patch.ArtifactID.Value = id
		patch.LocalArtifactPath = patchString{Present: true, Value: path}
	} else if patch.ArtifactID.Present {
		patch.LocalArtifactPath = patchString{Present: true, Null: true}
	}
	state := s.scope.Apply(patch)
	return s.contextualizeResult(ctx, state)
}

func (s *Server) CreateArtifact(ctx context.Context, req artifacts.CreateRequest) (artifacts.Artifact, error) {
	return s.artifacts.CreateContext(ctx, req)
}

func (s *Server) FindArtifacts(ctx context.Context, req tools.ArtifactFindRequest) ([]artifacts.Summary, error) {
	attrs := map[string]any{"operation": "artifact.find"}
	summaries, err := s.artifacts.ListContext(ctx)
	if err != nil {
		return nil, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	query := strings.ToLower(strings.TrimSpace(req.Query))
	typ := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(req.Type), " ", "_"))
	status := strings.ToLower(strings.TrimSpace(req.Status))
	limit := req.Limit
	var out []artifacts.Summary
	for _, summary := range summaries {
		if typ != "" && strings.ToLower(summary.Type) != typ {
			continue
		}
		if status != "" && strings.ToLower(summary.Status) != status {
			continue
		}
		if query != "" &&
			!strings.Contains(strings.ToLower(summary.ID), query) &&
			!strings.Contains(strings.ToLower(summary.Type), query) &&
			!strings.Contains(strings.ToLower(summary.Title), query) &&
			!strings.Contains(strings.ToLower(summary.Status), query) {
			continue
		}
		out = append(out, summary)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (s *Server) UploadScopedArtifact(ctx context.Context, markdown string) (artifacts.Artifact, error) {
	attrs := map[string]any{"operation": "artifact.upload"}
	id, path, err := s.scopedArtifact()
	if err != nil {
		return artifacts.Artifact{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	if strings.TrimSpace(markdown) == "" {
		raw, err := os.ReadFile(path)
		if err != nil {
			attrs["local_artifact_path"] = path
			return artifacts.Artifact{}, errs.New(
				"Artifact scope required",
				errs.WithSentinel(errs.ErrInvalid),
				errs.WithCode(errCodeArtifactScopeMissing),
				errs.WithSeverity(errs.SeverityWarning),
				errs.WithCause(err),
				errs.WithAttrs(attrs),
				errs.WithRetryable(false),
			)
		}
		markdown = string(raw)
	}
	artifact, err := s.artifacts.UploadMarkdownContext(ctx, id, markdown)
	if err != nil {
		attrs["artifact_id"] = id
		return artifacts.Artifact{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	if strings.TrimSpace(path) != "" {
		_ = os.WriteFile(path, []byte(markdown), 0o644)
	}
	s.surface.RefreshArtifactResource(s, artifact.Summary)
	return artifact, nil
}

func (s *Server) contextualizeResult(ctx context.Context, state scopeState) (tools.ContextualizeResult, error) {
	attrs := map[string]any{"operation": "scope.plan"}
	plan, err := s.plan(ctx, state)
	if err != nil {
		return tools.ContextualizeResult{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	changed := s.surface.ChangedCategories(plan.Active)
	tracker := s.sync.Begin(changed)
	s.surface.Synchronize(s, plan.Active)
	syncStatus := s.sync.Wait(ctx, tracker)
	var scoped *artifacts.Summary
	if state.ArtifactID != nil && *state.ArtifactID != "" {
		if artifact, err := s.artifacts.GetContext(ctx, *state.ArtifactID); err == nil {
			scoped = &artifact.Summary
		}
	}
	if scoped != nil {
		decorateArtifactSurface(&plan.Active, *scoped)
		decorateArtifactSurface(&plan.All, *scoped)
	}
	result := tools.ContextualizeResult{
		ScopeDocument: scopeDocument(state, plan, scoped),
		Focus:         state.Focus,
		ArtifactID:    state.ArtifactID,
		Sync:          syncStatus,
	}
	if syncStatus.TimedOut {
		result.FallbackCapabilities = &plan.All
	}
	return result, nil
}
