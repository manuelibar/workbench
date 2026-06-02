package tools

import (
	"context"
	"fmt"
	"sort"
	"strings"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/manuelibar/workbench/internal/artifacts"
)

type Runtime interface {
	ArtifactStore() *artifacts.Store
	ApplyContextPatch(context.Context, map[string]any) (ContextualizeResult, error)
	SelectArtifact(context.Context, string, string) (ContextualizeResult, error)
	ResolveArtifactID(context.Context, string) (string, error)
	RefreshSelectedArtifactResource(artifacts.Summary)
}

type implementation[In, Out any] interface {
	Name() string
	Group() string
	Description() string
	Handle(context.Context, Runtime, In) (Out, error)
}

type Definition interface {
	Name() string
	Group() string
	Description() string
	FullName() string
	AddTo(*mcpsdk.Server, Runtime)
}

type typedTool[In, Out any] struct {
	impl implementation[In, Out]
}

type Registry struct {
	tools       []Definition
	toolsByName map[string]Definition
}

var defaultRegistry = NewRegistry()

func NewRegistry() *Registry {
	return &Registry{toolsByName: map[string]Definition{}}
}

func DefaultRegistry() *Registry {
	return defaultRegistry
}

func (r *Registry) Register(tool Definition) {
	if tool == nil {
		panic("tool definition is nil")
	}
	r.ensureIndexes()
	name := tool.FullName()
	if name == "" {
		panic(fmt.Sprintf("tool %T has empty name", tool))
	}
	if _, ok := r.toolsByName[name]; ok {
		panic(fmt.Sprintf("duplicate tool %q", name))
	}
	r.toolsByName[name] = tool
	r.tools = append(r.tools, tool)
}

func (r *Registry) Tools() []Definition {
	out := append([]Definition(nil), r.tools...)
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

func (r *Registry) ensureIndexes() {
	if r.toolsByName != nil {
		return
	}
	r.toolsByName = map[string]Definition{}
}

func (t typedTool[In, Out]) Name() string {
	return strings.TrimSpace(t.impl.Name())
}

func (t typedTool[In, Out]) Group() string {
	return strings.Trim(strings.TrimSpace(t.impl.Group()), ".")
}

func (t typedTool[In, Out]) Description() string {
	return strings.TrimSpace(t.impl.Description())
}

func (t typedTool[In, Out]) FullName() string {
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

func (t typedTool[In, Out]) AddTo(s *mcpsdk.Server, runtime Runtime) {
	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        t.FullName(),
		Description: t.Description(),
	}, func(ctx context.Context, _ *mcpsdk.CallToolRequest, in In) (*mcpsdk.CallToolResult, Out, error) {
		out, err := t.impl.Handle(ctx, runtime, in)
		return nil, out, err
	})
}
