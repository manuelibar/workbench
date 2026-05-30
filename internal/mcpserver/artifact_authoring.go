package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/manuelibar/workbench/internal/domain"
	"github.com/manuelibar/workbench/internal/id"
	"github.com/manuelibar/workbench/internal/pgstore"
)

// artifact.begin ---------------------------------------------------------

type ArtifactBeginArgs struct {
	Type       string   `json:"type" jsonschema:"registered artifact type, e.g. rfc, adr, prd, spec"`
	Title      string   `json:"title" jsonschema:"human-readable artifact title"`
	Focus      string   `json:"focus,omitempty" jsonschema:"short steering string for this authoring pass"`
	Owners     []string `json:"owners,omitempty" jsonschema:"optional owner names or handles"`
	Tags       []string `json:"tags,omitempty" jsonschema:"optional discovery tags"`
	Parents    []string `json:"parents,omitempty" jsonschema:"optional parent artifact UUIDs"`
	SourceRefs []string `json:"source_refs,omitempty" jsonschema:"optional source refs such as workbench:///notes/{id}, issues, docs, or commits"`
}

type ArtifactBeginResult struct {
	Artifact ArtifactWire           `json:"artifact"`
	Version  ArtifactVersionWire    `json:"version"`
	Guidance ArtifactGuidanceResult `json:"guidance"`
}

func (s *Server) handleArtifactBegin(ctx context.Context, _ *mcp.CallToolRequest, args ArtifactBeginArgs) (*mcp.CallToolResult, ArtifactBeginResult, error) {
	pID, err := s.resolveProjectID("")
	if err != nil {
		return nil, ArtifactBeginResult{}, err
	}
	typ := normalizeArtifactType(args.Type)
	c, ok := artifactContractFor(typ)
	if !ok {
		return nil, ArtifactBeginResult{}, fmt.Errorf("artifact.begin: unsupported type %q; supported types: %s", args.Type, strings.Join(artifactContractTypes(), ", "))
	}
	title := strings.TrimSpace(args.Title)
	if title == "" {
		return nil, ArtifactBeginResult{}, errors.New("artifact.begin: title must not be empty")
	}

	artifactID, err := uuid.NewV7()
	if err != nil {
		return nil, ArtifactBeginResult{}, fmt.Errorf("artifact.begin: generate artifact id: %w", err)
	}
	effectiveFocus := strings.TrimSpace(args.Focus)
	if effectiveFocus == "" {
		effectiveFocus = s.currentSelection().Focus
	}
	now := time.Now().UTC()
	content := newArtifactContent(artifactID, typ, title, domain.ArtifactStatusDraft, effectiveFocus, now, c)
	if len(args.Owners) > 0 {
		content["owners"] = args.Owners
	}
	if len(args.Tags) > 0 {
		content["tags"] = args.Tags
	}
	if len(args.Parents) > 0 {
		content["parents"] = args.Parents
	}
	if len(args.SourceRefs) > 0 {
		content["source_refs"] = args.SourceRefs
	}
	contentText := artifactMarkdownProjection(content, c)

	artifact := domain.Artifact{
		ID:        artifactID,
		ProjectID: pID,
		Type:      typ,
		Status:    domain.ArtifactStatusDraft,
	}
	for _, raw := range args.Parents {
		parentID, err := uuid.Parse(raw)
		if err != nil {
			return nil, ArtifactBeginResult{}, fmt.Errorf("artifact.begin: parent: %w", err)
		}
		artifact.Parents = append(artifact.Parents, parentID)
	}
	if actor, ok := id.FromContext(ctx); ok {
		artifact.IdempotencyKey = actor.IdempotencyKey
	}
	created, err := s.store.CreateArtifact(ctx, pgstore.CreateArtifactInput{
		Artifact:    artifact,
		Content:     content,
		ContentText: contentText,
	})
	if err != nil {
		return nil, ArtifactBeginResult{}, err
	}

	refreshArgs := RefreshArgs{ArtifactID: created.ID.String()}
	if strings.TrimSpace(args.Focus) != "" {
		refreshArgs.Focus = args.Focus
	}
	if _, err := s.Refresh(ctx, refreshArgs); err != nil {
		return nil, ArtifactBeginResult{}, err
	}
	version, err := s.store.GetArtifactVersion(ctx, created.ID, 0)
	if err != nil {
		return nil, ArtifactBeginResult{}, err
	}
	validation, err := s.validateArtifact(ctx, created)
	if err != nil {
		return nil, ArtifactBeginResult{}, err
	}
	guidance := s.guidanceForValidation(created, c, effectiveFocus, validation)
	return nil, ArtifactBeginResult{
		Artifact: artifactToWire(created),
		Version: ArtifactVersionWire{
			Version:     version.Version,
			Content:     version.Content,
			ContentText: version.ContentText,
			CreatedAt:   version.CreatedAt,
		},
		Guidance: guidance,
	}, nil
}

