package mcp

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/manuelibar/workbench/internal/artifacts"
	mcpresources "github.com/manuelibar/workbench/internal/mcp/resources"
)

type CapabilityKind string

const (
	CapabilityTool             CapabilityKind = "tool"
	CapabilityResource         CapabilityKind = "resource"
	CapabilityResourceTemplate CapabilityKind = "resource_template"
	CapabilityPrompt           CapabilityKind = "prompt"
)

type capabilityVisibility string

const (
	visibleAlways           capabilityVisibility = "always"
	visibleArtifactSelected capabilityVisibility = "artifact_selected"
)

type CapabilityDefinition struct {
	ID          string               `json:"-"`
	Kind        CapabilityKind       `json:"kind"`
	Name        string               `json:"name,omitempty"`
	URI         string               `json:"uri,omitempty"`
	URITemplate string               `json:"uri_template,omitempty"`
	Description string               `json:"description"`
	Group       string               `json:"group"`
	Visibility  capabilityVisibility `json:"visibility"`
}

type CapabilitySummary struct {
	ID          string         `json:"-"`
	Kind        CapabilityKind `json:"kind"`
	Name        string         `json:"name,omitempty"`
	URI         string         `json:"uri,omitempty"`
	URITemplate string         `json:"uri_template,omitempty"`
	Description string         `json:"description"`
	Group       string         `json:"group"`
	Active      bool           `json:"active"`
}

type CapabilitySurface struct {
	Tools             []CapabilitySummary `json:"tools"`
	Resources         []CapabilitySummary `json:"resources"`
	ResourceTemplates []CapabilitySummary `json:"resource_templates"`
	Prompts           []CapabilitySummary `json:"prompts"`
}

type CapabilityCatalog struct {
	definitions []CapabilityDefinition
	byID        map[string]CapabilityDefinition
}

type CapabilityPlan struct {
	State  ContextState      `json:"state"`
	Active CapabilitySurface `json:"active"`
	All    CapabilitySurface `json:"all"`
}

type Planner interface {
	Plan(context.Context, ContextState, CapabilityCatalog) (CapabilityPlan, error)
}

type deterministicPlanner struct{}

func NewCapabilityCatalog() CapabilityCatalog {
	tools := registeredTools()
	defs := make([]CapabilityDefinition, 0, len(tools)+len(mcpresources.All())+len(mcpresources.Templates()))
	for _, tool := range tools {
		defs = append(defs, toolCapabilityDef(tool))
	}
	for _, resource := range mcpresources.All() {
		defs = append(defs, resourceCapabilityDef(resource))
	}
	for _, template := range mcpresources.Templates() {
		defs = append(defs, templateCapabilityDef(template))
	}
	c := CapabilityCatalog{definitions: defs, byID: map[string]CapabilityDefinition{}}
	for _, def := range defs {
		c.byID[def.ID] = def
	}
	return c
}

func toolCapabilityDef(def registeredTool) CapabilityDefinition {
	name := def.FullName()
	return toolDef(name, name, def.Description(), def.Group(), toolVisibility(name))
}

func resourceCapabilityDef(def mcpresources.Definition) CapabilityDefinition {
	return resourceDef(mcpresources.Key(def), def.URI(), def.Description(), def.Group(), capabilityVisibility(def.Visibility()))
}

func templateCapabilityDef(def mcpresources.TemplateDefinition) CapabilityDefinition {
	return templateDef(mcpresources.TemplateKey(def), def.URITemplate(), def.Description(), def.Group(), capabilityVisibility(def.Visibility()))
}

func toolDef(id, name, description, group string, visibility capabilityVisibility) CapabilityDefinition {
	return CapabilityDefinition{
		ID:          id,
		Kind:        CapabilityTool,
		Name:        name,
		Description: description,
		Group:       group,
		Visibility:  visibility,
	}
}

func toolVisibility(name string) capabilityVisibility {
	switch name {
	case "context", "artifact.begin", "artifact.get", "artifact.list":
		return visibleAlways
	}
	if strings.HasPrefix(name, "artifact.") {
		return visibleArtifactSelected
	}
	return visibleAlways
}

func resourceDef(id, uri, description, group string, visibility capabilityVisibility) CapabilityDefinition {
	return CapabilityDefinition{
		ID:          id,
		Kind:        CapabilityResource,
		URI:         uri,
		Description: description,
		Group:       group,
		Visibility:  visibility,
	}
}

func templateDef(id, uriTemplate, description, group string, visibility capabilityVisibility) CapabilityDefinition {
	return CapabilityDefinition{
		ID:          id,
		Kind:        CapabilityResourceTemplate,
		URITemplate: uriTemplate,
		Description: description,
		Group:       group,
		Visibility:  visibility,
	}
}

