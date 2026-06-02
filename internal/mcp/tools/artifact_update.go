package tools

import (
	"context"
	"strings"

	"github.com/manuelibar/workbench/internal/artifacts"
	"github.com/manuelibar/workbench/internal/errs"
)

type artifactUpdateTool struct{}

type ArtifactUpdateRequest struct {
	ArtifactID   string            `json:"artifact_id,omitempty" jsonschema:"artifact id; defaults to selected artifact"`
	Title        string            `json:"title,omitempty" jsonschema:"new artifact title"`
	Status       string            `json:"status,omitempty" jsonschema:"new artifact status"`
	SetSections  map[string]string `json:"set_sections,omitempty" jsonschema:"replace section bodies by section key"`
	ClearSection []string          `json:"clear_section,omitempty" jsonschema:"section keys to clear"`
}

func init() {
	defaultRegistry.Register(typedTool[ArtifactUpdateRequest, artifactPayload]{impl: artifactUpdateTool{}})
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

func (artifactUpdateTool) Handle(ctx context.Context, runtime Runtime, req ArtifactUpdateRequest) (artifactPayload, error) {
	attrs := map[string]any{"tool": "artifact.update"}
	id, err := runtime.ResolveArtifactID(ctx, req.ArtifactID)
	if err != nil {
		return artifactPayload{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	attrs["artifact_id"] = id
	artifact, err := runtime.ArtifactStore().UpdateContext(ctx, id, artifacts.UpdateRequest{
		Title:        req.Title,
		Status:       req.Status,
		SetSections:  req.SetSections,
		ClearSection: req.ClearSection,
	})
	if err != nil {
		return artifactPayload{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	if strings.TrimSpace(req.Title) != "" || strings.TrimSpace(req.Status) != "" {
		runtime.RefreshSelectedArtifactResource(artifact.Summary)
	}
	return artifactPayloadFrom(artifact), nil
}
