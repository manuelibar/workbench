package mcpserver_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

// backlogStub is a minimal in-memory implementation of backlog-service's
// HTTP surface, intended for wire-contract testing of internal/backlogclient
// and the workbench backlog.* MCP tools. It is NOT a faithful Postgres-backed
// replica — concurrency, OCC, and idempotency are modelled just enough for
// the tools to exercise their happy paths and one or two error paths.
type backlogStub struct {
	server *httptest.Server

	mu     sync.Mutex
	issues map[string]*stubIssue

	lastActorVal string
}

type stubIssue struct {
	ID         string    `json:"id"`
	ProjectID  string    `json:"project_id"`
	Type       string    `json:"type"`
	Status     string    `json:"status"`
	Priority   string    `json:"priority"`
	Title      string    `json:"title"`
	Body       string    `json:"body"`
	Labels     []string  `json:"labels"`
	ParentID   string    `json:"parent_id,omitempty"`
	Assignee   string    `json:"assignee,omitempty"`
	SourceRefs []string  `json:"source_refs"`
	Version    int       `json:"version"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func newBacklogStub(t *testing.T) *backlogStub {
	t.Helper()
	stub := &backlogStub{issues: map[string]*stubIssue{}}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/issues", stub.handleCreate)
	mux.HandleFunc("GET /v1/issues", stub.handleList)
	mux.HandleFunc("GET /v1/issues/{id}", stub.handleGet)
	mux.HandleFunc("PATCH /v1/issues/{id}", stub.handleUpdate)
	mux.HandleFunc("DELETE /v1/issues/{id}", stub.handleDelete)
	mux.HandleFunc("POST /v1/issues:take_next", stub.handleTakeNext)
	stub.server = httptest.NewServer(mux)
	t.Cleanup(stub.server.Close)
	return stub
}

func (s *backlogStub) lastActor() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastActorVal
}

func (s *backlogStub) recordActor(r *http.Request) {
	if a := r.Header.Get("X-Workbench-Actor"); a != "" {
		s.mu.Lock()
		s.lastActorVal = a
		s.mu.Unlock()
	}
}

func (s *backlogStub) handleCreate(w http.ResponseWriter, r *http.Request) {
	s.recordActor(r)
	var req struct {
		ProjectID  string   `json:"project_id"`
		Title      string   `json:"title"`
		Body       string   `json:"body"`
		Type       string   `json:"type"`
		Status     string   `json:"status"`
		Priority   string   `json:"priority"`
		Labels     []string `json:"labels"`
		ParentID   string   `json:"parent_id"`
		Assignee   string   `json:"assignee"`
		SourceRefs []string `json:"source_refs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Title == "" || req.ProjectID == "" {
		http.Error(w, "title and project_id required", http.StatusBadRequest)
		return
	}
	now := time.Now().UTC()
	issue := &stubIssue{
		ID:        uuid.New().String(),
		ProjectID: req.ProjectID,
		Type:      defaultIf(req.Type, "task"),
		Status:    defaultIf(req.Status, "todo"),
		Priority:  defaultIf(req.Priority, "med"),
		Title:     req.Title,
		Body:      req.Body,
		Labels:    lowerSlice(req.Labels),
		ParentID:  req.ParentID,
		Assignee:  req.Assignee,
		SourceRefs: emptyIfNil(req.SourceRefs),
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.mu.Lock()
	s.issues[issue.ID] = issue
	s.mu.Unlock()
	writeJSON(w, http.StatusCreated, issue)
}

