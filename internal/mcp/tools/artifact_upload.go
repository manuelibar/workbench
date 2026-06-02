package tools

import (
	"context"

	"github.com/manuelibar/workbench/internal/errs"
)

type artifactUploadTool struct{}

type ArtifactUploadRequest struct {
	Markdown string `json:"markdown,omitempty" jsonschema:"full artifact Markdown; when omitted, upload the server-managed local scoped file"`
}

type ArtifactUploadResult struct {
	Artifact artifactSummaryPayload `json:"artifact"`
}

func init() {
	register[ArtifactUploadRequest, ArtifactUploadResult](artifactUploadTool{})
}

func (artifactUploadTool) Name() string {
	return "upload"
}

func (artifactUploadTool) Group() string {
	return "artifact"
}

func (artifactUploadTool) Description() string {
	return "Upload the full Markdown for the artifact currently in scope."
}

func (artifactUploadTool) Handle(ctx context.Context, host Host, req ArtifactUploadRequest) (ArtifactUploadResult, error) {
	attrs := map[string]any{"tool": "artifact.upload"}
	artifact, err := host.UploadScopedArtifact(ctx, req.Markdown)
	if err != nil {
		return ArtifactUploadResult{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	return ArtifactUploadResult{Artifact: artifactSummaryPayload{
		ID:      artifact.ID,
		Type:    artifact.Type,
		Title:   artifact.Title,
		Status:  artifact.Status,
		Created: artifact.Created,
		Updated: artifact.Updated,
	}}, nil
}
