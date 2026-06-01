package mcp

import (
	"context"
	"strings"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/manuelibar/workbench/internal/errs"
)

func (s *Server) readContextResource(ctx context.Context, req *mcpsdk.ReadResourceRequest) (*mcpsdk.ReadResourceResult, error) {
	attrs := map[string]any{"resource": "workbench:///context"}
	state := s.context.Snapshot()
	plan, err := s.plan(ctx, state)
	if err != nil {
		return nil, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	var selected *ArtifactSummary
	if state.ArtifactID != nil && *state.ArtifactID != "" {
		if artifact, err := s.artifacts.Get(*state.ArtifactID); err == nil {
			selected = &artifact.ArtifactSummary
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
	id, ok := strings.CutPrefix(uri, "workbench:///artifacts/")
	if !ok {
		return ""
	}
	id = strings.TrimSpace(id)
	if id == "" || strings.Contains(id, "/") {
		return ""
	}
	return id
}
