package mcp

import (
	"context"
	"strings"

	"github.com/manuelibar/workbench/internal/artifacts"
	"github.com/manuelibar/workbench/internal/errs"
)

type artifactBeginTool struct{}

func init() {
	registerTool[ArtifactBeginRequest, ArtifactBeginResult](artifactBeginTool{})
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

func (artifactBeginTool) Handle(ctx context.Context, s *Server, req ArtifactBeginRequest) (ArtifactBeginResult, error) {
	attrs := map[string]any{"tool": "artifact.begin"}
	artifact, err := s.artifacts.BeginContext(ctx, artifacts.BeginRequest{
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
		state := s.context.Apply(ContextPatch{
			ArtifactID: PatchString{Present: true, Value: artifact.ID},
			Focus:      PatchString{Present: strings.TrimSpace(req.Focus) != "", Value: req.Focus},
		})
		contextResult, err := s.contextResult(ctx, state)
		if err != nil {
			return ArtifactBeginResult{}, errs.Decorate(err, errs.WithAttrs(attrs))
		}
		result.Context = &contextResult
	}
	return result, nil
}
