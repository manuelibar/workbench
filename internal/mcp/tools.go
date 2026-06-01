package mcp

import (
	"context"
	"strings"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/manuelibar/workbench/internal/artifacts"
	"github.com/manuelibar/workbench/internal/errs"
)

type ContextResult struct {
	ContextDocument      string               `json:"context_document"`
	Focus                *string              `json:"focus,omitempty"`
	ArtifactID           *string              `json:"artifact_id,omitempty"`
	Sync                 CapabilitySyncStatus `json:"sync"`
	FallbackCapabilities *CapabilitySurface   `json:"fallback_capabilities,omitempty"`
}

type ArtifactBeginResult struct {
	Artifact artifactPayload `json:"artifact"`
	Context  *ContextResult  `json:"context,omitempty"`
}

type ArtifactBeginRequest struct {
	Type   string `json:"type" jsonschema:"artifact contract type such as rfc, adr, charter, spec, risk, assumption"`
	Title  string `json:"title" jsonschema:"artifact title"`
	Status string `json:"status,omitempty" jsonschema:"artifact status; defaults to draft"`
	Focus  string `json:"focus,omitempty" jsonschema:"optional focus to record in the generated draft"`
	Select bool   `json:"select,omitempty" jsonschema:"select the new artifact in context after creation"`
}

type ArtifactListResult struct {
	Artifacts []artifactSummaryPayload `json:"artifacts"`
}

type ArtifactGetRequest struct {
	ArtifactID string `json:"artifact_id" jsonschema:"artifact id"`
}

type ArtifactUpdateRequest struct {
	ArtifactID   string            `json:"artifact_id,omitempty" jsonschema:"artifact id; defaults to selected artifact"`
	Title        string            `json:"title,omitempty" jsonschema:"new artifact title"`
	Status       string            `json:"status,omitempty" jsonschema:"new artifact status"`
	SetSections  map[string]string `json:"set_sections,omitempty" jsonschema:"replace section bodies by section key"`
	ClearSection []string          `json:"clear_section,omitempty" jsonschema:"section keys to clear"`
}

type ArtifactGuidanceRequest struct {
	ArtifactID string `json:"artifact_id,omitempty" jsonschema:"artifact id; defaults to selected artifact"`
	Type       string `json:"type,omitempty" jsonschema:"artifact contract type for guidance when no artifact is selected"`
}

type ArtifactGuidanceResult struct {
	ArtifactID string                  `json:"artifact_id,omitempty"`
	Contract   artifactContractPayload `json:"contract"`
	Next       []string                `json:"next"`
}

type ArtifactValidateRequest struct {
	ArtifactID string `json:"artifact_id,omitempty" jsonschema:"artifact id; defaults to selected artifact"`
}

type artifactSummaryPayload struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Title   string `json:"title"`
	Status  string `json:"status"`
	Created string `json:"created"`
	Updated string `json:"updated"`
	Path    string `json:"path"`
}

type artifactPayload struct {
	artifactSummaryPayload
	Markdown string            `json:"markdown"`
	Sections map[string]string `json:"sections,omitempty"`
}

type artifactValidationPayload struct {
	ArtifactID string   `json:"artifact_id"`
	Valid      bool     `json:"valid"`
	Issues     []string `json:"issues"`
}

type artifactSectionPayload struct {
	Key      string `json:"key"`
	Title    string `json:"title"`
	Prompt   string `json:"prompt,omitempty"`
	Required bool   `json:"required"`
}

