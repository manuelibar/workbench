package tools

type Visibility string

const (
	VisibleAlways           Visibility = "always"
	VisibleArtifactSelected Visibility = "artifact_selected"
)

type Definition interface {
	Name() string
	Description() *string
	Group() string
	Visibility() Visibility
	InputSchema() any
	OutputSchema() any
}

func Key(def Definition) string {
	return def.Name()
}

func All() []Definition {
	return []Definition{
		NewContextTool(),
		NewArtifactBeginTool(),
		NewArtifactListTool(),
		NewArtifactGetTool(),
		NewArtifactUpdateTool(),
		NewArtifactGuidanceTool(),
		NewArtifactValidateTool(),
	}
}

func ByName(name string) (Definition, bool) {
	for _, def := range All() {
		if def.Name() == name {
			return def, true
		}
	}
	return nil, false
}
