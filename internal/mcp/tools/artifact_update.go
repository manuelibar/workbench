package tools

import (
	wbjsonschema "github.com/manuelibar/workbench/internal/jsonschema"
	"github.com/manuelibar/workbench/internal/ptr"
)

const ArtifactUpdateName = "artifact.update"

type ArtifactUpdateTool struct{}

func NewArtifactUpdateTool() *ArtifactUpdateTool {
	return &ArtifactUpdateTool{}
}

func (t *ArtifactUpdateTool) Name() string {
	return ArtifactUpdateName
}

func (t *ArtifactUpdateTool) Description() *string {
	return ptr.Ptr("Update selected artifact metadata or section bodies.")
}

func (t *ArtifactUpdateTool) Group() string {
	return "artifacts"
}

func (t *ArtifactUpdateTool) Visibility() Visibility {
	return VisibleArtifactSelected
}

func (t *ArtifactUpdateTool) InputSchema() any {
	return wbjsonschema.StrictObject(map[string]*wbjsonschema.Schema{
		"artifact_id":   wbjsonschema.String(),
		"title":         wbjsonschema.String(),
		"status":        wbjsonschema.String(),
		"set_sections":  wbjsonschema.StringMap(),
		"clear_section": wbjsonschema.ArrayOf(wbjsonschema.String()),
	})
}

func (t *ArtifactUpdateTool) OutputSchema() any {
	return nil
}
