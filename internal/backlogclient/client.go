package backlogclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/manuelibar/workbench/internal/id"
)

// Outbound header names. Must match backlog-service's [httpapi] constants.
const (
	HeaderRequestID     = "X-Request-Id"
	HeaderCorrelationID = "X-Correlation-Id"
	HeaderCausationID   = "X-Causation-Id"
	HeaderIdempotency   = "Idempotency-Key"
	HeaderActor         = "X-Workbench-Actor"
)

// Client is the typed HTTP client for backlog-service. Construct with [New]
// and (typically) [Client.WithActor] to set the workbench actor identity.
//
// Client is safe for concurrent use.
type Client struct {
	baseURL string
	actor   string
	http    *http.Client
}

// New returns a [Client] talking to baseURL. Use [Client.WithActor] to bind
// the workbench actor (usually `user.ID.String()`); without it, take_next
// against backlog-service will fall back to whatever assignee the caller
// passes in the request body.
func New(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// WithActor returns a copy of c with actor attached as the
// `X-Workbench-Actor` header on every outbound request.
func (c *Client) WithActor(actor string) *Client {
	cp := *c
	cp.actor = actor
	return &cp
}

// BaseURL returns the base URL the client was constructed with.
func (c *Client) BaseURL() string { return c.baseURL }

// do executes one HTTP round-trip and decodes the response into Resp. A 204
// returns (nil, nil) so callers can distinguish "no content" from a missing
// resource (404 → ErrNotFound).
func do[Resp any](ctx context.Context, c *Client, method, path string, body any) (*Resp, error) {
	var buf io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("backlogclient: marshal: %w", err)
		}
		buf = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, buf)
	if err != nil {
		return nil, fmt.Errorf("backlogclient: new request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if a, ok := id.FromContext(ctx); ok {
		if a.RequestID != uuid.Nil {
			req.Header.Set(HeaderRequestID, a.RequestID.String())
		}
		if a.CorrelationID != uuid.Nil {
			req.Header.Set(HeaderCorrelationID, a.CorrelationID.String())
		}
		if a.CausationID != uuid.Nil {
			req.Header.Set(HeaderCausationID, a.CausationID.String())
		}
		if a.IdempotencyKey != "" {
			req.Header.Set(HeaderIdempotency, a.IdempotencyKey)
		}
	}
	if c.actor != "" {
		req.Header.Set(HeaderActor, c.actor)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("backlogclient: %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}
	if resp.StatusCode >= 400 {
		var er struct {
			Error string `json:"error"`
			Code  string `json:"code"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&er)
		switch resp.StatusCode {
		case http.StatusNotFound:
			return nil, fmt.Errorf("%w: %s", ErrNotFound, er.Error)
		case http.StatusConflict:
			return nil, fmt.Errorf("%w: %s", ErrVersionConflict, er.Error)
		case http.StatusBadRequest:
			return nil, fmt.Errorf("%w: %s", ErrValidation, er.Error)
		default:
			return nil, fmt.Errorf("%w: %d %s", ErrServer, resp.StatusCode, er.Error)
		}
	}
	var out Resp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		if err == io.EOF {
			return &out, nil
		}
		return nil, fmt.Errorf("backlogclient: decode: %w", err)
	}
	return &out, nil
}

// Create creates an issue. Returns the inserted (or idempotency-replayed)
// row.
func (c *Client) Create(ctx context.Context, req IssueCreateRequest) (*Issue, error) {
	return do[Issue](ctx, c, http.MethodPost, "/v1/issues", req)
}

// Get fetches one issue by id.
func (c *Client) Get(ctx context.Context, id string) (*Issue, error) {
	return do[Issue](ctx, c, http.MethodGet, "/v1/issues/"+url.PathEscape(id), nil)
}

// Update applies a partial update to an issue. expected_version on req
// must match the server's current version; otherwise the call returns
// [ErrVersionConflict].
func (c *Client) Update(ctx context.Context, id string, req IssueUpdateRequest) (*Issue, error) {
	return do[Issue](ctx, c, http.MethodPatch, "/v1/issues/"+url.PathEscape(id), req)
}

// Delete removes an issue. Returns [ErrNotFound] when no such id exists.
func (c *Client) Delete(ctx context.Context, id string) error {
	_, err := do[struct{}](ctx, c, http.MethodDelete, "/v1/issues/"+url.PathEscape(id), nil)
	return err
}

// IssueListQuery is the typed query for [Client.List].
type IssueListQuery struct {
	ProjectID string
	Status    string
	Type      string
	Priority  string
	Label     string
	ParentID  string
	Assignee  string
	SourceRef string
	Q         string
	Order     string
	Limit     int
	Offset    int
}

// Encode returns the URL-encoded query string (without leading "?").
func (q IssueListQuery) Encode() string {
	v := url.Values{}
	setIf(v, "project_id", q.ProjectID)
	setIf(v, "status", q.Status)
	setIf(v, "type", q.Type)
	setIf(v, "priority", q.Priority)
	setIf(v, "label", q.Label)
	setIf(v, "parent_id", q.ParentID)
	setIf(v, "assignee", q.Assignee)
	setIf(v, "source_ref", q.SourceRef)
	setIf(v, "q", q.Q)
	setIf(v, "order", q.Order)
	if q.Limit > 0 {
		v.Set("limit", strconv.Itoa(q.Limit))
	}
	if q.Offset > 0 {
		v.Set("offset", strconv.Itoa(q.Offset))
	}
	return v.Encode()
}

// List returns issues matching q.
func (c *Client) List(ctx context.Context, q IssueListQuery) (*IssueListResponse, error) {
	path := "/v1/issues"
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}
	return do[IssueListResponse](ctx, c, http.MethodGet, path, nil)
}

// TakeNext atomically claims the next available issue. Returns
// (nil, nil) when there's nothing to claim (HTTP 204).
func (c *Client) TakeNext(ctx context.Context, req TakeNextRequest) (*Issue, error) {
	return do[Issue](ctx, c, http.MethodPost, "/v1/issues:take_next", req)
}

// EventsQuery is the typed query for [Client.ListEvents].
type EventsQuery struct {
	SubjectKind string
	SubjectID   string
	Type        string
	Actor       string
	ProjectID   string
	Since       time.Time
	Until       time.Time
	Limit       int
	Offset      int
}

// Encode returns the URL-encoded query string (without leading "?").
func (q EventsQuery) Encode() string {
	v := url.Values{}
	setIf(v, "subject_kind", q.SubjectKind)
	setIf(v, "subject_id", q.SubjectID)
	setIf(v, "type", q.Type)
	setIf(v, "actor", q.Actor)
	setIf(v, "project_id", q.ProjectID)
	if !q.Since.IsZero() {
		v.Set("since", q.Since.Format(time.RFC3339Nano))
	}
	if !q.Until.IsZero() {
		v.Set("until", q.Until.Format(time.RFC3339Nano))
	}
	if q.Limit > 0 {
		v.Set("limit", strconv.Itoa(q.Limit))
	}
	if q.Offset > 0 {
		v.Set("offset", strconv.Itoa(q.Offset))
	}
	return v.Encode()
}

// ListEvents returns events matching q.
func (c *Client) ListEvents(ctx context.Context, q EventsQuery) (*EventListResponse, error) {
	path := "/v1/events"
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}
	return do[EventListResponse](ctx, c, http.MethodGet, path, nil)
}

func setIf(v url.Values, key, val string) {
	if val != "" {
		v.Set(key, val)
	}
}
