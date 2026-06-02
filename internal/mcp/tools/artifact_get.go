package tools

import (
	"context"

	"github.com/manuelibar/workbench/internal/errs"
)

type artifactGetTool struct{}

type ArtifactGetRequest struct {
	ArtifactID string `json:"artifact_id" jsonschema:"artifact id"`
}

func init() {
	register[ArtifactGetRequest, artifactPayload](artifactGetTool{})
}

func (artifactGetTool) Name() string {
	return "get"
}

func (artifactGetTool) Group() string {
	return "artifact"
}

func (artifactGetTool) Description() string {
	return "Read an artifact Markdown resource by stable id."
}

func (artifactGetTool) Handle(ctx context.Context, runtime Runtime, req ArtifactGetRequest) (artifactPayload, error) {
	attrs := map[string]any{
		"tool":        "artifact.get",
		"artifact_id": req.ArtifactID,
	}
	artifact, err := runtime.ArtifactStore().GetContext(ctx, req.ArtifactID)
	if err != nil {
		return artifactPayload{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	return artifactPayloadFrom(artifact), nil
}
