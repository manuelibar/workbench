package resources

import "strings"

type Visibility string

const (
	VisibleAlways           Visibility = "always"
	VisibleArtifactSelected Visibility = "artifact_selected"

	SelectedArtifactID = "resource.artifact.selected"

	ContextURI          = "workbench:///context"
	ArtifactTemplateURI = "workbench:///artifacts/{id}"
)

type Definition interface {
	URI() string
	Name() string
	Title() string
	Description() string
	MIMEType() string
	Group() string
	Visibility() Visibility
}

type TemplateDefinition interface {
	URITemplate() string
	Name() string
	Title() string
	Description() string
	MIMEType() string
	Group() string
	Visibility() Visibility
}

func Key(def Definition) string {
	if def.URI() != "" {
		return def.URI()
	}
	return SelectedArtifactID
}

func TemplateKey(def TemplateDefinition) string {
	return def.URITemplate()
}

func All() []Definition {
	return []Definition{
		NewContextResource(),
		NewSelectedArtifactResource(SelectedArtifact{}),
	}
}

func Templates() []TemplateDefinition {
	return []TemplateDefinition{
		NewArtifactTemplate(),
	}
}

func ArtifactURI(id string) string {
	return "workbench:///artifacts/" + id
}

func ArtifactIDFromURI(uri string) string {
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
