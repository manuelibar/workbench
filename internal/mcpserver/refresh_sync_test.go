package mcpserver_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestRefreshFallsBackWithCapabilityIndexWhenClientDoesNotRelist(t *testing.T) {
	ctx := context.Background()
	srv := newTestServer(t)
	srv.SetRefreshListWait(10 * time.Millisecond)
	cs := connect(ctx, t, srv)

	result := callTool(ctx, t, cs, "refresh", nil)
	sync := result["capability_sync"].(map[string]any)
	if sync["status"] != "fallback_index" {
		t.Fatalf("expected fallback_index status, got %v", sync)
	}
	if sync["timed_out"] != true {
		t.Fatalf("expected timed_out=true, got %v", sync)
	}
	if _, ok := sync["index"].(map[string]any); !ok {
		t.Fatalf("expected inline capability index after timeout, got %v", sync)
	}
}

func TestRefreshWaitsForToolAndResourceRelistBeforeReturningTidyResult(t *testing.T) {
	ctx := context.Background()
	srv := newTestServer(t)
	srv.SetRefreshListWait(500 * time.Millisecond)
	cs := connect(ctx, t, srv)

	resultCh := make(chan map[string]any, 1)
	errCh := make(chan error, 1)
	go func() {
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "refresh", Arguments: map[string]any{}})
		if err != nil {
			errCh <- err
			return
		}
		var out map[string]any
		if err := json.Unmarshal([]byte(res.Content[0].(*mcp.TextContent).Text), &out); err != nil {
			errCh <- err
			return
		}
		resultCh <- out
	}()

	time.Sleep(50 * time.Millisecond)
	if _, err := cs.ListTools(ctx, nil); err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	if _, err := cs.ListResources(ctx, nil); err != nil {
		t.Fatalf("ListResources: %v", err)
	}

	select {
	case err := <-errCh:
		t.Fatalf("refresh failed: %v", err)
	case result := <-resultCh:
		sync := result["capability_sync"].(map[string]any)
		if sync["status"] != "client_relisted" {
			t.Fatalf("expected client_relisted status, got %v", sync)
		}
		if _, ok := sync["index"]; ok {
			t.Fatalf("did not expect inline index after observed relist, got %v", sync)
		}
	case <-time.After(time.Second):
		t.Fatal("refresh did not return after list calls")
	}
}
