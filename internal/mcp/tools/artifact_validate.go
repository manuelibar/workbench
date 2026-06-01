package tools

import (
	wbjsonschema "github.com/manuelibar/workbench/internal/jsonschema"
	"github.com/manuelibar/workbench/internal/ptr"
)

const ArtifactValidateName = "artifact.validate"

type ArtifactValidateTool struct{}

func NewArtifactValidateTool() *ArtifactValidateTool {
	return &ArtifactValidateTool{}
}

func (t *ArtifactValidateTool) Name() string {
	return ArtifactValidateName
}

func (t *ArtifactValidateTool) Description() *string {
	return ptr.Ptr("Validate the selected artifact against its type contract.")
}

func (t *ArtifactValidateTool) Group() string {
	return "artifacts"
}

func (t *ArtifactValidateTool) Visibility() Visibility {
	return VisibleArtifactSelected
}

func (t *ArtifactValidateTool) InputSchema() any {
	return wbjsonschema.StrictObject(map[string]*wbjsonschema.Schema{
		"artifact_id": wbjsonschema.String(),
	})
}

func (t *ArtifactValidateTool) OutputSchema() any {
	return nil
}
