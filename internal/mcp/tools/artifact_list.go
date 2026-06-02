package tools

import (
	"context"

	"github.com/manuelibar/workbench/internal/errs"
)

type artifactListTool struct{}

type ArtifactListResult struct {
	Artifacts []artifactSummaryPayload `json:"artifacts"`
}

func init() {
	defaultRegistry.Register(typedTool[map[string]any, ArtifactListResult]{impl: artifactListTool{}})
}

func (artifactListTool) Name() string {
	return "list"
}

func (artifactListTool) Group() string {
	return "artifact"
}

func (artifactListTool) Description() string {
	return "List artifacts in the configured artifact store."
}

func (artifactListTool) Handle(ctx context.Context, runtime Runtime, _ map[string]any) (ArtifactListResult, error) {
	attrs := map[string]any{"tool": "artifact.list"}
	summaries, err := runtime.ArtifactStore().ListContext(ctx)
	if err != nil {
		return ArtifactListResult{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	return ArtifactListResult{Artifacts: artifactSummaryPayloadsFrom(summaries)}, nil
}
