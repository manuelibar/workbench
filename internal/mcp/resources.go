package mcp

import (
	"context"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/manuelibar/workbench/internal/artifacts"
	"github.com/manuelibar/workbench/internal/errs"
	mcpresources "github.com/manuelibar/workbench/internal/mcp/resources"
)

func (s *Server) readScopeResource(ctx context.Context, req *mcpsdk.ReadResourceRequest) (*mcpsdk.ReadResourceResult, error) {
	attrs := map[string]any{"resource": mcpresources.ScopeURI}
	state := s.scope.Snapshot()
	plan, err := s.plan(ctx, state)
	if err != nil {
		return nil, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	var scoped *artifacts.Summary
	if state.ArtifactID != nil && *state.ArtifactID != "" {
		if artifact, err := s.artifacts.GetContext(ctx, *state.ArtifactID); err == nil {
			scoped = &artifact.Summary
		}
	}
	text := scopeDocument(state, plan, scoped)
	return &mcpsdk.ReadResourceResult{
		Contents: []*mcpsdk.ResourceContents{{
			URI:      req.Params.URI,
			MIMEType: "text/markdown",
			Text:     text,
		}},
	}, nil
}

func (s *Server) readArtifactResource(ctx context.Context, req *mcpsdk.ReadResourceRequest) (*mcpsdk.ReadResourceResult, error) {
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
	markdown, err := s.readScopedArtifactMarkdown(ctx, id)
	if err != nil {
		return nil, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	return &mcpsdk.ReadResourceResult{
		Contents: []*mcpsdk.ResourceContents{{
			URI:      req.Params.URI,
			MIMEType: "text/markdown",
			Text:     markdown,
		}},
	}, nil
}

func artifactIDFromURI(uri string) string {
	return mcpresources.ArtifactIDFromURI(uri)
}
