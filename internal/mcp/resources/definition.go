package resources

import "strings"

type Visibility string

const (
	VisibleAlways         Visibility = "always"
	VisibleArtifactScoped Visibility = "artifact_scoped"

	ArtifactResourceID = "resource.artifact.scoped"

	ScopeURI = "workbench:///scope"
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
	if _, ok := def.(*ArtifactResource); ok {
		return ArtifactResourceID
	}
	if def.URI() != "" {
		return def.URI()
	}
	return ArtifactResourceID
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