// artifact.guidance / validate ------------------------------------------

type ArtifactSelectedArgs struct {
	ID string `json:"id,omitempty" jsonschema:"artifact UUID; defaults to the selected artifact"`
}

type ArtifactGuidanceResult struct {
	Artifact           ArtifactWire          `json:"artifact"`
	Contract           ArtifactContract      `json:"contract"`
	Focus              string                `json:"focus,omitempty"`
	Valid              bool                  `json:"valid"`
	MissingFields      []string              `json:"missing_fields,omitempty"`
	MissingFrontmatter []string              `json:"missing_frontmatter,omitempty"`
	MissingSections    []ArtifactSectionSpec `json:"missing_sections,omitempty"`
	Next               string                `json:"next"`
}

type ArtifactValidationResult struct {
	Artifact           ArtifactWire          `json:"artifact"`
	Type               string                `json:"type"`
	CheckedVersion     int                   `json:"checked_version"`
	Valid              bool                  `json:"valid"`
	MissingFields      []string              `json:"missing_fields,omitempty"`
	MissingFrontmatter []string              `json:"missing_frontmatter,omitempty"`
	MissingSections    []ArtifactSectionSpec `json:"missing_sections,omitempty"`
	Warnings           []string              `json:"warnings,omitempty"`
}

func (s *Server) handleArtifactGuidance(ctx context.Context, _ *mcp.CallToolRequest, args ArtifactSelectedArgs) (*mcp.CallToolResult, ArtifactGuidanceResult, error) {
	a, c, validation, err := s.loadArtifactContractValidation(ctx, args.ID)
	if err != nil {
		return nil, ArtifactGuidanceResult{}, err
	}
	return nil, s.guidanceForValidation(a, c, s.currentSelection().Focus, validation), nil
}

func (s *Server) handleArtifactValidate(ctx context.Context, _ *mcp.CallToolRequest, args ArtifactSelectedArgs) (*mcp.CallToolResult, ArtifactValidationResult, error) {
	_, _, validation, err := s.loadArtifactContractValidation(ctx, args.ID)
	if err != nil {
		return nil, ArtifactValidationResult{}, err
	}
	return nil, validation, nil
}

func (s *Server) loadArtifactContractValidation(ctx context.Context, rawID string) (domain.Artifact, ArtifactContract, ArtifactValidationResult, error) {
	artifactID, err := s.resolveArtifactID(rawID)
	if err != nil {
		return domain.Artifact{}, ArtifactContract{}, ArtifactValidationResult{}, err
	}
	a, err := s.store.GetArtifact(ctx, artifactID)
	if err != nil {
		return domain.Artifact{}, ArtifactContract{}, ArtifactValidationResult{}, err
	}
	c, ok := artifactContractFor(a.Type)
	if !ok {
		validation := ArtifactValidationResult{
			Artifact: artifactToWire(a),
			Type:     a.Type,
			Valid:    false,
			Warnings: []string{fmt.Sprintf("artifact type %q is not in the authoring contract registry", a.Type)},
		}
		return a, ArtifactContract{Type: a.Type}, validation, nil
	}
	validation, err := s.validateArtifact(ctx, a)
	if err != nil {
		return domain.Artifact{}, ArtifactContract{}, ArtifactValidationResult{}, err
	}
	return a, c, validation, nil
}

