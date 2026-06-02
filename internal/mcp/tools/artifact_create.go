package tools

import (
	"context"

	"github.com/manuelibar/workbench/internal/artifacts"
	"github.com/manuelibar/workbench/internal/errs"
)

type artifactCreateTool struct{}

type ArtifactCreateRequest struct {
	Type   string `json:"type" jsonschema:"artifact contract type such as rfc, adr, charter, spec, risk, assumption"`
	Title  string `json:"title" jsonschema:"artifact title"`
	Status string `json:"status,omitempty" jsonschema:"artifact status; defaults to draft"`
	Focus  string `json:"focus,omitempty" jsonschema:"optional focus to record in the generated draft"`
}

type ArtifactCreateResult struct {
	Artifact artifactSummaryPayload `json:"artifact"`
}

func init() {
	register[ArtifactCreateRequest, ArtifactCreateResult](artifactCreateTool{})
}

func (artifactCreateTool) Name() string {
	return "create"
}

func (artifactCreateTool) Group() string {
	return "artifact"
}

func (artifactCreateTool) Description() string {
	return "Create a typed Markdown artifact draft and return its summary."
}

func (artifactCreateTool) Handle(ctx context.Context, host Host, req ArtifactCreateRequest) (ArtifactCreateResult, error) {
	attrs := map[string]any{"tool": "artifact.create"}
	artifact, err := host.CreateArtifact(ctx, artifacts.CreateRequest{
		Type:   req.Type,
		Title:  req.Title,
		Status: req.Status,
		Focus:  req.Focus,
	})
	if err != nil {
		return ArtifactCreateResult{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	return ArtifactCreateResult{Artifact: artifactSummaryPayload{
		ID:      artifact.ID,
		Type:    artifact.Type,
		Title:   artifact.Title,
		Status:  artifact.Status,
		Created: artifact.Created,
		Updated: artifact.Updated,
	}}, nil
}
