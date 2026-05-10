package mcpserver_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	_ "github.com/manuelibar/workbench/internal/domain"
	"github.com/manuelibar/workbench/internal/mcpserver"
	"github.com/manuelibar/workbench/internal/pgstore"
)

const skillURI = "workbench://skill"

func integrationDSN() string {
	if dsn := os.Getenv("WORKBENCH_DB_URL"); dsn != "" {
		return dsn
	}
	return "postgres://workbench:workbench@127.0.0.1:5432/workbench?sslmode=disable"
}

// setUpIntegration boots a real Store + bootstraps user/ws + returns a
// constructed *mcpserver.Server. Skips when -short is passed. Each call
// resets the open WorkSession's selection to empty so tests start clean.
func setUpIntegration(t *testing.T) (*mcpserver.Server, context.Context) {
	t.Helper()
	if testing.Short() {
		t.Skip("integration test (requires Postgres on " + integrationDSN() + ")")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	store, err := pgstore.Open(ctx, integrationDSN())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(store.Close)

	if err := store.Migrate(ctx); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	user, err := store.EnsureSingletonUser(ctx, "")
	if err != nil {
		t.Fatalf("EnsureSingletonUser: %v", err)
	}
	ws, err := store.EnsureOpenWorkSession(ctx, user.ID, "")
	if err != nil {
		t.Fatalf("EnsureOpenWorkSession: %v", err)
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))
	return mcpserver.New(store, user, ws, logger), ctx
}

// connectClient pairs a fresh in-memory transport with the singleton
// SDK server and returns the connected *mcp.ClientSession.
func connectClient(t *testing.T, ctx context.Context, s *mcpserver.Server, name string) *mcp.ClientSession {
	t.Helper()
	tServer, tClient := mcp.NewInMemoryTransports()
	if _, err := s.SDKServer().Connect(ctx, tServer, nil); err != nil {
		t.Fatalf("server.Connect: %v", err)
	}
	c := mcp.NewClient(&mcp.Implementation{Name: name, Version: "v0.0.1"}, nil)
	cs, err := c.Connect(ctx, tClient, nil)
	if err != nil {
		t.Fatalf("client.Connect: %v", err)
	}
	t.Cleanup(func() { _ = cs.Close() })
	return cs
}

func toolNames(t *testing.T, ctx context.Context, cs *mcp.ClientSession) []string {
	t.Helper()
	var names []string
	for tt, err := range cs.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("Tools iter: %v", err)
		}
		names = append(names, tt.Name)
	}
	slices.Sort(names)
	return names
}

func resetSelection(t *testing.T, ctx context.Context, cs *mcp.ClientSession) {
	t.Helper()
	if _, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "refresh",
		Arguments: map[string]any{"clear": true},
	}); err != nil {
		t.Fatalf("refresh clear: %v", err)
	}
}

func TestServer_AlwaysOnTools(t *testing.T) {
	s, ctx := setUpIntegration(t)
	cs := connectClient(t, ctx, s, "test-client")
	resetSelection(t, ctx, cs)
	got := toolNames(t, ctx, cs)
	want := []string{
		"ask",
		"namespace.create", "namespace.list",
		"note.add", "note.delete", "note.get", "note.list", "note.search", "note.update",
		"refresh",
	}
	if !slices.Equal(got, want) {
		t.Errorf("tools = %v\nwant %v", got, want)
	}
}

func TestServer_SkillResource(t *testing.T) {
	s, ctx := setUpIntegration(t)
	cs := connectClient(t, ctx, s, "test-client")
	res, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: skillURI})
	if err != nil {
		t.Fatalf("ReadResource: %v", err)
	}
	if len(res.Contents) == 0 {
		t.Fatal("no contents")
	}
	if !strings.Contains(res.Contents[0].Text, "Workbench MCP") {
		t.Errorf("skill resource missing onboarding header; got: %.80q", res.Contents[0].Text)
	}
}

