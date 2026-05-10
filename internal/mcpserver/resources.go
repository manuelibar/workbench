package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/manuelibar/workbench/internal/onboarding"
)

const (
	// skillResourceURI is the always-on onboarding resource. Every connected
	// client should read this before calling tools. The body is the embedded
	// SKILL.md (see internal/onboarding).
	skillResourceURI = "workbench://skill"

	// URI templates for the project-scoped resources. Their handlers read
	// the URI directly from the request, parse out the variables, and look
	// the entity up in the currently-selected project.
	artifactURITemplate        = "workbench:///artifacts/{id}"
	artifactVersionURITemplate = "workbench:///artifacts/{id}/{version}"
	projectSkillURITemplate    = "workbench:///skills/{name}"
	projectPromptURITemplate   = "workbench:///prompts/{name}"
	blueprintURITemplate       = "workbench:///blueprints/{name}/{version}"
)

func (s *Server) registerResources(srv *mcp.Server) {
	srv.AddResource(
		&mcp.Resource{
			Name:        "Workbench onboarding skill",
			URI:         skillResourceURI,
			Description: "Top-level agent onboarding document. Read first.",
			MIMEType:    "text/markdown",
		},
		s.readSkillOnboardingResource,
	)

	srv.AddResourceTemplate(
		&mcp.ResourceTemplate{
			Name:        "Artifact (latest version)",
			URITemplate: artifactURITemplate,
			Description: "Read the latest version of an artifact in the currently-selected project.",
			MIMEType:    "text/plain",
		},
		s.readArtifactResource,
	)
	srv.AddResourceTemplate(
		&mcp.ResourceTemplate{
			Name:        "Artifact (specific version)",
			URITemplate: artifactVersionURITemplate,
			Description: "Read a specific version of an artifact.",
			MIMEType:    "text/plain",
		},
		s.readArtifactResource,
	)
	srv.AddResourceTemplate(
		&mcp.ResourceTemplate{
			Name:        "Skill",
			URITemplate: projectSkillURITemplate,
			Description: "Read a skill body (markdown) by name within the currently-selected project.",
			MIMEType:    "text/markdown",
		},
		s.readProjectSkillResource,
	)
	srv.AddResourceTemplate(
		&mcp.ResourceTemplate{
			Name:        "Prompt",
			URITemplate: projectPromptURITemplate,
			Description: "Read a prompt template body by name within the currently-selected project.",
			MIMEType:    "text/plain",
		},
		s.readProjectPromptResource,
	)
	srv.AddResourceTemplate(
		&mcp.ResourceTemplate{
			Name:        "Blueprint (specific version)",
			URITemplate: blueprintURITemplate,
			Description: "Read the JSON definition of a blueprint by (name, version) in the currently-selected project.",
			MIMEType:    "application/json",
		},
		s.readBlueprintResource,
	)
}

func (s *Server) readBlueprintResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	rest, ok := strings.CutPrefix(req.Params.URI, "workbench:///blueprints/")
	if !ok || rest == "" {
		return nil, fmt.Errorf("invalid blueprint URI: %q", req.Params.URI)
	}
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("invalid blueprint URI: %q (expected workbench:///blueprints/{name}/{version})", req.Params.URI)
	}
	name := parts[0]
	version, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("blueprint URI: version: %w", err)
	}
	sel := s.currentSelection()
	if sel.ProjectID == nil {
		return nil, errors.New("no project selected; blueprint resources are scoped to the current project")
	}
	b, err := s.store.GetBlueprintByVersion(ctx, *sel.ProjectID, name, version)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(b.Definition)
	if err != nil {
		return nil, fmt.Errorf("marshal blueprint definition: %w", err)
	}
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      req.Params.URI,
			MIMEType: "application/json",
			Text:     string(raw),
		}},
	}, nil
}

func (s *Server) readSkillOnboardingResource(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      req.Params.URI,
			MIMEType: "text/markdown",
			Text:     onboarding.SkillMarkdown(),
		}},
	}, nil
}

// readArtifactResource handles both `workbench:///artifacts/{id}` and
// `workbench:///artifacts/{id}/{version}` by parsing the URI directly.
func (s *Server) readArtifactResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	rest, ok := strings.CutPrefix(req.Params.URI, "workbench:///artifacts/")
	if !ok || rest == "" {
		return nil, fmt.Errorf("invalid artifact URI: %q", req.Params.URI)
	}
	parts := strings.SplitN(rest, "/", 2)
	id, err := uuid.Parse(parts[0])
	if err != nil {
		return nil, fmt.Errorf("artifact URI: id: %w", err)
	}
	version := 0
	if len(parts) == 2 && parts[1] != "" {
		v, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("artifact URI: version: %w", err)
		}
		version = v
	}
	v, err := s.store.GetArtifactVersion(ctx, id, version)
	if err != nil {
		return nil, err
	}
	text := v.ContentText
	if text == "" && len(v.Content) > 0 {
		raw, _ := json.Marshal(v.Content)
		text = string(raw)
	}
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      req.Params.URI,
			MIMEType: "text/plain",
			Text:     text,
		}},
	}, nil
}

func (s *Server) readProjectSkillResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	name, ok := strings.CutPrefix(req.Params.URI, "workbench:///skills/")
	if !ok || name == "" {
		return nil, fmt.Errorf("invalid skill URI: %q", req.Params.URI)
	}
	sel := s.currentSelection()
	if sel.ProjectID == nil {
		return nil, errors.New("no project selected; skill resources are scoped to the current project")
	}
	sk, err := s.store.GetSkillByName(ctx, *sel.ProjectID, name)
	if err != nil {
		return nil, err
	}
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      req.Params.URI,
			MIMEType: "text/markdown",
			Text:     sk.BodyMD,
		}},
	}, nil
}

func (s *Server) readProjectPromptResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	name, ok := strings.CutPrefix(req.Params.URI, "workbench:///prompts/")
	if !ok || name == "" {
		return nil, fmt.Errorf("invalid prompt URI: %q", req.Params.URI)
	}
	sel := s.currentSelection()
	if sel.ProjectID == nil {
		return nil, errors.New("no project selected; prompt resources are scoped to the current project")
	}
	p, err := s.store.GetPromptByName(ctx, *sel.ProjectID, name)
	if err != nil {
		return nil, err
	}
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      req.Params.URI,
			MIMEType: "text/plain",
			Text:     p.Body,
		}},
	}, nil
}
