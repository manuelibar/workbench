package tools

import (
	wbjsonschema "github.com/manuelibar/workbench/internal/jsonschema"
	"github.com/manuelibar/workbench/internal/ptr"
)

const ArtifactGuidanceName = "artifact.guidance"

type ArtifactGuidanceTool struct{}

func NewArtifactGuidanceTool() *ArtifactGuidanceTool {
	return &ArtifactGuidanceTool{}
}

func (t *ArtifactGuidanceTool) Name() string {
	return ArtifactGuidanceName
}

func (t *ArtifactGuidanceTool) Description() *string {
	return ptr.Ptr("Return deterministic contract guidance for the selected artifact.")
}

func (t *ArtifactGuidanceTool) Group() string {
	return "artifacts"
}

func (t *ArtifactGuidanceTool) Visibility() Visibility {
	return VisibleArtifactSelected
}

func (t *ArtifactGuidanceTool) InputSchema() any {
	return wbjsonschema.StrictObject(map[string]*wbjsonschema.Schema{
		"artifact_id": wbjsonschema.String(),
		"type":        wbjsonschema.String(),
	})
}

func (t *ArtifactGuidanceTool) OutputSchema() any {
	return nil
}
