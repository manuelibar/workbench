package tools

import (
	wbjsonschema "github.com/manuelibar/workbench/internal/jsonschema"
	"github.com/manuelibar/workbench/internal/ptr"
)

const ArtifactGetName = "artifact.get"

type ArtifactGetTool struct{}

func NewArtifactGetTool() *ArtifactGetTool {
	return &ArtifactGetTool{}
}

func (t *ArtifactGetTool) Name() string {
	return ArtifactGetName
}

func (t *ArtifactGetTool) Description() *string {
	return ptr.Ptr("Read one artifact by stable id.")
}

func (t *ArtifactGetTool) Group() string {
	return "artifacts"
}

func (t *ArtifactGetTool) Visibility() Visibility {
	return VisibleAlways
}

func (t *ArtifactGetTool) InputSchema() any {
	return wbjsonschema.StrictObject(map[string]*wbjsonschema.Schema{
		"artifact_id": wbjsonschema.String(),
	}, "artifact_id")
}

func (t *ArtifactGetTool) OutputSchema() any {
	return nil
}
