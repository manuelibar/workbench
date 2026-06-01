package tools

import (
	wbjsonschema "github.com/manuelibar/workbench/internal/jsonschema"
	"github.com/manuelibar/workbench/internal/ptr"
)

const ArtifactListName = "artifact.list"

type ArtifactListTool struct{}

func NewArtifactListTool() *ArtifactListTool {
	return &ArtifactListTool{}
}

func (t *ArtifactListTool) Name() string {
	return ArtifactListName
}

func (t *ArtifactListTool) Description() *string {
	return ptr.Ptr("List file-backed artifacts in docs/artifacts.")
}

func (t *ArtifactListTool) Group() string {
	return "artifacts"
}

func (t *ArtifactListTool) Visibility() Visibility {
	return VisibleAlways
}

func (t *ArtifactListTool) InputSchema() any {
	return wbjsonschema.EmptyObject()
}

func (t *ArtifactListTool) OutputSchema() any {
	return nil
}