func (deterministicPlanner) Plan(_ context.Context, state ContextState, catalog CapabilityCatalog) (CapabilityPlan, error) {
	active := map[string]bool{}
	for _, def := range catalog.definitions {
		if visible(def, state) {
			active[def.ID] = true
		}
	}
	return CapabilityPlan{
		State:  state,
		Active: catalog.surfaceFor(state, active, false),
		All:    catalog.surfaceFor(state, active, true),
	}, nil
}

func (c CapabilityCatalog) ValidatePlan(state ContextState, plan CapabilityPlan) error {
	seen := map[string]bool{}
	add := func(summary CapabilitySummary) error {
		def, ok := c.byID[summary.ID]
		if !ok {
			return fmt.Errorf("capability plan references unknown capability %q", summary.ID)
		}
		if def.Kind != summary.Kind {
			return fmt.Errorf("capability plan kind mismatch for %q", summary.ID)
		}
		if def.Visibility == visibleAlways && !summary.Active {
			return fmt.Errorf("capability plan omitted always-on capability %q", summary.ID)
		}
		if !summary.Active {
			return fmt.Errorf("capability plan contains inactive capability %q", summary.ID)
		}
		if !visible(def, state) && summary.Active {
			return fmt.Errorf("capability plan activated %q outside its visibility rule", summary.ID)
		}
		seen[summary.ID] = true
		return nil
	}
	for _, summary := range plan.Active.Tools {
		if err := add(summary); err != nil {
			return err
		}
	}
	for _, summary := range plan.Active.Resources {
		if err := add(summary); err != nil {
			return err
		}
	}
	for _, summary := range plan.Active.ResourceTemplates {
		if err := add(summary); err != nil {
			return err
		}
	}
	for _, def := range c.definitions {
		if visible(def, state) && !seen[def.ID] {
			return fmt.Errorf("capability plan missing visible capability %q", def.ID)
		}
	}
	return nil
}

func (c CapabilityCatalog) surfaceFor(state ContextState, active map[string]bool, includeInactive bool) CapabilitySurface {
	var surface CapabilitySurface
	for _, def := range c.definitions {
		isActive := active[def.ID]
		if !includeInactive && !isActive {
			continue
		}
		summary := def.summary(state, isActive)
		switch def.Kind {
		case CapabilityTool:
			surface.Tools = append(surface.Tools, summary)
		case CapabilityResource:
			if summary.URI != "" {
				surface.Resources = append(surface.Resources, summary)
			}
		case CapabilityResourceTemplate:
			surface.ResourceTemplates = append(surface.ResourceTemplates, summary)
		case CapabilityPrompt:
			surface.Prompts = append(surface.Prompts, summary)
		}
	}
	sortCapabilitySurface(&surface)
	return surface
}

func (d CapabilityDefinition) summary(state ContextState, active bool) CapabilitySummary {
	s := CapabilitySummary{
		ID:          d.ID,
		Kind:        d.Kind,
		Name:        d.Name,
		URI:         d.URI,
		URITemplate: d.URITemplate,
		Description: d.Description,
		Group:       d.Group,
		Active:      active,
	}
	if d.ID == mcpresources.SelectedArtifactID && state.ArtifactID != nil && *state.ArtifactID != "" {
		s.URI = mcpresources.ArtifactURI(*state.ArtifactID)
	}
	return s
}

func decorateSelectedArtifactSurface(surface *CapabilitySurface, artifact artifacts.Summary) {
	def := mcpresources.NewSelectedArtifactResource(selectedArtifactResource(artifact))
	for i := range surface.Resources {
		if surface.Resources[i].ID != mcpresources.SelectedArtifactID {
			continue
		}
		surface.Resources[i].Name = def.Name()
		surface.Resources[i].URI = def.URI()
		surface.Resources[i].Description = def.Description()
	}
}

func selectedArtifactResource(artifact artifacts.Summary) mcpresources.SelectedArtifact {
	return mcpresources.SelectedArtifact{
		ID:     artifact.ID,
		Type:   artifact.Type,
		Title:  artifact.Title,
		Status: artifact.Status,
	}
}

func visible(def CapabilityDefinition, state ContextState) bool {
	switch def.Visibility {
	case visibleAlways:
		return true
	case visibleArtifactSelected:
		return state.ArtifactID != nil && *state.ArtifactID != ""
	default:
		return false
	}
}

func sortCapabilitySurface(surface *CapabilitySurface) {
	sort.Slice(surface.Tools, func(i, j int) bool { return surface.Tools[i].Name < surface.Tools[j].Name })
	sort.Slice(surface.Resources, func(i, j int) bool { return surface.Resources[i].URI < surface.Resources[j].URI })
	sort.Slice(surface.ResourceTemplates, func(i, j int) bool {
		return surface.ResourceTemplates[i].URITemplate < surface.ResourceTemplates[j].URITemplate
	})
	sort.Slice(surface.Prompts, func(i, j int) bool { return surface.Prompts[i].Name < surface.Prompts[j].Name })
}
