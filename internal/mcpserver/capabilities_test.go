package mcpserver

import (
	"context"
	"slices"
	"testing"
	"time"
)

func TestCapabilityPlanningAndDiffing(t *testing.T) {
	server, err := New(Options{ArtifactDir: t.TempDir(), SyncTimeout: time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	emptyPlan, err := server.plan(context.Background(), ContextState{})
	if err != nil {
		t.Fatal(err)
	}
	if hasTool(emptyPlan.Index, "artifact.update") {
		t.Fatal("artifact.update active without selected artifact")
	}

	state := ContextState{ArtifactID: ptr("artifact-1")}
	selectedPlan, err := server.plan(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	if !hasTool(selectedPlan.Index, "artifact.update") {
		t.Fatal("artifact.update inactive with selected artifact")
	}
	if !hasResource(selectedPlan.Index, "workbench:///artifacts/artifact-1") {
		t.Fatal("selected artifact resource missing")
	}
	diff := server.diffPlan(selectedPlan)
	if !slices.Contains(diff, "tools") || !slices.Contains(diff, "resources") {
		t.Fatalf("diff = %v, want tools and resources", diff)
	}
}

func TestCapabilitySyncWaitAndTimeout(t *testing.T) {
	syncer := NewCapabilitySync(200 * time.Millisecond)
	tracker := syncer.Begin([]string{"tools", "resources"})
	go func() {
		time.Sleep(10 * time.Millisecond)
		syncer.MarkObserved(methodListTools)
		syncer.MarkObserved(methodListResources)
	}()
	status := syncer.Wait(context.Background(), tracker)
	if status.Status != "synced" || status.TimedOut {
		t.Fatalf("status = %+v, want synced", status)
	}

	syncer.SetTimeout(time.Millisecond)
	tracker = syncer.Begin([]string{"tools"})
	status = syncer.Wait(context.Background(), tracker)
	if status.Status != "timeout_fallback" || !status.TimedOut {
		t.Fatalf("status = %+v, want timeout fallback", status)
	}
}

func TestContextTimeoutReturnsFallbackCapabilityIndex(t *testing.T) {
	server, err := New(Options{ArtifactDir: t.TempDir(), SyncTimeout: time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	artifact, err := server.artifacts.Begin(BeginArtifactRequest{Type: "rfc", Title: "Timeout RFC"})
	if err != nil {
		t.Fatal(err)
	}
	state := server.context.Apply(ContextPatch{ArtifactID: PatchString{Present: true, Value: artifact.ID}})
	result, err := server.contextResult(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Sync.TimedOut {
		t.Fatalf("sync = %+v, want timeout", result.Sync)
	}
	if result.FallbackCapabilityIndex == nil {
		t.Fatal("missing fallback capability index")
	}
	if !hasTool(*result.FallbackCapabilityIndex, "artifact.update") {
		t.Fatal("fallback index missing artifact.update")
	}
}

func hasTool(index CapabilityIndex, name string) bool {
	for _, tool := range index.Tools {
		if tool.Name == name {
			return true
		}
	}
	return false
}

func hasResource(index CapabilityIndex, uri string) bool {
	for _, resource := range index.Resources {
		if resource.URI == uri {
			return true
		}
	}
	return false
}
