package mcp

import (
	"context"
	"fmt"
	"sort"
	"strings"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type toolImplementation[In, Out any] interface {
	Name() string
	Group() string
	Description() string
	Handle(context.Context, *Server, In) (Out, error)
}

type registeredTool interface {
	Name() string
	Group() string
	Description() string
	FullName() string
	addTo(*Server)
}

type typedTool[In, Out any] struct {
	impl toolImplementation[In, Out]
}

var toolRegistry []registeredTool

func registerTool[In, Out any](impl toolImplementation[In, Out]) {
	tool := typedTool[In, Out]{impl: impl}
	if tool.FullName() == "" {
		panic(fmt.Sprintf("tool %T has empty name", impl))
	}
	for _, existing := range toolRegistry {
		if existing.FullName() == tool.FullName() {
			panic(fmt.Sprintf("duplicate tool %q", tool.FullName()))
		}
	}
	toolRegistry = append(toolRegistry, tool)
}

func registeredTools() []registeredTool {
	out := append([]registeredTool(nil), toolRegistry...)
	sort.Slice(out, func(i, j int) bool {
		return out[i].FullName() < out[j].FullName()
	})
	return out
}

func registeredToolByName(name string) (registeredTool, bool) {
	name = strings.TrimSpace(name)
	for _, tool := range toolRegistry {
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

func (t typedTool[In, Out]) addTo(s *Server) {
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        t.FullName(),
		Description: t.Description(),
	}, func(ctx context.Context, _ *mcpsdk.CallToolRequest, in In) (*mcpsdk.CallToolResult, Out, error) {
		out, err := t.impl.Handle(ctx, s, in)
		return nil, out, err
	})
}
