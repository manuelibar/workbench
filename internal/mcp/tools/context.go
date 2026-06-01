package tools

import (
	wbjsonschema "github.com/manuelibar/workbench/internal/jsonschema"
	"github.com/manuelibar/workbench/internal/ptr"
)

const ContextName = "context"

type ContextTool struct{}

func NewContextTool() *ContextTool {
	return &ContextTool{}
}

func (t *ContextTool) Name() string {
	return ContextName
}

func (t *ContextTool) Description() *string {
	return ptr.Ptr("Read or patch focus/artifact context and return raw context plus MCP list sync status.")
}

func (t *ContextTool) Group() string {
	return "core"
}

func (t *ContextTool) Visibility() Visibility {
	return VisibleAlways
}

func (t *ContextTool) InputSchema() any {
	return wbjsonschema.StrictObject(map[string]*wbjsonschema.Schema{
		"focus":       wbjsonschema.NullableString(),
		"artifact_id": wbjsonschema.NullableString(),
	})
}

func (t *ContextTool) OutputSchema() any {
	return nil
}
