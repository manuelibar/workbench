package storage

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestServiceProcessesInboxObjectIntoIndexedMarkdown(t *testing.T) {
	ctx := context.Background()
	objects := newMemoryObjectStore()
	service, err := NewService(ServiceOptions{
		Objects:    objects,
		Normalizer: PassthroughNormalizer{},
	})
	if err != nil {
		t.Fatal(err)
	}
	service.now = func() time.Time { return time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC) }

	grant, err := service.UploadURL(ctx, UploadURLRequest{
		OrgID:        "acme",
		ProjectID:    "workbench",
		ResourceType: "docs",
		ResourceID:   "resource-1",
		Filename:     "source.txt",
		MIMEType:     "text/plain",
	})
	if err != nil {
		t.Fatal(err)
	}
	objects.put(grant.Key, []byte("# Source\n\nBody.\n"), "text/plain", metadataFor(grant.Ref, "text/plain", false))

	if err := service.ProcessObject(ctx, grant.Key); err != nil {
		t.Fatal(err)
	}
	if objects.exists(grant.Key) {
		t.Fatalf("inbox object %q was not deleted", grant.Key)
	}
	finalObject, err := objects.GetObject(ctx, grant.Ref.Key(), "")
	if err != nil {
		t.Fatal(err)
	}
	if finalObject.ContentType != "text/markdown" {
		t.Fatalf("content type = %q, want text/markdown", finalObject.ContentType)
	}
	if !strings.Contains(string(finalObject.Body), "storage:") ||
		!strings.Contains(string(finalObject.Body), "# Source") {
		t.Fatalf("final object was not indexed markdown:\n%s", finalObject.Body)
	}
	stats, err := service.Stats(ctx, grant.Ref)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Resource.SourceMIMEType != "text/plain" {
		t.Fatalf("source mime = %q", stats.Resource.SourceMIMEType)
	}
}

func TestHandlerExposesUploadStatsAndDownloadURL(t *testing.T) {
	ctx := context.Background()
	objects := newMemoryObjectStore()
	service, err := NewService(ServiceOptions{
		Objects:    objects,
		Normalizer: PassthroughNormalizer{},
	})
	if err != nil {
		t.Fatal(err)
	}
	ref := ResourceRef{OrgID: "acme", ProjectID: "workbench", ResourceType: "docs", ResourceID: "resource-1"}
	indexed, err := IndexMarkdown(ref, "# Source\n\nBody.\n", "text/markdown", time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	objects.put(ref.Key(), []byte(indexed.Markdown), "text/markdown", metadataFor(ref, "text/markdown", true))

	server := httptest.NewServer(NewHandler(service).Routes())
	defer server.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/api/storage/resource-1/stats?org_id=acme&project_id=workbench&resource_type=docs", nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("stats status = %d", res.StatusCode)
	}

	req, err = http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/api/storage/resource-1/download-url?org_id=acme&project_id=workbench&resource_type=docs", nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("download-url status = %d", res.StatusCode)
	}
}

type memoryObjectStore struct {
	mu      sync.Mutex
	objects map[string]Object
}

func newMemoryObjectStore() *memoryObjectStore {
	return &memoryObjectStore{objects: map[string]Object{}}
}

func (s *memoryObjectStore) PresignPut(_ context.Context, key, contentType string, metadata map[string]string, _ time.Duration) (SignedURL, error) {
	headers := map[string]string{"Content-Type": contentType}
	for key, value := range metadata {
		headers["x-amz-meta-"+strings.ReplaceAll(key, "_", "-")] = value
	}
	return SignedURL{URL: "https://example.invalid/" + key, Method: http.MethodPut, Headers: headers}, nil
}

func (s *memoryObjectStore) PresignGet(_ context.Context, key string, _ time.Duration) (SignedURL, error) {
	return SignedURL{URL: "https://example.invalid/" + key, Method: http.MethodGet}, nil
}

func (s *memoryObjectStore) StartMultipart(_ context.Context, key, _ string, _ map[string]string, partCount int, _ time.Duration) (MultipartUpload, error) {
	upload := MultipartUpload{Key: key, UploadID: "upload-1"}
	for i := 1; i <= partCount; i++ {
		upload.Parts = append(upload.Parts, MultipartPart{PartNumber: i, URL: "https://example.invalid/part", Method: http.MethodPut})
	}
	return upload, nil
}

func (s *memoryObjectStore) GetObject(_ context.Context, key, byteRange string) (Object, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	object, ok := s.objects[key]
	if !ok {
		return Object{}, ErrNotFound
	}
	body := append([]byte(nil), object.Body...)
	if strings.HasPrefix(byteRange, "bytes=0-") {
		end := 0
		for _, ch := range strings.TrimPrefix(byteRange, "bytes=0-") {
			if ch < '0' || ch > '9' {
				break
			}
			end = end*10 + int(ch-'0')
		}
		if end+1 < len(body) {
			body = body[:end+1]
		}
	}
	object.Body = body
	object.Metadata = cloneMap(object.Metadata)
	return object, nil
}

func (s *memoryObjectStore) PutObject(_ context.Context, key string, body []byte, contentType string, metadata map[string]string) error {
	s.put(key, body, contentType, metadata)
	return nil
}

func (s *memoryObjectStore) DeleteObject(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.objects, key)
	return nil
}

func (s *memoryObjectStore) HeadObject(_ context.Context, key string) (ObjectInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	object, ok := s.objects[key]
	if !ok {
		return ObjectInfo{}, ErrNotFound
	}
	return ObjectInfo{
		Key:         key,
		Size:        int64(len(object.Body)),
		ContentType: object.ContentType,
		Metadata:    cloneMap(object.Metadata),
	}, nil
}

func (s *memoryObjectStore) ListObjects(_ context.Context, prefix string) ([]ObjectInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []ObjectInfo
	for key, object := range s.objects {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		out = append(out, ObjectInfo{Key: key, Size: int64(len(object.Body)), ContentType: object.ContentType, Metadata: cloneMap(object.Metadata)})
	}
	return out, nil
}

func (s *memoryObjectStore) put(key string, body []byte, contentType string, metadata map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.objects[key] = Object{
		Key:         key,
		Body:        append([]byte(nil), body...),
		ContentType: contentType,
		Metadata:    cloneMap(metadata),
	}
}

func (s *memoryObjectStore) exists(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.objects[key]
	return ok
}

func cloneMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func TestMemoryStoreNotFound(t *testing.T) {
	_, err := newMemoryObjectStore().GetObject(context.Background(), "missing", "")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestDecodeJSONRejectsInvalidPayload(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/x", bytes.NewBufferString("{"))
	var payload UploadURLRequest
	if decodeJSON(rec, req, &payload) {
		t.Fatal("decodeJSON accepted invalid JSON")
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
