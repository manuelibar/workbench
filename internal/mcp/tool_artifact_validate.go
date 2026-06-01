package mcp

import (
	"context"

	"github.com/manuelibar/workbench/internal/errs"
)

type artifactValidateTool struct{}

func init() {
	registerTool[ArtifactValidateRequest, artifactValidationPayload](artifactValidateTool{})
}

func (artifactValidateTool) Name() string {
	return "validate"
}

func (artifactValidateTool) Group() string {
	return "artifact"
}

func (artifactValidateTool) Description() string {
	return "Validate selected artifact Markdown against its type contract."
}

func (artifactValidateTool) Handle(_ context.Context, s *Server, req ArtifactValidateRequest) (artifactValidationPayload, error) {
	attrs := map[string]any{"tool": "artifact.validate"}
	id, err := s.resolveArtifactID(req.ArtifactID)
	if err != nil {
		return artifactValidationPayload{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	attrs["artifact_id"] = id
	validation, err := s.artifacts.Validate(id)
	if err != nil {
		return artifactValidationPayload{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	return artifactValidationPayloadFrom(validation), nil
}
