package mcpserver

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
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

func (s *Server) handleContext(ctx context.Context, _ *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, ContextResult, error) {
	patch, err := ParseContextPatch(args)
	if err != nil {
		return nil, ContextResult{}, err
	}
	if patch.ArtifactID.Present && !patch.ArtifactID.Null && strings.TrimSpace(patch.ArtifactID.Value) != "" {
		id := strings.TrimSpace(patch.ArtifactID.Value)
		if !s.artifacts.Exists(id) {
			return nil, ContextResult{}, fmt.Errorf("context: artifact_id %q does not exist", id)
		}
	}
	state := s.context.Apply(patch)
	result, err := s.contextResult(ctx, state)
	return nil, result, err
}

func (s *Server) handleArtifactBegin(ctx context.Context, _ *mcp.CallToolRequest, req BeginArtifactRequest) (*mcp.CallToolResult, ArtifactBeginResult, error) {
	artifact, err := s.artifacts.Begin(req)
	if err != nil {
		return nil, ArtifactBeginResult{}, err
	}
	result := ArtifactBeginResult{Artifact: artifact}
	if req.Select {
		state := s.context.Apply(ContextPatch{
			ArtifactID: PatchString{Present: true, Value: artifact.ID},
			Focus:      PatchString{Present: strings.TrimSpace(req.Focus) != "", Value: req.Focus},
		})
		contextResult, err := s.contextResult(ctx, state)
		if err != nil {
			return nil, ArtifactBeginResult{}, err
		}
		result.Context = &contextResult
	}
	return nil, result, nil
}

func (s *Server) handleArtifactList(context.Context, *mcp.CallToolRequest, map[string]any) (*mcp.CallToolResult, ArtifactListResult, error) {
	artifacts, err := s.artifacts.List()
	if err != nil {
		return nil, ArtifactListResult{}, err
	}
	return nil, ArtifactListResult{Artifacts: artifacts}, nil
}

func (s *Server) handleArtifactGet(_ context.Context, _ *mcp.CallToolRequest, req ArtifactGetRequest) (*mcp.CallToolResult, Artifact, error) {
	artifact, err := s.artifacts.Get(req.ArtifactID)
	return nil, artifact, err
}

func (s *Server) handleArtifactUpdate(_ context.Context, _ *mcp.CallToolRequest, req UpdateArtifactRequest) (*mcp.CallToolResult, Artifact, error) {
	id, err := s.resolveArtifactID(req.ArtifactID)
	if err != nil {
		return nil, Artifact{}, err
	}
	artifact, err := s.artifacts.Update(id, req)
	return nil, artifact, err
}

func (s *Server) handleArtifactGuidance(_ context.Context, _ *mcp.CallToolRequest, req ArtifactGuidanceRequest) (*mcp.CallToolResult, ArtifactGuidanceResult, error) {
	id := strings.TrimSpace(req.ArtifactID)
	if id == "" {
		if req.Type == "" {
			var err error
			id, err = s.resolveArtifactID("")
			if err != nil {
				return nil, ArtifactGuidanceResult{}, err
			}
		}
	}
	contract, next, err := s.artifacts.Guidance(id, req.Type)
	if err != nil {
		return nil, ArtifactGuidanceResult{}, err
	}
	return nil, ArtifactGuidanceResult{ArtifactID: id, Contract: contract, Next: next}, nil
}

func (s *Server) handleArtifactValidate(_ context.Context, _ *mcp.CallToolRequest, req ArtifactValidateRequest) (*mcp.CallToolResult, ArtifactValidation, error) {
	id, err := s.resolveArtifactID(req.ArtifactID)
	if err != nil {
		return nil, ArtifactValidation{}, err
	}
	validation, err := s.artifacts.Validate(id)
	return nil, validation, err
}

func (s *Server) contextResult(ctx context.Context, state ContextState) (ContextResult, error) {
	plan, err := s.plan(ctx, state)
	if err != nil {
		return ContextResult{}, err
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
	id = strings.TrimSpace(id)
	if id != "" {
		if !s.artifacts.Exists(id) {
			return "", fmt.Errorf("artifact %q does not exist", id)
		}
		return id, nil
	}
	state := s.context.Snapshot()
	if state.ArtifactID == nil || *state.ArtifactID == "" {
		return "", fmt.Errorf("no artifact selected")
	}
	if !s.artifacts.Exists(*state.ArtifactID) {
		return "", fmt.Errorf("selected artifact %q does not exist", *state.ArtifactID)
	}
	return *state.ArtifactID, nil
}

func cloneStringPtr(in *string) *string {
	if in == nil {
		return nil
	}
	return ptr(*in)
}
