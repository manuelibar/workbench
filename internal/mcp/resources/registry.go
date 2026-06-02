package resources

import (
	"fmt"
	"sort"
	"strings"
)

type Registry struct {
	resources      []Definition
	resourcesByKey map[string]Definition
	resourcesByURI map[string]Definition
}

var defaultRegistry = NewRegistry()

func NewRegistry() *Registry {
	return &Registry{
		resourcesByKey: map[string]Definition{},
		resourcesByURI: map[string]Definition{},
	}
}

func DefaultRegistry() *Registry {
	return defaultRegistry
}

func register(def Definition) {
	defaultRegistry.register(def)
}

func (r *Registry) register(def Definition) {
	if def == nil {
		panic("resource definition is nil")
	}
	key := Key(def)
	if key == "" {
		panic(fmt.Sprintf("resource %T has empty key", def))
	}
	if r.resourcesByKey == nil {
		r.resourcesByKey = map[string]Definition{}
	}
	if r.resourcesByURI == nil {
		r.resourcesByURI = map[string]Definition{}
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

func (r *Registry) Resources() []Definition {
	out := append([]Definition(nil), r.resources...)
	sort.Slice(out, func(i, j int) bool {
		return Key(out[i]) < Key(out[j])
	})
	return out
}

func (r *Registry) ByURI(uri string) (Definition, bool) {
	uri = strings.TrimSpace(uri)
	if def, ok := r.resourcesByURI[uri]; ok {
		return def, true
	}
	if _, ok := r.resourcesByKey[ArtifactResourceID]; ok {
		id := ArtifactIDFromURI(uri)
		if id != "" {
			return NewArtifactResource(Artifact{ID: id}), true
		}
	}
	return nil, false
}
