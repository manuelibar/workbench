package tools

import (
	"context"
	"fmt"
	"sort"
	"strings"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/manuelibar/workbench/internal/artifacts"
)

type Host interface {
	ArtifactStore() *artifacts.Store
	ApplyContextPatch(context.Context, map[string]any) (ContextualizeResult, error)
	SelectArtifact(context.Context, string, string) (ContextualizeResult, error)
	ResolveArtifactID(context.Context, string) (string, error)
	RefreshSelectedArtifactResource(artifacts.Summary)
}

type Handler[In, Out any] interface {
	Name() string
	Group() string
	Description() string
	Handle(context.Context, Host, In) (Out, error)
}

type Definition interface {
	Name() string
	Group() string
	Description() string
	FullName() string
}

type bindableDefinition interface {
	Definition
	bind(*mcpsdk.Server, Host)
}

type toolDefinition[In, Out any] struct {
	handler Handler[In, Out]
}

type Registry struct {
	tools       []bindableDefinition
	toolsByName map[string]bindableDefinition
}

var defaultRegistry = NewRegistry()

func NewRegistry() *Registry {
	return &Registry{toolsByName: map[string]bindableDefinition{}}
}

func DefaultRegistry() *Registry {
	return defaultRegistry
}

func register[In, Out any](handler Handler[In, Out]) {
	defaultRegistry.register(toolDefinition[In, Out]{handler: handler})
}

func (r *Registry) register(tool bindableDefinition) {
	if tool == nil {
		panic("tool definition is nil")
	}
	name := tool.FullName()
	if name == "" {
		panic(fmt.Sprintf("tool %T has empty name", tool))
	}
	if r.toolsByName == nil {
		r.toolsByName = map[string]bindableDefinition{}
	}
	if _, ok := r.toolsByName[name]; ok {
		panic(fmt.Sprintf("duplicate tool %q", name))
	}
	r.toolsByName[name] = tool
	r.tools = append(r.tools, tool)
}

func (r *Registry) Tools() []Definition {
	out := make([]Definition, 0, len(r.tools))
	for _, tool := range r.tools {
		out = append(out, tool)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].FullName() < out[j].FullName()
	})
	return out
}

func (r *Registry) ByName(name string) (Definition, bool) {
	name = strings.TrimSpace(name)
	tool, ok := r.toolsByName[name]
	return tool, ok
}

func Bind(def Definition, s *mcpsdk.Server, host Host) {
	tool, ok := def.(bindableDefinition)
	if !ok {
		panic(fmt.Sprintf("tool definition %T cannot bind to MCP SDK server", def))
	}
	tool.bind(s, host)
}

func (t toolDefinition[In, Out]) Name() string {
	return strings.TrimSpace(t.handler.Name())
}

func (t toolDefinition[In, Out]) Group() string {
	return strings.Trim(strings.TrimSpace(t.handler.Group()), ".")
}

func (t toolDefinition[In, Out]) Description() string {
	return strings.TrimSpace(t.handler.Description())
}

func (t toolDefinition[In, Out]) FullName() string {
	group := t.Group()
	name := t.Name()
	if group == "" {
		return name
	}
	if name == "" {
		return group
	}
	return group + "." + name
}

func (t toolDefinition[In, Out]) bind(s *mcpsdk.Server, host Host) {
	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        t.FullName(),
		Description: t.Description(),
	}, func(ctx context.Context, _ *mcpsdk.CallToolRequest, in In) (*mcpsdk.CallToolResult, Out, error) {
		out, err := t.handler.Handle(ctx, host, in)
		return nil, out, err
	})
}
