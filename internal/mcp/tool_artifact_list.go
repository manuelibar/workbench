package mcp

import (
	"context"

	"github.com/manuelibar/workbench/internal/errs"
)

type artifactListTool struct{}

func init() {
	registerTool[map[string]any, ArtifactListResult](artifactListTool{})
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

func (artifactListTool) Handle(ctx context.Context, s *Server, _ map[string]any) (ArtifactListResult, error) {
	attrs := map[string]any{"tool": "artifact.list"}
	summaries, err := s.artifacts.ListContext(ctx)
	if err != nil {
		return ArtifactListResult{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	return ArtifactListResult{Artifacts: artifactSummaryPayloadsFrom(summaries)}, nil
}
