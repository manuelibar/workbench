package mcp

import (
	"context"
	"strings"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/manuelibar/workbench/internal/errs"
)

type ContextResult struct {
	ContextDocument         string               `json:"context_document"`
	Focus                   *string              `json:"focus,omitempty"`
	ArtifactID              *string              `json:"artifact_id,omitempty"`
	CapabilityIndex         CapabilityIndex      `json:"capability_index"`
	Sync                    CapabilitySyncStatus `json:"sync"`
	FallbackCapabilityIndex *CapabilityIndex     `json:"fallback_capability_index,omitempty"`
}

type ArtifactBeginResult struct {
	Artifact Artifact       `json:"artifact"`
	Context  *ContextResult `json:"context,omitempty"`
}

type ArtifactListResult struct {
	Artifacts []ArtifactSummary `json:"artifacts"`
}

type ArtifactGetRequest struct {
	ArtifactID string `json:"artifact_id" jsonschema:"artifact id"`
}

type ArtifactGuidanceRequest struct {
	ArtifactID string `json:"artifact_id,omitempty" jsonschema:"artifact id; defaults to selected artifact"`
	Type       string `json:"type,omitempty" jsonschema:"artifact contract type for guidance when no artifact is selected"`
}

type ArtifactGuidanceResult struct {
	ArtifactID string           `json:"artifact_id,omitempty"`
	Contract   ArtifactContract `json:"contract"`
	Next       []string         `json:"next"`
}

type ArtifactValidateRequest struct {
	ArtifactID string `json:"artifact_id,omitempty" jsonschema:"artifact id; defaults to selected artifact"`
}

