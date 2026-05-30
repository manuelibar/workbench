package mcpserver_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/manuelibar/workbench/internal/mcpserver"
	mcpskills "github.com/manuelibar/workbench/internal/mcpserver/skills"
)

func newTestServer(t *testing.T) *mcpserver.Server {
	t.Helper()
	log := slog.New(slog.NewTextHandler(testWriter{t}, &slog.HandlerOptions{Level: slog.LevelDebug}))
	return mcpserver.New(log, mcpserver.NewMemProjectStore(), mcpskills.NewEmbeddedRegistry())
}

type testWriter struct{ t *testing.T }

func (w testWriter) Write(p []byte) (int, error) { w.t.Log(string(p)); return len(p), nil }

func connect(ctx context.Context, t *testing.T, srv *mcpserver.Server) *mcp.ClientSession {
	t.Helper()
	t1, t2 := mcp.NewInMemoryTransports()
	if _, err := srv.SDKServer().Connect(ctx, t1, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0.1"}, nil).Connect(ctx, t2, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { cs.Close() })
	return cs
}

func callTool(ctx context.Context, t *testing.T, cs *mcp.ClientSession, name string, args map[string]any) map[string]any {
	t.Helper()
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("CallTool %q: %v", name, err)
	}
	if res.IsError {
		t.Fatalf("CallTool %q error: %v", name, res.Content)
	}
	if len(res.Content) == 0 {
		t.Fatalf("CallTool %q: empty content", name)
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(res.Content[0].(*mcp.TextContent).Text), &out); err != nil {
		t.Fatalf("CallTool %q unmarshal: %v", name, err)
	}
	return out
}

func readResource(ctx context.Context, t *testing.T, cs *mcp.ClientSession, uri string) string {
	t.Helper()
	res, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: uri})
	if err != nil {
		t.Fatalf("ReadResource %q: %v", uri, err)
	}
	if len(res.Contents) == 0 {
		t.Fatalf("ReadResource %q: empty", uri)
	}
	return res.Contents[0].Text
}

func asMap(t *testing.T, s string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, s)
	}
	return m
}

func skillsFrom(t *testing.T, result map[string]any) []map[string]any {
	t.Helper()
	arr, ok := result["skills"].([]any)
	if !ok {
		t.Fatalf("no skills[] in result: %v", result)
	}
	out := make([]map[string]any, len(arr))
	for i, v := range arr {
		out[i] = v.(map[string]any)
	}
	return out
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return string(b)
}

func assertNoLegacySkillURI(t *testing.T, label string, v any) {
	t.Helper()
	if strings.Contains(mustJSON(t, v), "workbench:///skills/") {
		t.Fatalf("%s emitted legacy workbench skill URI: %s", label, mustJSON(t, v))
	}
}

func listResources(ctx context.Context, t *testing.T, cs *mcp.ClientSession) (concrete []string, templates []string) {
	t.Helper()
	for r, err := range cs.Resources(ctx, nil) {
		if err != nil {
			t.Fatalf("resources iter: %v", err)
		}
		concrete = append(concrete, r.URI)
	}
	for r, err := range cs.ResourceTemplates(ctx, nil) {
		if err != nil {
			t.Fatalf("templates iter: %v", err)
		}
		templates = append(templates, r.URITemplate)
	}
	return
}

func TestRefreshOverviewMatchesScopeOverviewResource(t *testing.T) {
	ctx := context.Background()
	srv := newTestServer(t)
	cs := connect(ctx, t, srv)

	callTool(ctx, t, cs, "refresh", nil)
	created := callTool(ctx, t, cs, "project.create", map[string]any{
		"name":          "alpha",
		"description":   "Test project for shared overview.",
		"system_prompt": "Be terse.",
	})
	projectID := created["project"].(map[string]any)["id"].(string)

	result := callTool(ctx, t, cs, "refresh", map[string]any{"project_id": projectID})
	if result["overview_uri"] != "workbench:///scope/overview" {
		t.Fatalf("expected overview_uri to point at scope overview, got %v", result["overview_uri"])
	}
	refreshOverview, ok := result["overview"].(map[string]any)
	if !ok {
		t.Fatalf("refresh result missing overview object: %v", result)
	}
	resourceOverview := asMap(t, readResource(ctx, t, cs, "workbench:///scope/overview"))

	for _, field := range []string{"selection", "summary", "navigation", "recommended_resources", "capability_state_note", "self_resource_uri"} {
		if mustJSON(t, refreshOverview[field]) != mustJSON(t, resourceOverview[field]) {
			t.Fatalf("overview field %q mismatch\nrefresh:  %s\nresource: %s", field, mustJSON(t, refreshOverview[field]), mustJSON(t, resourceOverview[field]))
		}
	}
}

