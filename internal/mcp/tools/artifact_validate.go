package tools

import (
	"context"

	"github.com/manuelibar/workbench/internal/artifacts"
	"github.com/manuelibar/workbench/internal/errs"
)

type artifactValidateTool struct{}

type ArtifactValidateRequest struct {
	ArtifactID string `json:"artifact_id,omitempty" jsonschema:"artifact id; defaults to selected artifact"`
}

type artifactValidationPayload struct {
	ArtifactID string   `json:"artifact_id"`
	Valid      bool     `json:"valid"`
	Issues     []string `json:"issues"`
}

func init() {
	defaultRegistry.Register(typedTool[ArtifactValidateRequest, artifactValidationPayload]{impl: artifactValidateTool{}})
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

func (artifactValidateTool) Handle(ctx context.Context, runtime Runtime, req ArtifactValidateRequest) (artifactValidationPayload, error) {
	attrs := map[string]any{"tool": "artifact.validate"}
	id, err := runtime.ResolveArtifactID(ctx, req.ArtifactID)
	if err != nil {
		return artifactValidationPayload{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	attrs["artifact_id"] = id
	validation, err := runtime.ArtifactStore().ValidateContext(ctx, id)
	if err != nil {
		return artifactValidationPayload{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	return artifactValidationPayloadFrom(validation), nil
}

func artifactValidationPayloadFrom(validation artifacts.Validation) artifactValidationPayload {
	return artifactValidationPayload{
		ArtifactID: validation.ArtifactID,
		Valid:      validation.Valid,
		Issues:     append([]string(nil), validation.Issues...),
	}
}