func TestServer_RefreshNoSelection(t *testing.T) {
	s, ctx := setUpIntegration(t)
	cs := connectClient(t, ctx, s, "test-client")
	resetSelection(t, ctx, cs)

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "refresh",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool refresh: %v", err)
	}
	if res.IsError {
		t.Fatalf("refresh returned IsError; content=%v", res.Content)
	}
	var got mcpserver.RefreshResult
	raw, _ := json.Marshal(res.StructuredContent)
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal RefreshResult: %v", err)
	}
	if !got.Selection.IsEmpty() {
		t.Errorf("expected empty selection on a fresh WorkSession, got %+v", got.Selection)
	}
	if !got.Synced {
		t.Errorf("expected synced=true")
	}
	if len(got.Tools) == 0 {
		t.Errorf("expected at least one tool in refresh result")
	}
	hasSkill := false
	for _, r := range got.Resources {
		if r.URI == skillURI {
			hasSkill = true
			break
		}
	}
	if !hasSkill {
		t.Errorf("refresh result resources missing %q; got %+v", skillURI, got.Resources)
	}
}

func TestServer_NoteAddListThenRefreshShowsEvent(t *testing.T) {
	s, ctx := setUpIntegration(t)
	cs := connectClient(t, ctx, s, "note-rt")

	body := "phase4-rt-" + time.Now().UTC().Format("150405.000000")
	addRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "note.add",
		Arguments: map[string]any{"text": body, "tags": []string{"rt"}},
	})
	if err != nil {
		t.Fatalf("note.add: %v", err)
	}
	if addRes.IsError {
		t.Fatalf("note.add returned IsError; content=%v", addRes.Content)
	}

	var added struct {
		Note mcpserver.NoteWire `json:"note"`
	}
	raw, _ := json.Marshal(addRes.StructuredContent)
	if err := json.Unmarshal(raw, &added); err != nil {
		t.Fatalf("decode note.add result: %v", err)
	}
	if added.Note.BodyMD != body {
		t.Errorf("note.add returned body %q, want %q", added.Note.BodyMD, body)
	}

	searchRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "note.search",
		Arguments: map[string]any{"query": body},
	})
	if err != nil {
		t.Fatalf("note.search: %v", err)
	}
	var searched struct {
		Notes []mcpserver.NoteWire `json:"notes"`
		Count int                  `json:"count"`
	}
	raw, _ = json.Marshal(searchRes.StructuredContent)
	if err := json.Unmarshal(raw, &searched); err != nil {
		t.Fatalf("decode note.search result: %v", err)
	}
	if searched.Count == 0 {
		t.Errorf("note.search returned no results for %q", body)
	}

	refreshRes, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "refresh", Arguments: map[string]any{}})
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	var refreshed mcpserver.RefreshResult
	raw, _ = json.Marshal(refreshRes.StructuredContent)
	if err := json.Unmarshal(raw, &refreshed); err != nil {
		t.Fatalf("decode refresh result: %v", err)
	}
	if !refreshed.Synced {
		t.Error("refresh: expected synced=true")
	}
	seenAdd, seenSearch := false, false
	for _, ev := range refreshed.RecentEvents {
		if ev.SubjectKind == "tool" {
			switch ev.SubjectID {
			case "note.add":
				seenAdd = true
			case "note.search":
				seenSearch = true
			}
		}
	}
	if !seenAdd || !seenSearch {
		t.Errorf("recent_events missing tool calls: seenAdd=%v seenSearch=%v events=%+v",
			seenAdd, seenSearch, refreshed.RecentEvents)
	}

	if _, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "note.delete",
		Arguments: map[string]any{"id": added.Note.ID},
	}); err != nil {
		t.Fatalf("note.delete: %v", err)
	}
}