func TestScopeCapabilitiesResourceIsAbsent(t *testing.T) {
	ctx := context.Background()
	srv := newTestServer(t)
	cs := connect(ctx, t, srv)

	conc, _ := listResources(ctx, t, cs)
	for _, uri := range conc {
		if uri == "workbench:///scope/capabilities" {
			t.Fatalf("scope/capabilities should not be registered; got resources %v", conc)
		}
	}

	callTool(ctx, t, cs, "refresh", nil)
	conc, _ = listResources(ctx, t, cs)
	for _, uri := range conc {
		if uri == "workbench:///scope/capabilities" {
			t.Fatalf("scope/capabilities should not be registered after refresh; got resources %v", conc)
		}
	}

	if _, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: "workbench:///scope/capabilities"}); err == nil {
		t.Fatalf("expected scope/capabilities read to fail because MCP list methods are the manifest")
	}
}

func TestNoNewPayloadEmitsLegacyWorkbenchSkillURIs(t *testing.T) {
	ctx := context.Background()
	srv := newTestServer(t)
	cs := connect(ctx, t, srv)

	result := callTool(ctx, t, cs, "refresh", nil)
	assertNoLegacySkillURI(t, "refresh result", result)
	assertNoLegacySkillURI(t, "scope overview", asMap(t, readResource(ctx, t, cs, "workbench:///scope/overview")))
	assertNoLegacySkillURI(t, "skill manifest", asMap(t, readResource(ctx, t, cs, "skill://workbench-orient/manifest")))
}