func (s *Server) validateArtifact(ctx context.Context, a domain.Artifact) (ArtifactValidationResult, error) {
	c, ok := artifactContractFor(a.Type)
	if !ok {
		return ArtifactValidationResult{
			Artifact: artifactToWire(a),
			Type:     a.Type,
			Valid:    false,
			Warnings: []string{fmt.Sprintf("artifact type %q is not in the authoring contract registry", a.Type)},
		}, nil
	}
	v, err := s.store.GetArtifactVersion(ctx, a.ID, 0)
	if err != nil {
		if !errors.Is(err, pgstore.ErrNotFound) {
			return ArtifactValidationResult{}, err
		}
		v = domain.ArtifactVersion{ArtifactID: a.ID}
	}

	content := v.Content
	if content == nil {
		content = map[string]any{}
	}
	missingFields := make([]string, 0, len(requiredArtifactFrontmatter))
	for _, key := range requiredArtifactFrontmatter {
		if strings.TrimSpace(contentString(content, key)) == "" {
			missingFields = append(missingFields, key)
		}
	}
	missingSections := []ArtifactSectionSpec{}
	sectionByKey := map[string]artifactSectionContent{}
	for _, section := range artifactSectionsFromContent(content, c) {
		sectionByKey[section.Key] = section
	}
	for _, spec := range c.RequiredSections {
		section := sectionByKey[spec.Key]
		if sectionBodyMissing(section.Body) {
			missingSections = append(missingSections, spec)
		}
	}
	missingFrontmatter := missingFrontmatterKeys(v.ContentText)
	valid := len(missingFields) == 0 && len(missingFrontmatter) == 0 && len(missingSections) == 0
	return ArtifactValidationResult{
		Artifact:           artifactToWire(a),
		Type:               a.Type,
		CheckedVersion:     v.Version,
		Valid:              valid,
		MissingFields:      missingFields,
		MissingFrontmatter: missingFrontmatter,
		MissingSections:    missingSections,
	}, nil
}

func (s *Server) guidanceForValidation(a domain.Artifact, c ArtifactContract, focus string, validation ArtifactValidationResult) ArtifactGuidanceResult {
	next := "Artifact satisfies its contract. Move it to review or sign it off when the human has accepted it."
	if len(validation.MissingFields) > 0 || len(validation.MissingFrontmatter) > 0 {
		next = "Regenerate the artifact body so content_jsonb and Markdown frontmatter carry the required metadata."
	}
	if len(validation.MissingSections) > 0 {
		first := validation.MissingSections[0]
		next = fmt.Sprintf("Fill the %q section next: %s", first.Title, first.Prompt)
	}
	if len(validation.Warnings) > 0 {
		next = validation.Warnings[0]
	}
	return ArtifactGuidanceResult{
		Artifact:           artifactToWire(a),
		Contract:           c,
		Focus:              strings.TrimSpace(focus),
		Valid:              validation.Valid,
		MissingFields:      validation.MissingFields,
		MissingFrontmatter: validation.MissingFrontmatter,
		MissingSections:    validation.MissingSections,
		Next:               next,
	}
}

// artifact.elicit --------------------------------------------------------

type ArtifactElicitArgs struct {
	ID      string `json:"id,omitempty" jsonschema:"artifact UUID; defaults to the selected artifact"`
	Section string `json:"section,omitempty" jsonschema:"section key to elicit; defaults to the first missing required section"`
}

type ArtifactElicitResult struct {
	Artifact    ArtifactWire             `json:"artifact"`
	Action      string                   `json:"action"`
	Section     ArtifactSectionSpec      `json:"section,omitempty"`
	Updated     bool                     `json:"updated"`
	NewVersion  int                      `json:"new_version,omitempty"`
	Unsupported bool                     `json:"unsupported,omitempty"`
	Validation  ArtifactValidationResult `json:"validation"`
}

