package resources

import (
	"embed"
	"fmt"
)

//go:embed scope.md
var scopeFiles embed.FS

type ScopeResource struct{}

func init() {
	register(NewScopeResource())
}

func NewScopeResource() *ScopeResource {
	return &ScopeResource{}
}

func ScopeMarkdown() string {
	b, err := scopeFiles.ReadFile("scope.md")
	if err != nil {
		panic(fmt.Sprintf("embedded scope resource is missing: %v", err))
	}
	return string(b)
}

func (r *ScopeResource) URI() string {
	return ScopeURI
}

func (r *ScopeResource) Name() string {
	return "scope"
}

func (r *ScopeResource) Title() string {
	return "Workbench Scope"
}

func (r *ScopeResource) Description() string {
	return "Current raw Workbench scope document."
}

func (r *ScopeResource) MIMEType() string {
	return "text/markdown"
}

func (r *ScopeResource) Group() string {
	return "core"
}

func (r *ScopeResource) Visibility() Visibility {
	return VisibleAlways
}
