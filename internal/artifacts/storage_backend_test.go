package artifacts

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/manuelibar/workbench/internal/storageclient"
)

func TestStorageBackedStoreUsesStorageServiceForArtifacts(t *testing.T) {
	ctx := context.Background()
	fixture := newArtifactStorageFixture(t)
	defer fixture.server.Close()

	client, err := storageclient.NewClient(storageclient.ClientOptions{BaseURL: fixture.server.URL})
	if err != nil {
		t.Fatal(err)
	}
	store, err := NewStorageStore(StorageBackendOptions{
		Client:       client,
		OrgID:        "acme",
		ProjectID:    "workbench",
		ResourceType: "artifacts",
	}, NewRegistry())
	if err != nil {
		t.Fatal(err)
	}
	store.now = fixture.now

	artifact, err := store.BeginContext(ctx, BeginRequest{Type: "rfc", Title: "Storage RFC"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(artifact.Path, "storage:///acme/workbench/artifacts/") {
		t.Fatalf("artifact path = %q", artifact.Path)
	}
	read, err := store.GetContext(ctx, artifact.ID)
	if err != nil {
		t.Fatal(err)
	}
	if read.ID != artifact.ID || !strings.Contains(read.Markdown, "# Storage RFC") {
		t.Fatalf("read artifact mismatch: %+v", read.Summary)
	}
	updated, err := store.UpdateContext(ctx, artifact.ID, UpdateRequest{
		SetSections: map[string]string{"summary": "Stored through the storage service."},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(updated.Markdown, "Stored through the storage service.") {
		t.Fatalf("updated markdown missing body:\n%s", updated.Markdown)
	}
	list, err := store.ListContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].ID != artifact.ID {
		t.Fatalf("list = %+v", list)
	}
}

type artifactStorageFixture struct {
	server *httptest.Server
	mu     sync.Mutex
	body   map[string]string
	now    func() time.Time
}

func newArtifactStorageFixture(t *testing.T) *artifactStorageFixture {
	t.Helper()
	f := &artifactStorageFixture{
		body: map[string]string{},
		now:  func() time.Time { return time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC) },
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/storage/resources", f.list)
	mux.HandleFunc("POST /api/storage/{id}/update-url", f.updateURL)
	mux.HandleFunc("GET /api/storage/{id}/download-url", f.downloadURL)
	mux.HandleFunc("PUT /objects/{id}", f.putObject)
	mux.HandleFunc("GET /objects/{id}", f.getObject)
	f.server = httptest.NewServer(mux)
	return f
}

func (f *artifactStorageFixture) list(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var resources []storageclient.ResourceSummary
	for id := range f.body {
		ref := storageclient.ResourceRef{
			OrgID:        r.URL.Query().Get("org_id"),
			ProjectID:    r.URL.Query().Get("project_id"),
			ResourceType: r.URL.Query().Get("resource_type"),
			ResourceID:   id,
		}
		resources = append(resources, storageclient.ResourceSummary{Ref: ref, Key: ref.Key(), URI: ref.URI()})
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"resources": resources})
}

func (f *artifactStorageFixture) updateURL(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_ = json.NewEncoder(w).Encode(storageclient.UpdateURLResult{
		Ref: storageclient.ResourceRef{OrgID: "acme", ProjectID: "workbench", ResourceType: "artifacts", ResourceID: id},
		Key: "acme/workbench/artifacts/" + id + ".md",
		UploadURL: storageclient.SignedURL{
			URL:    f.server.URL + "/objects/" + id,
			Method: http.MethodPut,
			Headers: map[string]string{
				"Content-Type": "text/markdown",
			},
		},
	})
}

func (f *artifactStorageFixture) downloadURL(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_ = json.NewEncoder(w).Encode(storageclient.DownloadURLResult{
		Ref: storageclient.ResourceRef{OrgID: "acme", ProjectID: "workbench", ResourceType: "artifacts", ResourceID: id},
		Key: "acme/workbench/artifacts/" + id + ".md",
		DownloadURL: storageclient.SignedURL{
			URL:    f.server.URL + "/objects/" + id,
			Method: http.MethodGet,
		},
		SupportsRange: true,
	})
}

func (f *artifactStorageFixture) putObject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	f.mu.Lock()
	f.body[id] = string(raw)
	f.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

func (f *artifactStorageFixture) getObject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	f.mu.Lock()
	body, ok := f.body[id]
	f.mu.Unlock()
	if !ok {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/markdown")
	_, _ = w.Write([]byte(body))
}