func TestServer_NamespaceCreateAndSelect(t *testing.T) {
	s, ctx := setUpIntegration(t)
	cs := connectClient(t, ctx, s, "ns-rt")
	resetSelection(t, ctx, cs)

	nsName := "phase5-" + uuid.NewString()[:8]
	createRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "namespace.create",
		Arguments: map[string]any{"name": nsName, "description": "phase5 round-trip"},
	})
	if err != nil {
		t.Fatalf("namespace.create: %v", err)
	}
	if createRes.IsError {
		t.Fatalf("namespace.create IsError; content=%v", createRes.Content)
	}
	var created struct {
		Namespace mcpserver.NamespaceWire `json:"namespace"`
	}
	raw, _ := json.Marshal(createRes.StructuredContent)
	if err := json.Unmarshal(raw, &created); err != nil {
		t.Fatalf("decode namespace.create result: %v", err)
	}
	if created.Namespace.Name != nsName {
		t.Fatalf("name roundtrip differs: %q vs %q", created.Namespace.Name, nsName)
	}
	t.Cleanup(func() {
		_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      "namespace.delete",
			Arguments: map[string]any{"id": created.Namespace.ID},
		})
	})

	// Pre-select: surface should NOT include namespace.delete yet.
	pre := toolNames(t, ctx, cs)
	if slices.Contains(pre, "namespace.delete") {
		t.Errorf("expected namespace.delete to be hidden before selection; got %v", pre)
	}

	// Select via refresh.
	refRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "refresh",
		Arguments: map[string]any{"namespace_id": created.Namespace.ID},
	})
	if err != nil {
		t.Fatalf("refresh select namespace: %v", err)
	}
	if refRes.IsError {
		t.Fatalf("refresh select IsError; content=%v", refRes.Content)
	}

	post := toolNames(t, ctx, cs)
	for _, expected := range []string{"namespace.get", "namespace.update", "namespace.delete", "project.create", "project.list"} {
		if !slices.Contains(post, expected) {
			t.Errorf("expected %q in surface after selecting namespace; got %v", expected, post)
		}
	}

	// namespace.get with no id should default to selected.
	getRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "namespace.get",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("namespace.get: %v", err)
	}
	if getRes.IsError {
		t.Fatalf("namespace.get IsError; content=%v", getRes.Content)
	}
}

