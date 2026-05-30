package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"slices"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	kernel "github.com/manuelibar/workbench/internal/mcpserver"
)

func TestMain(m *testing.M) {
	if os.Getenv("WORKBENCH_TEST_STDIO_SERVER") == "1" {
		os.Exit(run())
	}
	os.Exit(m.Run())
}

func TestStdioMCPContextIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	toolsChanged := make(chan struct{}, 1)
	resourcesChanged := make(chan struct{}, 1)
	client := mcp.NewClient(&mcp.Implementation{Name: "stdio-test-client"}, &mcp.ClientOptions{
		ToolListChangedHandler: func(context.Context, *mcp.ToolListChangedRequest) {
			toolsChanged <- struct{}{}
		},
		ResourceListChangedHandler: func(context.Context, *mcp.ResourceListChangedRequest) {
			resourcesChanged <- struct{}{}
		},
	})
	cmd := exec.Command(os.Args[0])
	cmd.Env = append(os.Environ(),
		"WORKBENCH_TEST_STDIO_SERVER=1",
		"WORKBENCH_ARTIFACT_DIR="+t.TempDir(),
		"WORKBENCH_CONTEXT_SYNC_TIMEOUT=2s",
		"WORKBENCH_LOG_LEVEL=error",
	)
	session, err := client.Connect(ctx, &mcp.CommandTransport{
		Command:           cmd,
		TerminateDuration: time.Second,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	if _, err := session.ListTools(ctx, nil); err != nil {
		t.Fatal(err)
	}
	resultCh := make(chan *mcp.CallToolResult, 1)
	errCh := make(chan error, 1)
	go func() {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name: "artifact.begin",
			Arguments: map[string]any{
				"type":   "rfc",
				"title":  "Stdio sync RFC",
				"focus":  "prove stdio relist sync",
				"select": true,
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
	if _, err := session.ListTools(ctx, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := session.ListResources(ctx, nil); err != nil {
		t.Fatal(err)
	}

	var toolResult *mcp.CallToolResult
	select {
	case err := <-errCh:
		t.Fatal(err)
	case toolResult = <-resultCh:
	case <-ctx.Done():
		t.Fatal("artifact.begin did not return")
	}
	begin := decodeStructured[kernel.ArtifactBeginResult](t, toolResult)
	if begin.Context == nil {
		t.Fatal("artifact.begin(select=true) did not return context")
	}
	if begin.Context.Sync.TimedOut {
		t.Fatalf("sync timed out after relists: %+v", begin.Context.Sync)
	}
	if !slices.Contains(begin.Context.Sync.Observed, "tools/list") ||
		!slices.Contains(begin.Context.Sync.Observed, "resources/list") {
		t.Fatalf("observed = %v, want tools/resources", begin.Context.Sync.Observed)
	}
	resource, err := session.ReadResource(ctx, &mcp.ReadResourceParams{URI: "workbench:///context"})
	if err != nil {
		t.Fatal(err)
	}
	if got := resource.Contents[0].Text; got != begin.Context.ContextDocument {
		t.Fatalf("context resource mismatch\nresource:\n%s\ntool:\n%s", got, begin.Context.ContextDocument)
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
