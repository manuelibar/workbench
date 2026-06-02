package mcp

import (
	"context"
	"strings"

	"github.com/manuelibar/workbench/internal/artifacts"
	"github.com/manuelibar/workbench/internal/errs"
	"github.com/manuelibar/workbench/internal/mcp/tools"
)

func (s *Server) ApplyContextPatch(ctx context.Context, args map[string]any) (tools.ContextualizeResult, error) {
	attrs := map[string]any{"operation": "context.apply"}
	patch, err := ParseContextPatch(args)
	if err != nil {
		return tools.ContextualizeResult{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	if patch.ArtifactID.Present && !patch.ArtifactID.Null && strings.TrimSpace(patch.ArtifactID.Value) != "" {
		id := strings.TrimSpace(patch.ArtifactID.Value)
		attrs["artifact_id"] = id
		if err := s.artifacts.CheckExistsContext(ctx, id); err != nil {
			return tools.ContextualizeResult{}, errs.Decorate(err, errs.WithAttrs(attrs))
		}
	}
	state := s.context.Apply(patch)
	return s.contextualizeResult(ctx, state)
}

func (s *Server) SelectArtifact(ctx context.Context, artifactID, focus string) (tools.ContextualizeResult, error) {
	state := s.context.Apply(ContextPatch{
		ArtifactID: PatchString{Present: true, Value: artifactID},
		Focus:      PatchString{Present: strings.TrimSpace(focus) != "", Value: focus},
	})
	return s.contextualizeResult(ctx, state)
}

func (s *Server) ResolveArtifactID(ctx context.Context, id string) (string, error) {
	return s.resolveArtifactID(ctx, id)
}

func (s *Server) RefreshSelectedArtifactResource(artifact artifacts.Summary) {
	s.refreshSelectedArtifactResource(artifact)
}

func (s *Server) contextualizeResult(ctx context.Context, state ContextState) (tools.ContextualizeResult, error) {
	attrs := map[string]any{"operation": "context.plan"}
	plan, err := s.plan(ctx, state)
	if err != nil {
		return tools.ContextualizeResult{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	changed := s.diffPlan(plan)
	tracker := s.sync.Begin(changed)
	s.applyPlan(plan)
	syncStatus := s.sync.Wait(ctx, tracker)
	var selected *artifacts.Summary
	if state.ArtifactID != nil && *state.ArtifactID != "" {
		if artifact, err := s.artifacts.GetContext(ctx, *state.ArtifactID); err == nil {
			selected = &artifact.Summary
		}
	}
	if selected != nil {
		decorateSelectedArtifactSurface(&plan.Active, *selected)
		decorateSelectedArtifactSurface(&plan.All, *selected)
	}
	result := tools.ContextualizeResult{
		ContextDocument: contextDocument(state, plan, selected),
		Focus:           state.Focus,
		ArtifactID:      state.ArtifactID,
		Sync:            syncStatus,
	}
	if syncStatus.TimedOut {
		result.FallbackCapabilities = &plan.All
	}
	return result, nil
}

func (s *Server) diffPlan(plan CapabilityPlan) []string {
	s.surfaceMu.Lock()
	defer s.surfaceMu.Unlock()
	var changed []string
	if !sameStringSet(s.active.tools, toolsFromSurface(plan.Active)) {
		changed = append(changed, "tools")
	}
	if !sameStringSet(s.active.resources, resourcesFromSurface(plan.Active)) {
		changed = append(changed, "resources")
	}
	if !sameStringSet(s.active.resourceTemplates, templatesFromSurface(plan.Active)) {
		changed = append(changed, "resource_templates")
	}
	return changed
}

func (s *Server) resolveArtifactID(ctx context.Context, id string) (string, error) {
	attrs := map[string]any{"operation": "artifact.resolve"}
	id = strings.TrimSpace(id)
	if id != "" {
		attrs["artifact_id"] = id
		if err := s.artifacts.CheckExistsContext(ctx, id); err != nil {
			return "", errs.Decorate(err, errs.WithAttrs(attrs))
		}
		return id, nil
	}
	state := s.context.Snapshot()
	if state.ArtifactID == nil || *state.ArtifactID == "" {
		return "", errs.New(
			"Artifact selection required",
			errs.WithSentinel(errs.ErrInvalid),
			errs.WithCode(errCodeArtifactSelectionMissing),
			errs.WithSeverity(errs.SeverityWarning),
			errs.WithAttrs(attrs),
			errs.WithRetryable(false),
		)
	}
	attrs["artifact_id"] = *state.ArtifactID
	attrs["selection"] = true
	if err := s.artifacts.CheckExistsContext(ctx, *state.ArtifactID); err != nil {
		return "", errs.Decorate(err, errs.WithAttrs(attrs))
	}
	return *state.ArtifactID, nil
}