func TestServer_ProjectScopedSurface(t *testing.T) {
	s, ctx := setUpIntegration(t)
	cs := connectClient(t, ctx, s, "p6-rt")
	resetSelection(t, ctx, cs)

	// 1. Create namespace + project + select.
	nsName := "p6-ns-" + uuid.NewString()[:8]
	nsCreate, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "namespace.create",
		Arguments: map[string]any{"name": nsName},
	})
	if err != nil || nsCreate.IsError {
		t.Fatalf("namespace.create: err=%v isErr=%v content=%v", err, nsCreate.IsError, nsCreate.Content)
	}
	var nsOut struct {
		Namespace mcpserver.NamespaceWire `json:"namespace"`
	}
	json.Unmarshal(mustJSON(t, nsCreate.StructuredContent), &nsOut)
	t.Cleanup(func() {
		_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      "namespace.delete",
			Arguments: map[string]any{"id": nsOut.Namespace.ID},
		})
	})

	if _, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "refresh",
		Arguments: map[string]any{"namespace_id": nsOut.Namespace.ID},
	}); err != nil {
		t.Fatalf("select namespace: %v", err)
	}

	pName := "p6-proj-" + uuid.NewString()[:8]
	projCreate, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "project.create",
		Arguments: map[string]any{"name": pName},
	})
	if err != nil || projCreate.IsError {
		t.Fatalf("project.create: err=%v isErr=%v content=%v", err, projCreate.IsError, projCreate.Content)
	}
	var projOut struct {
		Project mcpserver.ProjectWire `json:"project"`
	}
	json.Unmarshal(mustJSON(t, projCreate.StructuredContent), &projOut)

	if _, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "refresh",
		Arguments: map[string]any{"project_id": projOut.Project.ID},
	}); err != nil {
		t.Fatalf("select project: %v", err)
	}

	post := toolNames(t, ctx, cs)
	for _, expected := range []string{"artifact.create", "artifact.list", "skill.create", "prompt.create", "project.update"} {
		if !slices.Contains(post, expected) {
			t.Errorf("expected %q in surface after selecting project; got %v", expected, post)
		}
	}

	// 2. Round-trip an artifact through create / update / get.
	artCreate, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "artifact.create",
		Arguments: map[string]any{
			"type":         "note",
			"content":      map[string]any{"hello": "world"},
			"content_text": "v1 body",
		},
	})
	if err != nil || artCreate.IsError {
		t.Fatalf("artifact.create: err=%v isErr=%v content=%v", err, artCreate.IsError, artCreate.Content)
	}
	var artOut struct {
		Artifact mcpserver.ArtifactWire `json:"artifact"`
	}
	json.Unmarshal(mustJSON(t, artCreate.StructuredContent), &artOut)
	if artOut.Artifact.LatestVersion != 1 {
		t.Errorf("expected latest_version=1, got %d", artOut.Artifact.LatestVersion)
	}

	updateRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "artifact.update",
		Arguments: map[string]any{
			"id":           artOut.Artifact.ID,
			"content_text": "v2 body",
		},
	})
	if err != nil || updateRes.IsError {
		t.Fatalf("artifact.update: err=%v isErr=%v content=%v", err, updateRes.IsError, updateRes.Content)
	}
	var updateOut struct {
		Artifact   mcpserver.ArtifactWire `json:"artifact"`
		NewVersion int                    `json:"new_version,omitempty"`
	}
	json.Unmarshal(mustJSON(t, updateRes.StructuredContent), &updateOut)
	if updateOut.Artifact.LatestVersion != 2 {
		t.Errorf("expected latest_version=2 after update, got %d", updateOut.Artifact.LatestVersion)
	}

	// 3. Skill round-trip — including the workbench:///skills/{name} resource.
	skName := "intro"
	skCreate, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "skill.create",
		Arguments: map[string]any{"name": skName, "body_md": "# Hello\nWorld"},
	})
	if err != nil || skCreate.IsError {
		t.Fatalf("skill.create: err=%v isErr=%v content=%v", err, skCreate.IsError, skCreate.Content)
	}

	skURI := "workbench:///skills/" + skName
	rr, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: skURI})
	if err != nil {
		t.Fatalf("ReadResource %s: %v", skURI, err)
	}
	if len(rr.Contents) == 0 || !strings.Contains(rr.Contents[0].Text, "Hello") {
		t.Errorf("skill resource content unexpected: %+v", rr.Contents)
	}

	// 4. Prompt round-trip.
	pCreate, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "prompt.create",
		Arguments: map[string]any{
			"name": "review",
			"body": "Review this change: {{change}}",
			"args": []map[string]any{{"name": "change", "required": true}},
		},
	})
	if err != nil || pCreate.IsError {
		t.Fatalf("prompt.create: err=%v isErr=%v content=%v", err, pCreate.IsError, pCreate.Content)
	}

	pURI := "workbench:///prompts/review"
	rr, err = cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: pURI})
	if err != nil {
		t.Fatalf("ReadResource %s: %v", pURI, err)
	}
	if len(rr.Contents) == 0 || !strings.Contains(rr.Contents[0].Text, "{{change}}") {
		t.Errorf("prompt resource content unexpected: %+v", rr.Contents)
	}

	// 5. project.delete also cleans up the descendants and clears project parts of selection.
	if _, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "project.delete",
	}); err != nil {
		t.Fatalf("project.delete: %v", err)
	}
	postDelete := toolNames(t, ctx, cs)
	if slices.Contains(postDelete, "artifact.create") {
		t.Errorf("expected project-scoped tools to disappear after project.delete; got %v", postDelete)
	}
}

// mustJSON re-marshals StructuredContent to bytes for unmarshaling into a
// typed struct. Returns nil bytes on error (test will then fail on Unmarshal).
func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal structuredContent: %v", err)
	}
	return raw
}

