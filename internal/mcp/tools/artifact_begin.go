package tools

import (
	"context"

	"github.com/manuelibar/workbench/internal/artifacts"
	"github.com/manuelibar/workbench/internal/errs"
)

type artifactBeginTool struct{}

type ArtifactBeginRequest struct {
	Type   string `json:"type" jsonschema:"artifact contract type such as rfc, adr, charter, spec, risk, assumption"`
	Title  string `json:"title" jsonschema:"artifact title"`
	Status string `json:"status,omitempty" jsonschema:"artifact status; defaults to draft"`
	Focus  string `json:"focus,omitempty" jsonschema:"optional focus to record in the generated draft"`
	Select bool   `json:"select,omitempty" jsonschema:"select the new artifact in context after creation"`
}

type ArtifactBeginResult struct {
	Artifact artifactPayload      `json:"artifact"`
	Context  *ContextualizeResult `json:"context,omitempty"`
}

func init() {
	defaultRegistry.Register(typedTool[ArtifactBeginRequest, ArtifactBeginResult]{impl: artifactBeginTool{}})
}

func (artifactBeginTool) Name() string {
	return "begin"
}

func (artifactBeginTool) Group() string {
	return "artifact"
}

func (artifactBeginTool) Description() string {
	return "Create a typed Markdown artifact draft in the configured artifact store."
}

func (artifactBeginTool) Handle(ctx context.Context, runtime Runtime, req ArtifactBeginRequest) (ArtifactBeginResult, error) {
	attrs := map[string]any{"tool": "artifact.begin"}
	artifact, err := runtime.ArtifactStore().BeginContext(ctx, artifacts.BeginRequest{
		Type:   req.Type,
		Title:  req.Title,
		Status: req.Status,
		Focus:  req.Focus,
	})
	if err != nil {
		return ArtifactBeginResult{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	attrs["artifact_id"] = artifact.ID
	result := ArtifactBeginResult{Artifact: artifactPayloadFrom(artifact)}
	if req.Select {
		contextualizeResult, err := runtime.SelectArtifact(ctx, artifact.ID, req.Focus)
		if err != nil {
			return ArtifactBeginResult{}, errs.Decorate(err, errs.WithAttrs(attrs))
		}
		result.Context = &contextualizeResult
	}
	return result, nil
}
