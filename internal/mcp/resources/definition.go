package resources

import "strings"

type Visibility string

const (
	VisibleAlways           Visibility = "always"
	VisibleArtifactSelected Visibility = "artifact_selected"

	SelectedArtifactID = "resource.artifact.selected"

	ContextURI = "workbench:///context"
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

func Key(def Definition) string {
	if _, ok := def.(*SelectedArtifactResource); ok {
		return SelectedArtifactID
	}
	if def.URI() != "" {
		return def.URI()
	}
	return SelectedArtifactID
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