func TestServer_FullSelectionChainAndArtifactLifecycle(t *testing.T) {
	s, ctx := setUpIntegration(t)
	cs := connectClient(t, ctx, s, "p7-rt")
	resetSelection(t, ctx, cs)

	// Namespace.
	nsCreate, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "namespace.create",
		Arguments: map[string]any{"name": "p7-ns-" + uuid.NewString()[:8]},
	})
	if err != nil {
		t.Fatalf("namespace.create: %v", err)
	}
	if nsCreate.IsError {
		t.Fatalf("namespace.create IsError; content=%v", nsCreate.Content)
	}
	var ns struct {
		Namespace mcpserver.NamespaceWire `json:"namespace"`
	}
	json.Unmarshal(mustJSON(t, nsCreate.StructuredContent), &ns)
	t.Cleanup(func() {
		_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      "namespace.delete",
			Arguments: map[string]any{"id": ns.Namespace.ID},
		})
	})

	// Project.
	if _, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "refresh", Arguments: map[string]any{"namespace_id": ns.Namespace.ID},
	}); err != nil {
		t.Fatalf("select namespace: %v", err)
	}
	pjCreate, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "project.create",
		Arguments: map[string]any{"name": "p7-proj-" + uuid.NewString()[:8]},
	})
	if err != nil {
		t.Fatalf("project.create: %v", err)
	}
	if pjCreate.IsError {
		t.Fatalf("project.create IsError; content=%v", pjCreate.Content)
	}
	var pj struct {
		Project mcpserver.ProjectWire `json:"project"`
	}
	json.Unmarshal(mustJSON(t, pjCreate.StructuredContent), &pj)

	// Select project — selection auto-resolves namespace.
	if _, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "refresh", Arguments: map[string]any{"project_id": pj.Project.ID},
	}); err != nil {
		t.Fatalf("select project: %v", err)
	}

	// Blueprint v1.
	bpCreate, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "blueprint.create",
		Arguments: map[string]any{"name": "review-loop", "definition": map[string]any{"v": 1}},
	})
	if err != nil {
		t.Fatalf("blueprint.create: %v", err)
	}
	if bpCreate.IsError {
		t.Fatalf("blueprint.create IsError; content=%v", bpCreate.Content)
	}
	var bpV1 struct {
		Blueprint mcpserver.BlueprintWire `json:"blueprint"`
	}
	json.Unmarshal(mustJSON(t, bpCreate.StructuredContent), &bpV1)
	if bpV1.Blueprint.Version != 1 {
		t.Errorf("expected blueprint version=1; got %d", bpV1.Blueprint.Version)
	}

	// Select blueprint v1.
	if _, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "refresh", Arguments: map[string]any{"blueprint_id": bpV1.Blueprint.ID},
	}); err != nil {
		t.Fatalf("select blueprint: %v", err)
	}
	post := toolNames(t, ctx, cs)
	for _, expected := range []string{"mode.create", "mode.list", "blueprint.update"} {
		if !slices.Contains(post, expected) {
			t.Errorf("expected %q after selecting blueprint; got %v", expected, post)
		}
	}

	// Mode on v1 (latest) — succeeds.
	mc, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "mode.create",
		Arguments: map[string]any{"name": "review", "system_prompt": "Be terse."},
	})
	if err != nil {
		t.Fatalf("mode.create on v1: %v", err)
	}
	if mc.IsError {
		t.Fatalf("mode.create on v1 IsError; content=%v", mc.Content)
	}

	// Roll a v2 of the blueprint.
	bpUp, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "blueprint.update",
		Arguments: map[string]any{"name": "review-loop", "definition": map[string]any{"v": 2}},
	})
	if err != nil {
		t.Fatalf("blueprint.update: %v", err)
	}
	if bpUp.IsError {
		t.Fatalf("blueprint.update IsError; content=%v", bpUp.Content)
	}
	var bpV2 struct {
		Blueprint mcpserver.BlueprintWire `json:"blueprint"`
	}
	json.Unmarshal(mustJSON(t, bpUp.StructuredContent), &bpV2)
	if bpV2.Blueprint.Version != 2 {
		t.Errorf("expected blueprint version=2; got %d", bpV2.Blueprint.Version)
	}

	// Mode on v1 — must fail because v1 is no longer latest.
	mc2, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "mode.create",
		Arguments: map[string]any{"name": "ship", "system_prompt": "Yes."},
	})
	if err == nil && !mc2.IsError {
		t.Errorf("expected mode.create on stale blueprint to fail; got success")
	}

	// Read blueprint v1 via resource.
	bpURI := "workbench:///blueprints/review-loop/1"
	rr, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: bpURI})
	if err != nil {
		t.Fatalf("ReadResource %s: %v", bpURI, err)
	}
	if !strings.Contains(rr.Contents[0].Text, "\"v\":1") {
		t.Errorf("blueprint v1 resource: unexpected content: %q", rr.Contents[0].Text)
	}

	// Backlog round-trip + lifecycle.
	addRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "backlog.add",
		Arguments: map[string]any{"title": "ship the readme"},
	})
	if err != nil {
		t.Fatalf("backlog.add: %v", err)
	}
	if addRes.IsError {
		t.Fatalf("backlog.add IsError; content=%v", addRes.Content)
	}
	var added struct {
		Artifact mcpserver.ArtifactWire `json:"artifact"`
	}
	json.Unmarshal(mustJSON(t, addRes.StructuredContent), &added)
	if added.Artifact.Type != "task" || added.Artifact.Status != "draft" {
		t.Errorf("backlog.add: type=%q status=%q (expected type=task status=draft)", added.Artifact.Type, added.Artifact.Status)
	}

	tnRes, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "backlog.take_next"})
	if err != nil {
		t.Fatalf("backlog.take_next: %v", err)
	}
	if tnRes.IsError {
		t.Fatalf("backlog.take_next IsError; content=%v", tnRes.Content)
	}
	var tn mcpserver.BacklogTakeNextResult
	json.Unmarshal(mustJSON(t, tnRes.StructuredContent), &tn)
	if !tn.Found || tn.Task.ID != added.Artifact.ID {
		t.Errorf("backlog.take_next: expected the just-added task; got %+v", tn)
	}

	signRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "artifact.sign_off",
		Arguments: map[string]any{"id": added.Artifact.ID},
	})
	if err != nil {
		t.Fatalf("artifact.sign_off: %v", err)
	}
	if signRes.IsError {
		t.Fatalf("artifact.sign_off IsError; content=%v", signRes.Content)
	}
	var signed struct {
		Artifact mcpserver.ArtifactWire `json:"artifact"`
	}
	json.Unmarshal(mustJSON(t, signRes.StructuredContent), &signed)
	if signed.Artifact.Status != "signed_off" {
		t.Errorf("artifact.sign_off did not change status: %+v", signed)
	}

	// Cleanup project (cascades artifacts/skills/prompts/blueprints/modes).
	if _, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "project.delete"}); err != nil {
		t.Fatalf("project.delete: %v", err)
	}
}