func (s *Server) handleContext(ctx context.Context, _ *mcpsdk.CallToolRequest, args map[string]any) (*mcpsdk.CallToolResult, ContextResult, error) {
	attrs := map[string]any{"tool": "context"}
	patch, err := ParseContextPatch(args)
	if err != nil {
		return nil, ContextResult{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	if patch.ArtifactID.Present && !patch.ArtifactID.Null && strings.TrimSpace(patch.ArtifactID.Value) != "" {
		id := strings.TrimSpace(patch.ArtifactID.Value)
		attrs["artifact_id"] = id
		if err := s.artifacts.CheckExists(id); err != nil {
			return nil, ContextResult{}, errs.Decorate(err, errs.WithAttrs(attrs))
		}
	}
	state := s.context.Apply(patch)
	result, err := s.contextResult(ctx, state)
	if err != nil {
		return nil, ContextResult{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	return nil, result, nil
}

func (s *Server) handleArtifactBegin(ctx context.Context, _ *mcpsdk.CallToolRequest, req BeginArtifactRequest) (*mcpsdk.CallToolResult, ArtifactBeginResult, error) {
	attrs := map[string]any{"tool": "artifact.begin"}
	artifact, err := s.artifacts.Begin(req)
	if err != nil {
		return nil, ArtifactBeginResult{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	attrs["artifact_id"] = artifact.ID
	result := ArtifactBeginResult{Artifact: artifact}
	if req.Select {
		state := s.context.Apply(ContextPatch{
			ArtifactID: PatchString{Present: true, Value: artifact.ID},
			Focus:      PatchString{Present: strings.TrimSpace(req.Focus) != "", Value: req.Focus},
		})
		contextResult, err := s.contextResult(ctx, state)
		if err != nil {
			return nil, ArtifactBeginResult{}, errs.Decorate(err, errs.WithAttrs(attrs))
		}
		result.Context = &contextResult
	}
	return nil, result, nil
}

func (s *Server) handleArtifactList(context.Context, *mcpsdk.CallToolRequest, map[string]any) (*mcpsdk.CallToolResult, ArtifactListResult, error) {
	attrs := map[string]any{"tool": "artifact.list"}
	artifacts, err := s.artifacts.List()
	if err != nil {
		return nil, ArtifactListResult{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	return nil, ArtifactListResult{Artifacts: artifacts}, nil
}

func (s *Server) handleArtifactGet(_ context.Context, _ *mcpsdk.CallToolRequest, req ArtifactGetRequest) (*mcpsdk.CallToolResult, Artifact, error) {
	attrs := map[string]any{
		"tool":        "artifact.get",
		"artifact_id": req.ArtifactID,
	}
	artifact, err := s.artifacts.Get(req.ArtifactID)
	if err != nil {
		return nil, Artifact{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	return nil, artifact, nil
}

func (s *Server) handleArtifactUpdate(_ context.Context, _ *mcpsdk.CallToolRequest, req UpdateArtifactRequest) (*mcpsdk.CallToolResult, Artifact, error) {
	attrs := map[string]any{"tool": "artifact.update"}
	id, err := s.resolveArtifactID(req.ArtifactID)
	if err != nil {
		return nil, Artifact{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	attrs["artifact_id"] = id
	artifact, err := s.artifacts.Update(id, req)
	if err != nil {
		return nil, Artifact{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	return nil, artifact, nil
}

func (s *Server) handleArtifactGuidance(_ context.Context, _ *mcpsdk.CallToolRequest, req ArtifactGuidanceRequest) (*mcpsdk.CallToolResult, ArtifactGuidanceResult, error) {
	attrs := map[string]any{"tool": "artifact.guidance"}
	id := strings.TrimSpace(req.ArtifactID)
	if id != "" {
		attrs["artifact_id"] = id
	}
	if id == "" {
		if req.Type == "" {
			var err error
			id, err = s.resolveArtifactID("")
			if err != nil {
				return nil, ArtifactGuidanceResult{}, errs.Decorate(err, errs.WithAttrs(attrs))
			}
			attrs["artifact_id"] = id
		}
	}
	if req.Type != "" {
		attrs["artifact_type"] = req.Type
	}
	contract, next, err := s.artifacts.Guidance(id, req.Type)
	if err != nil {
		return nil, ArtifactGuidanceResult{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	return nil, ArtifactGuidanceResult{ArtifactID: id, Contract: contract, Next: next}, nil
}

func (s *Server) handleArtifactValidate(_ context.Context, _ *mcpsdk.CallToolRequest, req ArtifactValidateRequest) (*mcpsdk.CallToolResult, ArtifactValidation, error) {
	attrs := map[string]any{"tool": "artifact.validate"}
	id, err := s.resolveArtifactID(req.ArtifactID)
	if err != nil {
		return nil, ArtifactValidation{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	attrs["artifact_id"] = id
	validation, err := s.artifacts.Validate(id)
	if err != nil {
		return nil, ArtifactValidation{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	return nil, validation, nil
}

func (s *Server) contextResult(ctx context.Context, state ContextState) (ContextResult, error) {
	attrs := map[string]any{"operation": "context.plan"}
	plan, err := s.plan(ctx, state)
	if err != nil {
		return ContextResult{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	changed := s.diffPlan(plan)
	tracker := s.sync.Begin(changed)
	s.applyPlan(plan)
	syncStatus := s.sync.Wait(ctx, tracker)
	var selected *ArtifactSummary
	if state.ArtifactID != nil && *state.ArtifactID != "" {
		if artifact, err := s.artifacts.Get(*state.ArtifactID); err == nil {
			selected = &artifact.ArtifactSummary
		}
	}
	doc := contextDocument(state, plan, selected)
	result := ContextResult{
		ContextDocument: doc,
		Focus:           cloneStringPtr(state.Focus),
		ArtifactID:      cloneStringPtr(state.ArtifactID),
		CapabilityIndex: plan.Index,
		Sync:            syncStatus,
	}
	if syncStatus.TimedOut {
		result.FallbackCapabilityIndex = &plan.All
	}
	return result, nil
}

func (s *Server) diffPlan(plan CapabilityPlan) []string {
	s.surfaceMu.Lock()
	defer s.surfaceMu.Unlock()
	var changed []string
	if !sameStringSet(s.active.tools, toolsFromIndex(plan.Index)) {
		changed = append(changed, "tools")
	}
	if !sameStringSet(s.active.resources, resourcesFromIndex(plan.Index)) {
		changed = append(changed, "resources")
	}
	if !sameStringSet(s.active.resourceTemplates, templatesFromIndex(plan.Index)) {
		changed = append(changed, "resource_templates")
	}
	return changed
}

func (s *Server) resolveArtifactID(id string) (string, error) {
	attrs := map[string]any{"operation": "artifact.resolve"}
	id = strings.TrimSpace(id)
	if id != "" {
		attrs["artifact_id"] = id
		if err := s.artifacts.CheckExists(id); err != nil {
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
	if err := s.artifacts.CheckExists(*state.ArtifactID); err != nil {
		return "", errs.Decorate(err, errs.WithAttrs(attrs))
	}
	return *state.ArtifactID, nil
}

func cloneStringPtr(in *string) *string {
	if in == nil {
		return nil
	}
	return ptr(*in)
}
