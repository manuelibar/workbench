package mcpserver

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *Server) readContextResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	state := s.context.Snapshot()
	plan, err := s.plan(ctx, state)
	if err != nil {
		return nil, err
	}
	var selected *ArtifactSummary
	if state.ArtifactID != nil && *state.ArtifactID != "" {
		if artifact, err := s.artifacts.Get(*state.ArtifactID); err == nil {
			selected = &artifact.ArtifactSummary
		}
	}
	text := contextDocument(state, plan, selected)
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      req.Params.URI,
			MIMEType: "text/markdown",
			Text:     text,
		}},
	}, nil
}

func (s *Server) readArtifactResource(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	id := artifactIDFromURI(req.Params.URI)
	if id == "" {
		return nil, fmt.Errorf("artifact resource URI must be workbench:///artifacts/{id}")
	}
	artifact, err := s.artifacts.Get(id)
	if err != nil {
		return nil, err
	}
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
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
