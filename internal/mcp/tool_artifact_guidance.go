package mcp

import (
	"context"
	"strings"

	"github.com/manuelibar/workbench/internal/errs"
)

type artifactGuidanceTool struct{}

func init() {
	registerTool[ArtifactGuidanceRequest, ArtifactGuidanceResult](artifactGuidanceTool{})
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

func (artifactGuidanceTool) Handle(ctx context.Context, s *Server, req ArtifactGuidanceRequest) (ArtifactGuidanceResult, error) {
	attrs := map[string]any{"tool": "artifact.guidance"}
	id := strings.TrimSpace(req.ArtifactID)
	if id != "" {
		attrs["artifact_id"] = id
	}
	if id == "" {
		if req.Type == "" {
			var err error
			id, err = s.resolveArtifactID(ctx, "")
			if err != nil {
				return ArtifactGuidanceResult{}, errs.Decorate(err, errs.WithAttrs(attrs))
			}
			attrs["artifact_id"] = id
		}
	}
	if req.Type != "" {
		attrs["artifact_type"] = req.Type
	}
	contract, next, err := s.artifacts.GuidanceContext(ctx, id, req.Type)
	if err != nil {
		return ArtifactGuidanceResult{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	return ArtifactGuidanceResult{ArtifactID: id, Contract: artifactContractPayloadFrom(contract), Next: next}, nil
}
