package mcp

import (
	"context"
	"strings"

	"github.com/manuelibar/workbench/internal/artifacts"
	"github.com/manuelibar/workbench/internal/errs"
)

type artifactUpdateTool struct{}

func init() {
	registerTool[ArtifactUpdateRequest, artifactPayload](artifactUpdateTool{})
}

func (artifactUpdateTool) Name() string {
	return "update"
}

func (artifactUpdateTool) Group() string {
	return "artifact"
}

func (artifactUpdateTool) Description() string {
	return "Update selected artifact metadata or replace/clear named Markdown sections."
}

func (artifactUpdateTool) Handle(ctx context.Context, s *Server, req ArtifactUpdateRequest) (artifactPayload, error) {
	attrs := map[string]any{"tool": "artifact.update"}
	id, err := s.resolveArtifactID(ctx, req.ArtifactID)
	if err != nil {
		return artifactPayload{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	attrs["artifact_id"] = id
	artifact, err := s.artifacts.UpdateContext(ctx, id, artifacts.UpdateRequest{
		Title:        req.Title,
		Status:       req.Status,
		SetSections:  req.SetSections,
		ClearSection: req.ClearSection,
	})
	if err != nil {
		return artifactPayload{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	if strings.TrimSpace(req.Title) != "" || strings.TrimSpace(req.Status) != "" {
		s.refreshSelectedArtifactResource(artifact.Summary)
	}
	return artifactPayloadFrom(artifact), nil
}
