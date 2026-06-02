package tools

import (
	"context"
	"strings"

	"github.com/manuelibar/workbench/internal/artifacts"
	"github.com/manuelibar/workbench/internal/errs"
)

type artifactGuidanceTool struct{}

type ArtifactGuidanceRequest struct {
	ArtifactID string `json:"artifact_id,omitempty" jsonschema:"artifact id; defaults to selected artifact"`
	Type       string `json:"type,omitempty" jsonschema:"artifact contract type for guidance when no artifact is selected"`
}

type ArtifactGuidanceResult struct {
	ArtifactID string                  `json:"artifact_id,omitempty"`
	Contract   artifactContractPayload `json:"contract"`
	Next       []string                `json:"next"`
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

func init() {
	defaultRegistry.Register(typedTool[ArtifactGuidanceRequest, ArtifactGuidanceResult]{impl: artifactGuidanceTool{}})
}

func (artifactGuidanceTool) Name() string {
	return "guidance"
}

func (artifactGuidanceTool) Group() string {
	return "artifact"
}

func (artifactGuidanceTool) Description() string {
	return "Return artifact contract guidance and next expected authoring steps."
}

func (artifactGuidanceTool) Handle(ctx context.Context, runtime Runtime, req ArtifactGuidanceRequest) (ArtifactGuidanceResult, error) {
	attrs := map[string]any{"tool": "artifact.guidance"}
	id := strings.TrimSpace(req.ArtifactID)
	if id != "" {
		attrs["artifact_id"] = id
	}
	if id == "" {
		if req.Type == "" {
			var err error
			id, err = runtime.ResolveArtifactID(ctx, "")
			if err != nil {
				return ArtifactGuidanceResult{}, errs.Decorate(err, errs.WithAttrs(attrs))
			}
			attrs["artifact_id"] = id
		}
	}
	if req.Type != "" {
		attrs["artifact_type"] = req.Type
	}
	contract, next, err := runtime.ArtifactStore().GuidanceContext(ctx, id, req.Type)
	if err != nil {
		return ArtifactGuidanceResult{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	return ArtifactGuidanceResult{ArtifactID: id, Contract: artifactContractPayloadFrom(contract), Next: next}, nil
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
