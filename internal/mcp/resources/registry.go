package resources

import (
	"fmt"
	"sort"
	"strings"
)

type Registry struct {
	resources              []Definition
	resourcesByKey         map[string]Definition
	resourcesByURI         map[string]Definition
	templates              []TemplateDefinition
	templatesByKey         map[string]TemplateDefinition
	templatesByURITemplate map[string]TemplateDefinition
}

var defaultRegistry = NewRegistry()

func NewRegistry() *Registry {
	return &Registry{
		resourcesByKey:         map[string]Definition{},
		resourcesByURI:         map[string]Definition{},
		templatesByKey:         map[string]TemplateDefinition{},
		templatesByURITemplate: map[string]TemplateDefinition{},
	}
}

func DefaultRegistry() *Registry {
	return defaultRegistry
}

func (r *Registry) Register(def Definition) {
	if def == nil {
		panic("resource definition is nil")
	}
	r.ensureIndexes()
	key := Key(def)
	if key == "" {
		panic(fmt.Sprintf("resource %T has empty key", def))
	}
	if _, ok := r.resourcesByKey[key]; ok {
		panic(fmt.Sprintf("duplicate resource %q", key))
	}
	if uri := strings.TrimSpace(def.URI()); uri != "" {
		r.resourcesByURI[uri] = def
	}
	r.resourcesByKey[key] = def
	r.resources = append(r.resources, def)
}

func (r *Registry) RegisterTemplate(def TemplateDefinition) {
	if def == nil {
		panic("resource template definition is nil")
	}
	r.ensureIndexes()
	key := TemplateKey(def)
	if key == "" {
		panic(fmt.Sprintf("resource template %T has empty key", def))
	}
	if _, ok := r.templatesByKey[key]; ok {
		panic(fmt.Sprintf("duplicate resource template %q", key))
	}
	r.templatesByKey[key] = def
	r.templatesByURITemplate[strings.TrimSpace(def.URITemplate())] = def
	r.templates = append(r.templates, def)
}

func (r *Registry) Resources() []Definition {
	out := append([]Definition(nil), r.resources...)
	sort.Slice(out, func(i, j int) bool {
		return Key(out[i]) < Key(out[j])
	})
	return out
}

func (r *Registry) ResourceTemplates() []TemplateDefinition {
	out := append([]TemplateDefinition(nil), r.templates...)
	sort.Slice(out, func(i, j int) bool {
		return TemplateKey(out[i]) < TemplateKey(out[j])
	})
	return out
}

func (r *Registry) ByURI(uri string) (Definition, bool) {
	uri = strings.TrimSpace(uri)
	if def, ok := r.resourcesByURI[uri]; ok {
		return def, true
	}
	if _, ok := r.resourcesByKey[SelectedArtifactID]; ok {
		id := ArtifactIDFromURI(uri)
		if id != "" {
			return NewSelectedArtifactResource(SelectedArtifact{ID: id}), true
		}
	}
	return nil, false
}

func (r *Registry) TemplateByURITemplate(uriTemplate string) (TemplateDefinition, bool) {
	uriTemplate = strings.TrimSpace(uriTemplate)
	def, ok := r.templatesByURITemplate[uriTemplate]
	return def, ok
}

func (r *Registry) ensureIndexes() {
	if r.resourcesByKey == nil {
		r.resourcesByKey = map[string]Definition{}
	}
	if r.resourcesByURI == nil {
		r.resourcesByURI = map[string]Definition{}
	}
	if r.templatesByKey == nil {
		r.templatesByKey = map[string]TemplateDefinition{}
	}
	if r.templatesByURITemplate == nil {
		r.templatesByURITemplate = map[string]TemplateDefinition{}
	}
}