// TestAcceptance runs the ten-step acceptance sequence.
//
// Key behaviour being tested:
//   - refresh() is the synchronization point: it registers/deregisters skill
//     resources, waits for observed capability list calls or returns a fallback
//     index, then returns.
//   - After refresh(), resources/list reflects the new capability set.
//   - Two independent server instances are fully isolated.
func TestAcceptance(t *testing.T) {
	ctx := context.Background()

	// Step 1: two independent server instances simulate two parallel agent sessions.
	srvA := newTestServer(t)
	srvB := newTestServer(t)
	csA := connect(ctx, t, srvA)
	csB := connect(ctx, t, srvB)

	// Step 2: both see only permanent core tools before refresh.
	wantTools := map[string]bool{
		"refresh": true, "feedback": true, "query": true,
	}
	for tool, err := range csA.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("tools iter: %v", err)
		}
		if !wantTools[tool.Name] {
			t.Fatalf("unexpected tool %q", tool.Name)
		}
	}

	// Step 3: before any refresh(), resource surface is static-only.
	// 5 concrete (scope/overview, github/config, project snapshot, tasks, knowledge) + 3 templates.
	// Skill resources are not yet registered.
	conc, tmpl := listResources(ctx, t, csA)
	if len(conc) != 5 {
		t.Fatalf("expected 5 static concrete resources before refresh, got %v", conc)
	}
	if len(tmpl) != 3 {
		t.Fatalf("expected 3 static templates before refresh, got %v", tmpl)
	}

	// Step 4: refresh() — selection empty, workbench-orient skills registered.
	// After this call: 7 concrete (5 static + orient/manifest + orient/SKILL.md).
	result := callTool(ctx, t, csA, "refresh", nil)
	sel := result["selection"].(map[string]any)
	if v := sel["project_id"]; v != nil && v != "" {
		t.Fatalf("expected empty selection, got %v", sel)
	}
	sk := skillsFrom(t, result)
	if len(sk) != 1 || sk[0]["name"] != "workbench-orient" {
		t.Fatalf("expected [workbench-orient], got %v", sk)
	}
	if !strings.Contains(sk[0]["instructions"].(string), "refresh") {
		t.Fatalf("workbench-orient instructions look wrong")
	}
	// Verify dynamic resources are now registered.
	conc, tmpl = listResources(ctx, t, csA)
	if len(conc) != 7 { // 5 static + 2 orient
		t.Fatalf("expected 7 concrete after refresh(), got %v", conc)
	}
	if len(tmpl) != 3 {
		t.Fatalf("expected 3 templates after refresh(), got %v", tmpl)
	}

	// Step 5: scope/overview resource mirrors the inline refresh overview.
	overview := asMap(t, readResource(ctx, t, csA, "workbench:///scope/overview"))
	if overview["self_resource_uri"] != "workbench:///scope/overview" {
		t.Fatalf("scope overview should identify its own URI, got %v", overview)
	}
	if mustJSON(t, result["overview"].(map[string]any)["selection"]) != mustJSON(t, overview["selection"]) {
		t.Fatalf("scope overview selection drifted from refresh result: %v", overview)
	}

	// Step 6: create a project.
	created := callTool(ctx, t, csA, "project.create", map[string]any{
		"name":          "alpha",
		"description":   "Test project for the acceptance suite.",
		"system_prompt": "Be terse.",
	})
	projectID := created["project"].(map[string]any)["id"].(string)
	if _, err := uuid.Parse(projectID); err != nil {
		t.Fatalf("bad project id: %v", err)
	}

	// Step 7: refresh(project_id=alpha) — one call, full context returned inline.
	// Resources expand: 5 static + 6 skill (orient×2 + system-prompt×2 + go-guidelines×2) = 11 concrete.
	result = callTool(ctx, t, csA, "refresh", map[string]any{"project_id": projectID})
	if result["selection"].(map[string]any)["project_id"] != projectID {
		t.Fatalf("selection not updated: %v", result["selection"])
	}
	sk = skillsFrom(t, result)
	if len(sk) != 3 {
		t.Fatalf("expected 3 skills after project select, got %d: %v", len(sk), sk)
	}
	var sysPrompt map[string]any
	var goGuidelines map[string]any
	for _, s := range sk {
		if s["name"] == "workbench-system-prompt" {
			sysPrompt = s
		}
		if s["name"] == "go-coding-guidelines" {
			goGuidelines = s
		}
	}
	if sysPrompt == nil {
		t.Fatalf("workbench-system-prompt not in skills: %v", sk)
	}
	instr := sysPrompt["instructions"].(string)
	if !strings.Contains(instr, "Be terse.") {
		t.Fatalf("system prompt not rendered:\n%s", instr)
	}
	if !strings.Contains(instr, "alpha") {
		t.Fatalf("project name not rendered:\n%s", instr)
	}
	if goGuidelines == nil {
		t.Fatalf("go-coding-guidelines not in skills: %v", sk)
	}
	if !strings.Contains(goGuidelines["instructions"].(string), "gofmt") {
		t.Fatalf("go guidelines not rendered: %v", goGuidelines)
	}
	// Dynamic resource count after project select.
	conc, _ = listResources(ctx, t, csA)
	if len(conc) != 11 { // 5 static + 6 skill resources
		t.Fatalf("expected 11 concrete after project select, got %v", conc)
	}

	// Step 8 (secondary path): skill SKILL.md resource still renders correctly.
	skillMD := readResource(ctx, t, csA, "skill://workbench-system-prompt/SKILL.md")
	if !strings.Contains(skillMD, "Be terse.") {
		t.Fatalf("SKILL.md resource not rendered: %s", skillMD)
	}

	// Step 9: server B is fully isolated — completely unaware of A's selection.
	resultB := callTool(ctx, t, csB, "refresh", nil)
	skB := skillsFrom(t, resultB)
	if len(skB) != 1 || skB[0]["name"] != "workbench-orient" {
		t.Fatalf("isolation broken: B sees %v", skB)
	}
	concB, _ := listResources(ctx, t, csB)
	if len(concB) != 7 { // B only has static + orient resources
		t.Fatalf("B resource isolation broken: got %v", concB)
	}

	// Step 10: A clears selection → back to orient only.
	result = callTool(ctx, t, csA, "refresh", nil)
	sk = skillsFrom(t, result)
	if len(sk) != 1 || sk[0]["name"] != "workbench-orient" {
		t.Fatalf("expected [workbench-orient] after deselect, got %v", sk)
	}
	conc, _ = listResources(ctx, t, csA)
	if len(conc) != 7 { // back to 5 static + 2 orient
		t.Fatalf("expected 7 concrete after deselect, got %v", conc)
	}

	t.Log("all 10 acceptance steps passed")
}
