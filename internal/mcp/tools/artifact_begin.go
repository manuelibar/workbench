package tools

import (
	wbjsonschema "github.com/manuelibar/workbench/internal/jsonschema"
	"github.com/manuelibar/workbench/internal/ptr"
)

const ArtifactBeginName = "artifact.begin"

type ArtifactBeginTool struct{}

func NewArtifactBeginTool() *ArtifactBeginTool {
	return &ArtifactBeginTool{}
}

func (t *ArtifactBeginTool) Name() string {
	return ArtifactBeginName
}

func (t *ArtifactBeginTool) Description() *string {
	return ptr.Ptr("Create a typed Markdown artifact draft under docs/artifacts.")
}

func (t *ArtifactBeginTool) Group() string {
	return "artifacts"
}

func (t *ArtifactBeginTool) Visibility() Visibility {
	return VisibleAlways
}

func (t *ArtifactBeginTool) InputSchema() any {
	return wbjsonschema.StrictObject(map[string]*wbjsonschema.Schema{
		"type":   wbjsonschema.String(),
		"title":  wbjsonschema.String(),
		"status": wbjsonschema.String(),
		"focus":  wbjsonschema.String(),
		"select": wbjsonschema.Bool(),
	}, "type", "title")
}

func (t *ArtifactBeginTool) OutputSchema() any {
	return nil
}
