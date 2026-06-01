package mcp

import (
	"context"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/manuelibar/workbench/internal/artifacts"
	"github.com/manuelibar/workbench/internal/errs"
	mcpresources "github.com/manuelibar/workbench/internal/mcp/resources"
)

func (s *Server) readContextResource(ctx context.Context, req *mcpsdk.ReadResourceRequest) (*mcpsdk.ReadResourceResult, error) {
	attrs := map[string]any{"resource": mcpresources.ContextURI}
	state := s.context.Snapshot()
	plan, err := s.plan(ctx, state)
	if err != nil {
		return nil, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	var selected *artifacts.Summary
	if state.ArtifactID != nil && *state.ArtifactID != "" {
		if artifact, err := s.artifacts.Get(*state.ArtifactID); err == nil {
			selected = &artifact.Summary
		}
	}
	text := contextDocument(state, plan, selected)
	return &mcpsdk.ReadResourceResult{
		Contents: []*mcpsdk.ResourceContents{{
			URI:      req.Params.URI,
			MIMEType: "text/markdown",
			Text:     text,
		}},
	}, nil
}

func (s *Server) readArtifactResource(_ context.Context, req *mcpsdk.ReadResourceRequest) (*mcpsdk.ReadResourceResult, error) {
	attrs := map[string]any{"uri": req.Params.URI}
	id := artifactIDFromURI(req.Params.URI)
	if id == "" {
		return nil, errs.New(
			"Resource URI is invalid",
			errs.WithSentinel(errs.ErrInvalid),
			errs.WithCode(errCodeResourceURIInvalid),
			errs.WithSeverity(errs.SeverityWarning),
			errs.WithAttrs(attrs),
			errs.WithRetryable(false),
		)
	}
	attrs["resource"] = req.Params.URI
	attrs["artifact_id"] = id
	artifact, err := s.artifacts.Get(id)
	if err != nil {
		return nil, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	return &mcpsdk.ReadResourceResult{
		Contents: []*mcpsdk.ResourceContents{{
			URI:      req.Params.URI,
			MIMEType: "text/markdown",
			Text:     artifact.Markdown,
		}},
	}, nil
}

func artifactIDFromURI(uri string) string {
	return mcpresources.ArtifactIDFromURI(uri)
}
