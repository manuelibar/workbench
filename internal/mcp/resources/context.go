package resources

import (
	"embed"
	"fmt"
)

//go:embed context.md
var contextFiles embed.FS

type ContextResource struct{}

func NewContextResource() *ContextResource {
	return &ContextResource{}
}

func ContextMarkdown() string {
	b, err := contextFiles.ReadFile("context.md")
	if err != nil {
		panic(fmt.Sprintf("embedded context resource is missing: %v", err))
	}
	return string(b)
}

func (r *ContextResource) URI() string {
	return ContextURI
}

func (r *ContextResource) Name() string {
	return "context"
}

func (r *ContextResource) Title() string {
	return "Workbench Context"
}

func (r *ContextResource) Description() string {
	return "Current raw Workbench context document."
}

func (r *ContextResource) MIMEType() string {
	return "text/markdown"
}

func (r *ContextResource) Group() string {
	return "core"
}

func (r *ContextResource) Visibility() Visibility {
	return VisibleAlways
}
