package mcp

import (
	"context"

	"github.com/manuelibar/workbench/internal/errs"
)

type artifactGetTool struct{}

func init() {
	registerTool[ArtifactGetRequest, artifactPayload](artifactGetTool{})
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

func (artifactGetTool) Handle(ctx context.Context, s *Server, req ArtifactGetRequest) (artifactPayload, error) {
	attrs := map[string]any{
		"tool":        "artifact.get",
		"artifact_id": req.ArtifactID,
	}
	artifact, err := s.artifacts.GetContext(ctx, req.ArtifactID)
	if err != nil {
		return artifactPayload{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	return artifactPayloadFrom(artifact), nil
}