func (s *backlogStub) handleList(w http.ResponseWriter, r *http.Request) {
	s.recordActor(r)
	q := r.URL.Query()
	s.mu.Lock()
	defer s.mu.Unlock()
	items := []*stubIssue{}
	for _, i := range s.issues {
		if v := q.Get("project_id"); v != "" && i.ProjectID != v {
			continue
		}
		if v := q.Get("status"); v != "" && i.Status != v {
			continue
		}
		if v := q.Get("priority"); v != "" && i.Priority != v {
			continue
		}
		if v := q.Get("type"); v != "" && i.Type != v {
			continue
		}
		if v := q.Get("assignee"); v != "" && i.Assignee != v {
			continue
		}
		if v := q.Get("label"); v != "" {
			v = strings.ToLower(v)
			found := false
			for _, l := range i.Labels {
				if l == v {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		if v := q.Get("source_ref"); v != "" {
			found := false
			for _, ref := range i.SourceRefs {
				if ref == v {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		items = append(items, i)
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *backlogStub) handleGet(w http.ResponseWriter, r *http.Request) {
	s.recordActor(r)
	id := r.PathValue("id")
	s.mu.Lock()
	defer s.mu.Unlock()
	i, ok := s.issues[id]
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "no such issue")
		return
	}
	writeJSON(w, http.StatusOK, i)
}

func (s *backlogStub) handleUpdate(w http.ResponseWriter, r *http.Request) {
	s.recordActor(r)
	id := r.PathValue("id")
	var req struct {
		ExpectedVersion int       `json:"expected_version"`
		Title           *string   `json:"title"`
		Body            *string   `json:"body"`
		Status          *string   `json:"status"`
		Type            *string   `json:"type"`
		Priority        *string   `json:"priority"`
		Labels          *[]string `json:"labels"`
		ParentID        *string   `json:"parent_id"`
		Assignee        *string   `json:"assignee"`
		SourceRefs      *[]string `json:"source_refs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	i, ok := s.issues[id]
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "no such issue")
		return
	}
	if req.ExpectedVersion != 0 && req.ExpectedVersion != i.Version {
		writeError(w, http.StatusConflict, "version_conflict", "version mismatch")
		return
	}
	if req.Title != nil {
		i.Title = *req.Title
	}
	if req.Body != nil {
		i.Body = *req.Body
	}
	if req.Status != nil {
		i.Status = *req.Status
	}
	if req.Type != nil {
		i.Type = *req.Type
	}
	if req.Priority != nil {
		i.Priority = *req.Priority
	}
	if req.Labels != nil {
		i.Labels = lowerSlice(*req.Labels)
	}
	if req.ParentID != nil {
		i.ParentID = *req.ParentID
	}
	if req.Assignee != nil {
		i.Assignee = *req.Assignee
	}
	if req.SourceRefs != nil {
		i.SourceRefs = emptyIfNil(*req.SourceRefs)
	}
	i.Version++
	i.UpdatedAt = time.Now().UTC()
	writeJSON(w, http.StatusOK, i)
}

func (s *backlogStub) handleDelete(w http.ResponseWriter, r *http.Request) {
	s.recordActor(r)
	id := r.PathValue("id")
	s.mu.Lock()
	_, ok := s.issues[id]
	if ok {
		delete(s.issues, id)
	}
	s.mu.Unlock()
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "no such issue")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *backlogStub) handleTakeNext(w http.ResponseWriter, r *http.Request) {
	s.recordActor(r)
	actor := r.Header.Get("X-Workbench-Actor")
	var req struct {
		ProjectID string `json:"project_id"`
		Assignee  string `json:"assignee"`
	}
	if r.ContentLength != 0 {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	// pick oldest todo issue (priority ignored in this stub for simplicity)
	var pick *stubIssue
	for _, i := range s.issues {
		if i.Status != "todo" {
			continue
		}
		if req.ProjectID != "" && i.ProjectID != req.ProjectID {
			continue
		}
		if pick == nil || i.CreatedAt.Before(pick.CreatedAt) {
			pick = i
		}
	}
	if pick == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	pick.Status = "in_progress"
	if pick.Assignee == "" {
		if req.Assignee != "" {
			pick.Assignee = req.Assignee
		} else {
			pick.Assignee = actor
		}
	}
	pick.Version++
	pick.UpdatedAt = time.Now().UTC()
	writeJSON(w, http.StatusOK, pick)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, map[string]string{"error": msg, "code": code})
}

func defaultIf(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

func lowerSlice(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		out = append(out, strings.ToLower(s))
	}
	return out
}

func emptyIfNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