func TestServer_TwoClientsFanOut(t *testing.T) {
	s, ctx := setUpIntegration(t)
	cs1 := connectClient(t, ctx, s, "client-1")
	cs2 := connectClient(t, ctx, s, "client-2")
	resetSelection(t, ctx, cs1)

	// Pre-select: both see the bootstrap surface.
	pre1 := toolNames(t, ctx, cs1)
	pre2 := toolNames(t, ctx, cs2)
	if !slices.Equal(pre1, pre2) {
		t.Fatalf("pre-selection mismatch:\n  c1: %v\n  c2: %v", pre1, pre2)
	}
	if slices.Contains(pre1, "namespace.delete") {
		t.Fatalf("expected bootstrap surface; got %v", pre1)
	}

	// Client 1 creates a namespace and selects it.
	nsName := "fanout-" + uuid.NewString()[:8]
	createRes, err := cs1.CallTool(ctx, &mcp.CallToolParams{
		Name:      "namespace.create",
		Arguments: map[string]any{"name": nsName},
	})
	if err != nil {
		t.Fatalf("namespace.create: %v", err)
	}
	var created struct {
		Namespace mcpserver.NamespaceWire `json:"namespace"`
	}
	raw, _ := json.Marshal(createRes.StructuredContent)
	if err := json.Unmarshal(raw, &created); err != nil {
		t.Fatalf("decode namespace.create result: %v", err)
	}
	t.Cleanup(func() {
		_, _ = cs1.CallTool(ctx, &mcp.CallToolParams{
			Name:      "namespace.delete",
			Arguments: map[string]any{"id": created.Namespace.ID},
		})
	})

	if _, err := cs1.CallTool(ctx, &mcp.CallToolParams{
		Name:      "refresh",
		Arguments: map[string]any{"namespace_id": created.Namespace.ID},
	}); err != nil {
		t.Fatalf("refresh: %v", err)
	}

	// Client 2 lists tools again — should see namespace-scoped surface
	// because the singleton *mcp.Server's tool set was mutated.
	post2 := toolNames(t, ctx, cs2)
	for _, expected := range []string{"namespace.delete", "project.create"} {
		if !slices.Contains(post2, expected) {
			t.Errorf("client 2 missing %q after client 1 selection; got %v", expected, post2)
		}
	}
}
