package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/manuelibar/workbench/internal/artifacts"
	"github.com/manuelibar/workbench/internal/errs"
	mcptools "github.com/manuelibar/workbench/internal/mcp/tools"
)

func TestMCPContextRelistAndResources(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server, err := New(Options{ArtifactDir: t.TempDir(), SyncTimeout: 2 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	artifact, err := server.artifacts.Begin(artifacts.BeginRequest{Type: "rfc", Title: "Relist RFC"})
	if err != nil {
		t.Fatal(err)
	}

	toolsChanged := make(chan struct{}, 1)
	resourcesChanged := make(chan struct{}, 1)
	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test-client"}, &mcpsdk.ClientOptions{
		ToolListChangedHandler: func(context.Context, *mcpsdk.ToolListChangedRequest) {
			toolsChanged <- struct{}{}
		},
		ResourceListChangedHandler: func(context.Context, *mcpsdk.ResourceListChangedRequest) {
			resourcesChanged <- struct{}{}
		},
	})
	ct, st := mcpsdk.NewInMemoryTransports()
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

	resultCh := make(chan *mcpsdk.CallToolResult, 1)
	errCh := make(chan error, 1)
	go func() {
		res, err := cs.CallTool(ctx, &mcpsdk.CallToolParams{
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

	var toolResult *mcpsdk.CallToolResult
	select {
	case err := <-errCh:
		t.Fatal(err)
	case toolResult = <-resultCh:
	case <-ctx.Done():
		t.Fatal("context call did not return")
	}
	contextResult := decodeStructured[mcptools.ContextResult](t, toolResult)
	if contextResult.Sync.TimedOut {
		t.Fatalf("context timed out after relists: %+v", contextResult.Sync)
	}
	if !slices.Contains(contextResult.Sync.Observed, methodListTools) ||
		!slices.Contains(contextResult.Sync.Observed, methodListResources) {
		t.Fatalf("observed = %v, want tools/resources", contextResult.Sync.Observed)
	}

	contextResource, err := cs.ReadResource(ctx, &mcpsdk.ReadResourceParams{URI: "workbench:///context"})
	if err != nil {
		t.Fatal(err)
	}
	if got := contextResource.Contents[0].Text; got != contextResult.ContextDocument {
		t.Fatalf("context resource mismatch\nresource:\n%s\ntool:\n%s", got, contextResult.ContextDocument)
	}
	artifactResource, err := cs.ReadResource(ctx, &mcpsdk.ReadResourceParams{URI: "workbench:///artifacts/" + artifact.ID})
	if err != nil {
		t.Fatal(err)
	}
	if got := artifactResource.Contents[0].Text; got != artifact.Markdown {
		t.Fatalf("artifact resource mismatch\nresource:\n%s\ndisk:\n%s", got, artifact.Markdown)
	}
}

func TestMCPBoundarySanitizesClassifiedToolErrors(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server, session, cleanup := mcpTestSession(t, ctx)
	defer cleanup()

	result, err := session.CallTool(ctx, &mcpsdk.CallToolParams{
		Name: "artifact.get",
		Arguments: map[string]any{
			"artifact_id": "missing-artifact",
		},
	})
	if err != nil {
		t.Fatalf("CallTool returned protocol error: %v", err)
	}
	if !result.IsError {
		t.Fatal("CallTool IsError=false, want true")
	}
	text := result.Content[0].(*mcpsdk.TextContent).Text
	if text != "Artifact not found" {
		t.Fatalf("tool error text = %q", text)
	}
	if strings.Contains(text, "missing-artifact") || strings.Contains(text, server.artifacts.Dir()) {
		t.Fatalf("tool error text leaked private detail: %q", text)
	}

	var structured struct {
		Error struct {
			Title     string `json:"title"`
			Code      string `json:"code"`
			Retryable bool   `json:"retryable"`
		} `json:"error"`
	}
	decodeStructuredInto(t, result, &structured)
	if structured.Error.Title != "Artifact not found" {
		t.Fatalf("structured title = %q", structured.Error.Title)
	}
	if structured.Error.Code != artifacts.CodeNotFound.String() {
		t.Fatalf("structured code = %q", structured.Error.Code)
	}
	if structured.Error.Retryable {
		t.Fatal("structured retryable = true")
	}
}

func TestMCPBoundaryLeavesSDKValidationToolErrorsUnchanged(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, session, cleanup := mcpTestSession(t, ctx)
	defer cleanup()

	result, err := session.CallTool(ctx, &mcpsdk.CallToolParams{
		Name: "artifact.get",
		Arguments: map[string]any{
			"artifact_id": 42,
		},
	})
	if err != nil {
		t.Fatalf("CallTool returned protocol error: %v", err)
	}
	if !result.IsError {
		t.Fatal("CallTool IsError=false, want true")
	}
	if result.StructuredContent != nil {
		raw, _ := json.Marshal(result.StructuredContent)
		t.Fatalf("SDK validation error got structured content: %s", raw)
	}
	text := result.Content[0].(*mcpsdk.TextContent).Text
	if !strings.Contains(text, "artifact_id") {
		t.Fatalf("SDK validation text = %q, want artifact_id mention", text)
	}
	if strings.Contains(text, artifacts.CodeNotFound.String()) ||
		strings.Contains(text, "Artifact not found") {
		t.Fatalf("SDK validation text was sanitized as a classified error: %q", text)
	}
}

func TestMCPBoundaryMapsClassifiedResourceErrors(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server, session, cleanup := mcpTestSession(t, ctx)
	defer cleanup()

	_, err := session.ReadResource(ctx, &mcpsdk.ReadResourceParams{URI: "workbench:///artifacts/missing-artifact"})
	assertJSONRPCError(t, err, mcpsdk.CodeResourceNotFound, "Artifact not found")

	server.SDKServer().AddResource(&mcpsdk.Resource{
		URI:      "workbench:///invalid-resource",
		Name:     "invalid-resource",
		MIMEType: "text/plain",
	}, func(context.Context, *mcpsdk.ReadResourceRequest) (*mcpsdk.ReadResourceResult, error) {
		return nil, errs.New(
			"Resource URI is invalid",
			errs.WithSentinel(errs.ErrInvalid),
			errs.WithCode(errCodeResourceURIInvalid),
			errs.WithSeverity(errs.SeverityWarning),
		)
	})
	_, err = session.ReadResource(ctx, &mcpsdk.ReadResourceParams{URI: "workbench:///invalid-resource"})
	assertJSONRPCError(t, err, jsonrpc.CodeInvalidParams, "Resource URI is invalid")

	server.SDKServer().AddResource(&mcpsdk.Resource{
		URI:      "workbench:///dependency-failed",
		Name:     "dependency-failed",
		MIMEType: "text/plain",
	}, func(context.Context, *mcpsdk.ReadResourceRequest) (*mcpsdk.ReadResourceResult, error) {
		return nil, errs.New(
			"Dependency failed",
			errs.WithSentinel(errs.ErrDependencyFailed),
			errs.WithCode("workbench.test.dependency_failed"),
			errs.WithSeverity(errs.SeverityError),
		)
	})
	_, err = session.ReadResource(ctx, &mcpsdk.ReadResourceParams{URI: "workbench:///dependency-failed"})
	assertJSONRPCError(t, err, jsonrpc.CodeInternalError, "Dependency failed")
}

func decodeStructured[T any](t *testing.T, res *mcpsdk.CallToolResult) T {
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

func decodeStructuredInto(t *testing.T, res *mcpsdk.CallToolResult, out any) {
	t.Helper()
	raw, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, out); err != nil {
		t.Fatalf("decode structured content: %v\n%s", err, raw)
	}
}

func mcpTestSession(t *testing.T, ctx context.Context) (*Server, *mcpsdk.ClientSession, func()) {
	t.Helper()
	server, err := New(Options{ArtifactDir: t.TempDir(), SyncTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test-client"}, nil)
	ct, st := mcpsdk.NewInMemoryTransports()
	ss, err := server.SDKServer().Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		ss.Close()
		t.Fatal(err)
	}
	cleanup := func() {
		cs.Close()
		ss.Close()
	}
	return server, cs, cleanup
}

func assertJSONRPCError(t *testing.T, err error, code int64, message string) {
	t.Helper()
	if err == nil {
		t.Fatalf("got nil error, want JSON-RPC code %d", code)
	}
	var rpcErr *jsonrpc.Error
	if !errors.As(err, &rpcErr) {
		t.Fatalf("got error type %T, want jsonrpc.Error: %v", err, err)
	}
	if rpcErr.Code != code {
		t.Fatalf("JSON-RPC code = %d, want %d", rpcErr.Code, code)
	}
	if rpcErr.Message != message {
		t.Fatalf("JSON-RPC message = %q, want %q", rpcErr.Message, message)
	}
	if len(rpcErr.Data) > 0 {
		t.Fatalf("JSON-RPC data leaked details: %s", rpcErr.Data)
	}
}

func toolListed(result *mcpsdk.ListToolsResult, name string) bool {
	for _, tool := range result.Tools {
		if tool.Name == name {
			return true
		}
	}
	return false
}

func toolNames(result *mcpsdk.ListToolsResult) []string {
	names := make([]string, 0, len(result.Tools))
	for _, tool := range result.Tools {
		names = append(names, tool.Name)
	}
	return names
}

func resourceListed(result *mcpsdk.ListResourcesResult, uri string) bool {
	for _, resource := range result.Resources {
		if resource.URI == uri {
			return true
		}
	}
	return false
}
