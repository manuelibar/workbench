package mcpserver

import (
	"context"
	"fmt"
	"sort"
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
	ID          string               `json:"id"`
	Kind        CapabilityKind       `json:"kind"`
	Name        string               `json:"name,omitempty"`
	URI         string               `json:"uri,omitempty"`
	URITemplate string               `json:"uri_template,omitempty"`
	Description string               `json:"description"`
	Group       string               `json:"group"`
	Visibility  capabilityVisibility `json:"visibility"`
}

type CapabilitySummary struct {
	ID          string         `json:"id"`
	Kind        CapabilityKind `json:"kind"`
	Name        string         `json:"name,omitempty"`
	URI         string         `json:"uri,omitempty"`
	URITemplate string         `json:"uri_template,omitempty"`
	Description string         `json:"description"`
	Group       string         `json:"group"`
	Active      bool           `json:"active"`
}

type CapabilityIndex struct {
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
	State ContextState    `json:"state"`
	Index CapabilityIndex `json:"index"`
	All   CapabilityIndex `json:"all"`
}

type Planner interface {
	Plan(context.Context, ContextState, CapabilityCatalog) (CapabilityPlan, error)
}

type deterministicPlanner struct{}

func NewCapabilityCatalog() CapabilityCatalog {
	defs := []CapabilityDefinition{
		toolDef("tool.context", "context", "Read or patch focus/artifact context and return the raw context document plus capability sync status.", "core", visibleAlways),
		toolDef("tool.artifact.begin", "artifact.begin", "Create a typed Markdown artifact draft under docs/artifacts.", "artifacts", visibleAlways),
		toolDef("tool.artifact.list", "artifact.list", "List file-backed artifacts in docs/artifacts.", "artifacts", visibleAlways),
		toolDef("tool.artifact.get", "artifact.get", "Read one artifact by stable id.", "artifacts", visibleAlways),
		toolDef("tool.artifact.update", "artifact.update", "Update selected artifact metadata or section bodies.", "artifacts", visibleArtifactSelected),
		toolDef("tool.artifact.guidance", "artifact.guidance", "Return deterministic contract guidance for the selected artifact.", "artifacts", visibleArtifactSelected),
		toolDef("tool.artifact.validate", "artifact.validate", "Validate the selected artifact against its type contract.", "artifacts", visibleArtifactSelected),
		resourceDef("resource.context", "workbench:///context", "Current raw Workbench context document.", "core", visibleAlways),
		resourceDef("resource.artifact.selected", "", "Selected artifact Markdown resource.", "artifacts", visibleArtifactSelected),
		templateDef("template.artifacts", "workbench:///artifacts/{id}", "Read an artifact Markdown file by stable id.", "artifacts", visibleAlways),
	}
	c := CapabilityCatalog{definitions: defs, byID: map[string]CapabilityDefinition{}}
	for _, def := range defs {
		c.byID[def.ID] = def
	}
	return c
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
		State: state,
		Index: catalog.indexFor(state, active, false),
		All:   catalog.indexFor(state, active, true),
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
			return fmt.Errorf("capability plan index contains inactive capability %q", summary.ID)
		}
		if !visible(def, state) && summary.Active {
			return fmt.Errorf("capability plan activated %q outside its visibility rule", summary.ID)
		}
		seen[summary.ID] = true
		return nil
	}
	for _, summary := range plan.Index.Tools {
		if err := add(summary); err != nil {
			return err
		}
	}
	for _, summary := range plan.Index.Resources {
		if err := add(summary); err != nil {
			return err
		}
	}
	for _, summary := range plan.Index.ResourceTemplates {
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

func (c CapabilityCatalog) indexFor(state ContextState, active map[string]bool, includeInactive bool) CapabilityIndex {
	var index CapabilityIndex
	for _, def := range c.definitions {
		isActive := active[def.ID]
		if !includeInactive && !isActive {
			continue
		}
		summary := def.summary(state, isActive)
		switch def.Kind {
		case CapabilityTool:
			index.Tools = append(index.Tools, summary)
		case CapabilityResource:
			if summary.URI != "" {
				index.Resources = append(index.Resources, summary)
			}
		case CapabilityResourceTemplate:
			index.ResourceTemplates = append(index.ResourceTemplates, summary)
		case CapabilityPrompt:
			index.Prompts = append(index.Prompts, summary)
		}
	}
	sortCapabilityIndex(&index)
	return index
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
	if d.ID == "resource.artifact.selected" && state.ArtifactID != nil && *state.ArtifactID != "" {
		s.URI = "workbench:///artifacts/" + *state.ArtifactID
	}
	return s
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

func sortCapabilityIndex(index *CapabilityIndex) {
	sort.Slice(index.Tools, func(i, j int) bool { return index.Tools[i].Name < index.Tools[j].Name })
	sort.Slice(index.Resources, func(i, j int) bool { return index.Resources[i].URI < index.Resources[j].URI })
	sort.Slice(index.ResourceTemplates, func(i, j int) bool {
		return index.ResourceTemplates[i].URITemplate < index.ResourceTemplates[j].URITemplate
	})
	sort.Slice(index.Prompts, func(i, j int) bool { return index.Prompts[i].Name < index.Prompts[j].Name })
}