type artifactContractPayload struct {
	Type             string                   `json:"type"`
	Title            string                   `json:"title"`
	Purpose          string                   `json:"purpose"`
	RequiredSections []artifactSectionPayload `json:"required_sections"`
	OptionalSections []artifactSectionPayload `json:"optional_sections,omitempty"`
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

func (s *Server) handleArtifactBegin(ctx context.Context, _ *mcpsdk.CallToolRequest, req ArtifactBeginRequest) (*mcpsdk.CallToolResult, ArtifactBeginResult, error) {
	attrs := map[string]any{"tool": "artifact.begin"}
	artifact, err := s.artifacts.Begin(artifacts.BeginRequest{
		Type:   req.Type,
		Title:  req.Title,
		Status: req.Status,
		Focus:  req.Focus,
	})
	if err != nil {
		return nil, ArtifactBeginResult{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	attrs["artifact_id"] = artifact.ID
	result := ArtifactBeginResult{Artifact: artifactPayloadFrom(artifact)}
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
	summaries, err := s.artifacts.List()
	if err != nil {
		return nil, ArtifactListResult{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	return nil, ArtifactListResult{Artifacts: artifactSummaryPayloadsFrom(summaries)}, nil
}

func (s *Server) handleArtifactGet(_ context.Context, _ *mcpsdk.CallToolRequest, req ArtifactGetRequest) (*mcpsdk.CallToolResult, artifactPayload, error) {
	attrs := map[string]any{
		"tool":        "artifact.get",
		"artifact_id": req.ArtifactID,
	}
	artifact, err := s.artifacts.Get(req.ArtifactID)
	if err != nil {
		return nil, artifactPayload{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	return nil, artifactPayloadFrom(artifact), nil
}

func (s *Server) handleArtifactUpdate(_ context.Context, _ *mcpsdk.CallToolRequest, req ArtifactUpdateRequest) (*mcpsdk.CallToolResult, artifactPayload, error) {
	attrs := map[string]any{"tool": "artifact.update"}
	id, err := s.resolveArtifactID(req.ArtifactID)
	if err != nil {
		return nil, artifactPayload{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	attrs["artifact_id"] = id
	artifact, err := s.artifacts.Update(id, artifacts.UpdateRequest{
		Title:        req.Title,
		Status:       req.Status,
		SetSections:  req.SetSections,
		ClearSection: req.ClearSection,
	})
	if err != nil {
		return nil, artifactPayload{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	if strings.TrimSpace(req.Title) != "" || strings.TrimSpace(req.Status) != "" {
		s.refreshSelectedArtifactResource(artifact.Summary)
	}
	return nil, artifactPayloadFrom(artifact), nil
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
	return nil, ArtifactGuidanceResult{ArtifactID: id, Contract: artifactContractPayloadFrom(contract), Next: next}, nil
}

func (s *Server) handleArtifactValidate(_ context.Context, _ *mcpsdk.CallToolRequest, req ArtifactValidateRequest) (*mcpsdk.CallToolResult, artifactValidationPayload, error) {
	attrs := map[string]any{"tool": "artifact.validate"}
	id, err := s.resolveArtifactID(req.ArtifactID)
	if err != nil {
		return nil, artifactValidationPayload{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	attrs["artifact_id"] = id
	validation, err := s.artifacts.Validate(id)
	if err != nil {
		return nil, artifactValidationPayload{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	return nil, artifactValidationPayloadFrom(validation), nil
}

func artifactSummaryPayloadsFrom(summaries []artifacts.Summary) []artifactSummaryPayload {
	out := make([]artifactSummaryPayload, 0, len(summaries))
	for _, summary := range summaries {
		out = append(out, artifactSummaryPayloadFrom(summary))
	}
	return out
}

func artifactSummaryPayloadFrom(summary artifacts.Summary) artifactSummaryPayload {
	return artifactSummaryPayload{
		ID:      summary.ID,
		Type:    summary.Type,
		Title:   summary.Title,
		Status:  summary.Status,
		Created: summary.Created,
		Updated: summary.Updated,
		Path:    summary.Path,
	}
}

func artifactPayloadFrom(artifact artifacts.Artifact) artifactPayload {
	return artifactPayload{
		artifactSummaryPayload: artifactSummaryPayloadFrom(artifact.Summary),
		Markdown:               artifact.Markdown,
		Sections:               cloneStringMap(artifact.Sections),
	}
}

func artifactValidationPayloadFrom(validation artifacts.Validation) artifactValidationPayload {
	return artifactValidationPayload{
		ArtifactID: validation.ArtifactID,
		Valid:      validation.Valid,
		Issues:     append([]string(nil), validation.Issues...),
	}
}

func artifactContractPayloadFrom(contract artifacts.Contract) artifactContractPayload {
	return artifactContractPayload{
		Type:             contract.Type,
		Title:            contract.Title,
		Purpose:          contract.Purpose,
		RequiredSections: artifactSectionPayloadsFrom(contract.RequiredSections),
		OptionalSections: artifactSectionPayloadsFrom(contract.OptionalSections),
	}
}

func artifactSectionPayloadsFrom(sections []artifacts.SectionSpec) []artifactSectionPayload {
	out := make([]artifactSectionPayload, 0, len(sections))
	for _, section := range sections {
		out = append(out, artifactSectionPayload{
			Key:      section.Key,
			Title:    section.Title,
			Prompt:   section.Prompt,
			Required: section.Required,
		})
	}
	return out
}

func cloneStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
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
	var selected *artifacts.Summary
	if state.ArtifactID != nil && *state.ArtifactID != "" {
		if artifact, err := s.artifacts.Get(*state.ArtifactID); err == nil {
			selected = &artifact.Summary
		}
	}
	if selected != nil {
		decorateSelectedArtifactSurface(&plan.Active, *selected)
		decorateSelectedArtifactSurface(&plan.All, *selected)
	}
	doc := contextDocument(state, plan, selected)
	result := ContextResult{
		ContextDocument: doc,
		Focus:           cloneStringPtr(state.Focus),
		ArtifactID:      cloneStringPtr(state.ArtifactID),
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
