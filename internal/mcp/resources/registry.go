package resources

import (
	"fmt"
	"sort"
	"strings"
)

var registry []Definition
var templateRegistry []TemplateDefinition

func register(def Definition) {
	key := Key(def)
	if key == "" {
		panic(fmt.Sprintf("resource %T has empty key", def))
	}
	for _, existing := range registry {
		if Key(existing) == key {
			panic(fmt.Sprintf("duplicate resource %q", key))
		}
	}
	registry = append(registry, def)
}

func registerTemplate(def TemplateDefinition) {
	key := TemplateKey(def)
	if key == "" {
		panic(fmt.Sprintf("resource template %T has empty key", def))
	}
	for _, existing := range templateRegistry {
		if TemplateKey(existing) == key {
			panic(fmt.Sprintf("duplicate resource template %q", key))
		}
	}
	templateRegistry = append(templateRegistry, def)
}

func Registered() []Definition {
	out := append([]Definition(nil), registry...)
	sort.Slice(out, func(i, j int) bool {
		return Key(out[i]) < Key(out[j])
	})
	return out
}

func RegisteredTemplates() []TemplateDefinition {
	out := append([]TemplateDefinition(nil), templateRegistry...)
	sort.Slice(out, func(i, j int) bool {
		return TemplateKey(out[i]) < TemplateKey(out[j])
	})
	return out
}

func ByURI(uri string) (Definition, bool) {
	uri = strings.TrimSpace(uri)
	for _, def := range registry {
		if def.URI() == uri {
			return def, true
		}
		if Key(def) == SelectedArtifactID {
			id := ArtifactIDFromURI(uri)
			if id != "" {
				return NewSelectedArtifactResource(SelectedArtifact{ID: id}), true
			}
		}
	}
	return nil, false
}

func TemplateByURITemplate(uriTemplate string) (TemplateDefinition, bool) {
	uriTemplate = strings.TrimSpace(uriTemplate)
	for _, def := range templateRegistry {
		if def.URITemplate() == uriTemplate {
			return def, true
		}
	}
	return nil, false
}

func All() []Definition {
	return Registered()
}

func Templates() []TemplateDefinition {
	return RegisteredTemplates()
}