func (s *Server) handleArtifactElicit(ctx context.Context, req *mcp.CallToolRequest, args ArtifactElicitArgs) (*mcp.CallToolResult, ArtifactElicitResult, error) {
	a, c, validation, err := s.loadArtifactContractValidation(ctx, args.ID)
	if err != nil {
		return nil, ArtifactElicitResult{}, err
	}
	target, ok := chooseElicitSection(c, validation, args.Section)
	if !ok {
		return nil, ArtifactElicitResult{
			Artifact:   artifactToWire(a),
			Action:     "complete",
			Updated:    false,
			Validation: validation,
		}, nil
	}

	title := artifactTitleForPrompt(ctx, s, a)
	focus := s.currentSelection().Focus
	message := fmt.Sprintf("Artifact %q (%s) is missing the %q section. %s", title, a.Type, target.Title, target.Prompt)
	if strings.TrimSpace(focus) != "" {
		message += " Current focus: " + strings.TrimSpace(focus) + "."
	}
	message += " Provide Markdown content for that section."

	res, err := req.Session.Elicit(ctx, &mcp.ElicitParams{
		Message: message,
		RequestedSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"body": map[string]any{
					"type":        "string",
					"description": "Markdown content for the requested artifact section.",
				},
			},
			"required": []string{"body"},
		},
	})
	if err != nil {
		if strings.Contains(err.Error(), "does not support") && strings.Contains(err.Error(), "elicitation") {
			return nil, ArtifactElicitResult{
				Artifact:    artifactToWire(a),
				Action:      "unsupported",
				Section:     target,
				Unsupported: true,
				Validation:  validation,
			}, nil
		}
		return nil, ArtifactElicitResult{}, fmt.Errorf("artifact.elicit: elicit: %w", err)
	}
	if res.Action != "accept" {
		return nil, ArtifactElicitResult{
			Artifact:   artifactToWire(a),
			Action:     res.Action,
			Section:    target,
			Updated:    false,
			Validation: validation,
		}, nil
	}
	body, _ := res.Content["body"].(string)
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, ArtifactElicitResult{}, errors.New("artifact.elicit: accepted response body must not be empty")
	}

	version, err := s.store.GetArtifactVersion(ctx, a.ID, 0)
	if err != nil {
		return nil, ArtifactElicitResult{}, err
	}
	now := time.Now().UTC()
	content := artifactContentWithSection(a, version.Content, c, target.Key, body, focus, now)
	contentText := artifactMarkdownProjection(content, c)
	updated, newVersion, err := s.store.AppendArtifactVersion(ctx, a.ID, content, contentText)
	if err != nil {
		return nil, ArtifactElicitResult{}, err
	}
	updatedValidation, err := s.validateArtifact(ctx, updated)
	if err != nil {
		return nil, ArtifactElicitResult{}, err
	}
	return nil, ArtifactElicitResult{
		Artifact:   artifactToWire(updated),
		Action:     res.Action,
		Section:    target,
		Updated:    true,
		NewVersion: newVersion,
		Validation: updatedValidation,
	}, nil
}

func chooseElicitSection(c ArtifactContract, validation ArtifactValidationResult, rawSection string) (ArtifactSectionSpec, bool) {
	rawSection = strings.TrimSpace(rawSection)
	if rawSection != "" {
		for _, spec := range c.RequiredSections {
			if spec.Key == rawSection {
				return spec, true
			}
		}
		for _, spec := range c.OptionalSections {
			if spec.Key == rawSection {
				return spec, true
			}
		}
		return ArtifactSectionSpec{Key: rawSection, Title: rawSection, Required: true}, true
	}
	if len(validation.MissingSections) == 0 {
		return ArtifactSectionSpec{}, false
	}
	return validation.MissingSections[0], true
}

func artifactTitleForPrompt(ctx context.Context, s *Server, a domain.Artifact) string {
	v, err := s.store.GetArtifactVersion(ctx, a.ID, 0)
	if err == nil && v.Content != nil {
		if title := strings.TrimSpace(contentString(v.Content, "title")); title != "" {
			return title
		}
	}
	return a.ID.String()
}

func (s *Server) resolveArtifactID(raw string) (uuid.UUID, error) {
	if raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return uuid.Nil, fmt.Errorf("artifact id: %w", err)
		}
		return id, nil
	}
	sel := s.currentSelection()
	if sel.ArtifactID == nil {
		return uuid.Nil, errors.New("no artifact selected; pass id or call refresh(artifact_id=...)")
	}
	return *sel.ArtifactID, nil
}

func (s *Server) registerArtifactAuthoringProject(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "artifact.begin",
		Description: "Create a typed draft artifact from the Workbench contract registry, write version 1 with normalized JSON plus Markdown frontmatter, select it, and return authoring guidance.",
	}, s.handleArtifactBegin)
}

func (s *Server) registerArtifactAuthoringSelected(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "artifact.guidance",
		Description: "Return the selected artifact contract, validation gaps, focus, and next authoring step.",
	}, s.handleArtifactGuidance)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "artifact.validate",
		Description: "Validate the selected artifact against its registered contract and body convention.",
	}, s.handleArtifactValidate)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "artifact.elicit",
		Description: "Ask the human via MCP elicitation for the first missing artifact section, then append a new version if accepted.",
	}, s.handleArtifactElicit)
}
