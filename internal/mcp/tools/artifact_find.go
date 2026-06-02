package tools

import (
	"context"

	"github.com/manuelibar/workbench/internal/errs"
)

type artifactFindTool struct{}

type ArtifactFindRequest struct {
	Query  string `json:"query,omitempty" jsonschema:"case-insensitive text query matched against artifact id, title, type, and status"`
	Type   string `json:"type,omitempty" jsonschema:"artifact type filter"`
	Status string `json:"status,omitempty" jsonschema:"artifact status filter"`
	Limit  int    `json:"limit,omitempty" jsonschema:"maximum summaries to return"`
}

type ArtifactFindResult struct {
	Artifacts []artifactSummaryPayload `json:"artifacts"`
}

func init() {
	register[ArtifactFindRequest, ArtifactFindResult](artifactFindTool{})
}

func (artifactFindTool) Name() string {
	return "find"
}

func (artifactFindTool) Group() string {
	return "artifact"
}

func (artifactFindTool) Description() string {
	return "Find artifact summaries by query, type, status, and optional limit."
}

func (artifactFindTool) Handle(ctx context.Context, host Host, req ArtifactFindRequest) (ArtifactFindResult, error) {
	attrs := map[string]any{"tool": "artifact.find"}
	summaries, err := host.FindArtifacts(ctx, req)
	if err != nil {
		return ArtifactFindResult{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	result := ArtifactFindResult{
		Artifacts: make([]artifactSummaryPayload, 0, len(summaries)),
	}
	for _, summary := range summaries {
		result.Artifacts = append(result.Artifacts, artifactSummaryPayload{
			ID:      summary.ID,
			Type:    summary.Type,
			Title:   summary.Title,
			Status:  summary.Status,
			Created: summary.Created,
			Updated: summary.Updated,
		})
	}
	if result.Artifacts == nil {
		result.Artifacts = []artifactSummaryPayload{}
	}
	return result, nil
}
