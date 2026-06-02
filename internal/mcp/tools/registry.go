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

var registry []Definition

func register[In, Out any](impl implementation[In, Out]) {
	tool := typedTool[In, Out]{impl: impl}
	if tool.FullName() == "" {
		panic(fmt.Sprintf("tool %T has empty name", impl))
	}
	for _, existing := range registry {
		if existing.FullName() == tool.FullName() {
			panic(fmt.Sprintf("duplicate tool %q", tool.FullName()))
		}
	}
	registry = append(registry, tool)
}

func Registered() []Definition {
	out := append([]Definition(nil), registry...)
	sort.Slice(out, func(i, j int) bool {
		return out[i].FullName() < out[j].FullName()
	})
	return out
}

func ByName(name string) (Definition, bool) {
	name = strings.TrimSpace(name)
	for _, tool := range registry {
		if tool.FullName() == name {
			return tool, true
		}
	}
	return nil, false
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
