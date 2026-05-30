package mcpserver

import (
	"context"
	"encoding/json"
	"slices"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMCPContextRelistAndResources(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server, err := New(Options{ArtifactDir: t.TempDir(), SyncTimeout: 2 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	artifact, err := server.artifacts.Begin(BeginArtifactRequest{Type: "rfc", Title: "Relist RFC"})
	if err != nil {
		t.Fatal(err)
	}

	toolsChanged := make(chan struct{}, 1)
	resourcesChanged := make(chan struct{}, 1)
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, &mcp.ClientOptions{
		ToolListChangedHandler: func(context.Context, *mcp.ToolListChangedRequest) {
			toolsChanged <- struct{}{}
		},
		ResourceListChangedHandler: func(context.Context, *mcp.ResourceListChangedRequest) {
			resourcesChanged <- struct{}{}
		},
	})
	ct, st := mcp.NewInMemoryTransports()
	ss, err := server.SDKServer().Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ss.Close()
	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cs.Close()

	if _, err := cs.ListTools(ctx, nil); err != nil {
		t.Fatal(err)
	}

	resultCh := make(chan *mcp.CallToolResult, 1)
	errCh := make(chan error, 1)
	go func() {
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "context",
			Arguments: map[string]any{
				"focus":       "verify relist",
				"artifact_id": artifact.ID,
			},
		})
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- res
	}()

	select {
	case <-toolsChanged:
	case <-ctx.Done():
		t.Fatal("timed out waiting for tools list_changed")
	}
	select {
	case <-resourcesChanged:
	case <-ctx.Done():
		t.Fatal("timed out waiting for resources list_changed")
	}
	tools, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !toolListed(tools, "artifact.validate") {
		t.Fatalf("artifact.validate not listed after selection: %v", toolNames(tools))
	}
	resources, err := cs.ListResources(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !resourceListed(resources, "workbench:///artifacts/"+artifact.ID) {
		t.Fatalf("selected artifact resource not listed: %+v", resources.Resources)
	}

	var toolResult *mcp.CallToolResult
	select {
	case err := <-errCh:
		t.Fatal(err)
	case toolResult = <-resultCh:
	case <-ctx.Done():
		t.Fatal("context call did not return")
	}
	contextResult := decodeStructured[ContextResult](t, toolResult)
	if contextResult.Sync.TimedOut {
		t.Fatalf("context timed out after relists: %+v", contextResult.Sync)
	}
	if !slices.Contains(contextResult.Sync.Observed, methodListTools) ||
		!slices.Contains(contextResult.Sync.Observed, methodListResources) {
		t.Fatalf("observed = %v, want tools/resources", contextResult.Sync.Observed)
	}

	contextResource, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: "workbench:///context"})
	if err != nil {
		t.Fatal(err)
	}
	if got := contextResource.Contents[0].Text; got != contextResult.ContextDocument {
		t.Fatalf("context resource mismatch\nresource:\n%s\ntool:\n%s", got, contextResult.ContextDocument)
	}
	artifactResource, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: "workbench:///artifacts/" + artifact.ID})
	if err != nil {
		t.Fatal(err)
	}
	if got := artifactResource.Contents[0].Text; got != artifact.Markdown {
		t.Fatalf("artifact resource mismatch\nresource:\n%s\ndisk:\n%s", got, artifact.Markdown)
	}
}

func decodeStructured[T any](t *testing.T, res *mcp.CallToolResult) T {
	t.Helper()
	raw, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("decode structured content: %v\n%s", err, raw)
	}
	return out
}

func toolListed(result *mcp.ListToolsResult, name string) bool {
	for _, tool := range result.Tools {
		if tool.Name == name {
			return true
		}
	}
	return false
}

func toolNames(result *mcp.ListToolsResult) []string {
	names := make([]string, 0, len(result.Tools))
	for _, tool := range result.Tools {
		names = append(names, tool.Name)
	}
	return names
}

func resourceListed(result *mcp.ListResourcesResult, uri string) bool {
	for _, resource := range result.Resources {
		if resource.URI == uri {
			return true
		}
	}
	return false
}
