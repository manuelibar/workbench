package storageclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

var (
	ErrInvalid          = errors.New("storage client: invalid")
	ErrNotFound         = errors.New("storage client: not found")
	ErrDependencyFailed = errors.New("storage client: dependency failed")
)

type Error struct {
	Kind    error
	Code    string
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	if e.Kind != nil {
		return e.Kind.Error()
	}
	return "storage client error"
}

func (e *Error) Unwrap() error {
	return e.Err
}

func (e *Error) Is(target error) bool {
	return target != nil && e != nil && errors.Is(e.Kind, target)
}

type ResourceRef struct {
	OrgID        string `json:"org_id"`
	ProjectID    string `json:"project_id"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
}

func NewResourceRef(orgID, projectID, resourceType, resourceID string) (ResourceRef, error) {
	ref := ResourceRef{
		OrgID:        strings.TrimSpace(orgID),
		ProjectID:    strings.TrimSpace(projectID),
		ResourceType: strings.TrimSpace(resourceType),
		ResourceID:   strings.TrimSpace(resourceID),
	}
	if err := ref.Validate(); err != nil {
		return ResourceRef{}, err
	}
	return ref, nil
}

func (r ResourceRef) Validate() error {
	for name, value := range map[string]string{
		"org_id":        r.OrgID,
		"project_id":    r.ProjectID,
		"resource_type": r.ResourceType,
		"resource_id":   r.ResourceID,
	} {
		if strings.TrimSpace(value) == "" {
			return invalid("storage.ref.required", "%s is required", name)
		}
		if !safePathPart(value) {
			return invalid("storage.ref.invalid", "%s is invalid", name)
		}
	}
	return nil
}

func (r ResourceRef) Key() string {
	return path.Join(r.OrgID, r.ProjectID, r.ResourceType, r.ResourceID+".md")
}

func (r ResourceRef) URI() string {
	return "storage:///" + r.Key()
}

type ResourceSummary struct {
	Ref          ResourceRef `json:"resource"`
	Key          string      `json:"key"`
	URI          string      `json:"uri"`
	Size         int64       `json:"size"`
	ETag         string      `json:"etag,omitempty"`
	LastModified string      `json:"last_modified,omitempty"`
}

type SignedURL struct {
	URL       string            `json:"url"`
	Method    string            `json:"method"`
	Headers   map[string]string `json:"headers,omitempty"`
	ExpiresAt string            `json:"expires_at,omitempty"`
}

type DownloadURLResult struct {
	Ref           ResourceRef `json:"resource"`
	Key           string      `json:"key"`
	DownloadURL   SignedURL   `json:"download_url"`
	SupportsRange bool        `json:"supports_range"`
}

type UpdateURLRequest struct {
	ResourceRef
	MIMEType string `json:"mime_type"`
}

type UpdateURLResult struct {
	Ref       ResourceRef `json:"resource"`
	Key       string      `json:"key"`
	UploadURL SignedURL   `json:"upload_url"`
}

type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
}

type ClientOptions struct {
	BaseURL    string
	HTTPClient *http.Client
	Timeout    time.Duration
}

func NewClient(opts ClientOptions) (*Client, error) {
	raw := strings.TrimSpace(opts.BaseURL)
	if raw == "" {
		return nil, invalid("storage.client.base_url_required", "storage base URL is required")
	}
	baseURL, err := url.Parse(raw)
	if err != nil || baseURL.Scheme == "" || baseURL.Host == "" {
		return nil, invalid("storage.client.base_url_invalid", "storage base URL is invalid")
	}
	httpClient := opts.HTTPClient
	if httpClient == nil {
		timeout := opts.Timeout
		if timeout <= 0 {
			timeout = 30 * time.Second
		}
		httpClient = &http.Client{Timeout: timeout}
	}
	return &Client{baseURL: baseURL, httpClient: httpClient}, nil
}

func (c *Client) ListResources(ctx context.Context, orgID, projectID, resourceType string) ([]ResourceSummary, error) {
	u := c.endpoint("/api/storage/resources")
	query := u.Query()
	query.Set("org_id", strings.TrimSpace(orgID))
	query.Set("project_id", strings.TrimSpace(projectID))
	query.Set("resource_type", strings.TrimSpace(resourceType))
	u.RawQuery = query.Encode()
	var result struct {
		Resources []ResourceSummary `json:"resources"`
	}
	if err := c.doJSON(ctx, http.MethodGet, u.String(), nil, &result); err != nil {
		return nil, err
	}
	return result.Resources, nil
}

func (c *Client) DownloadMarkdown(ctx context.Context, ref ResourceRef) (string, error) {
	if err := ref.Validate(); err != nil {
		return "", err
	}
	u := c.refEndpoint(ref, "download-url")
	var result DownloadURLResult
	if err := c.doJSON(ctx, http.MethodGet, u.String(), nil, &result); err != nil {
		return "", err
	}
	method := result.DownloadURL.Method
	if method == "" {
		method = http.MethodGet
	}
	req, err := http.NewRequestWithContext(ctx, method, result.DownloadURL.URL, nil)
	if err != nil {
		return "", dependency("storage.client.download_request", "Storage download failed", err)
	}
	for key, value := range result.DownloadURL.Headers {
		req.Header.Set(key, value)
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", dependency("storage.client.download", "Storage download failed", err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return "", notFound("storage.client.download_not_found", "Storage resource not found")
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", dependency("storage.client.download_status", "Storage download failed", fmt.Errorf("unexpected status %d", res.StatusCode))
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", dependency("storage.client.download_read", "Storage download failed", err)
	}
	return string(body), nil
}

func (c *Client) PutMarkdown(ctx context.Context, ref ResourceRef, markdown string) error {
	if err := ref.Validate(); err != nil {
		return err
	}
	payload := UpdateURLRequest{
		ResourceRef: ref,
		MIMEType:    "text/markdown",
	}
	u := c.endpoint(path.Join("/api/storage", ref.ResourceID, "update-url"))
	var result UpdateURLResult
	if err := c.doJSON(ctx, http.MethodPost, u.String(), payload, &result); err != nil {
		return err
	}
	method := result.UploadURL.Method
	if method == "" {
		method = http.MethodPut
	}
	req, err := http.NewRequestWithContext(ctx, method, result.UploadURL.URL, strings.NewReader(markdown))
	if err != nil {
		return dependency("storage.client.upload_request", "Storage upload failed", err)
	}
	for key, value := range result.UploadURL.Headers {
		req.Header.Set(key, value)
	}
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "text/markdown")
	}
	req.ContentLength = int64(len(markdown))
	res, err := c.httpClient.Do(req)
	if err != nil {
		return dependency("storage.client.upload", "Storage upload failed", err)
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return dependency("storage.client.upload_status", "Storage upload failed", fmt.Errorf("unexpected status %d", res.StatusCode))
	}
	return nil
}

func (c *Client) endpoint(suffix string) *url.URL {
	u := *c.baseURL
	u.Path = path.Join(c.baseURL.Path, suffix)
	return &u
}

func (c *Client) refEndpoint(ref ResourceRef, action string) *url.URL {
	u := c.endpoint(path.Join("/api/storage", ref.ResourceID, action))
	query := u.Query()
	query.Set("org_id", ref.OrgID)
	query.Set("project_id", ref.ProjectID)
	query.Set("resource_type", ref.ResourceType)
	u.RawQuery = query.Encode()
	return u
}

func (c *Client) doJSON(ctx context.Context, method, rawURL string, payload any, out any) error {
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return dependency("storage.client.marshal", "Storage request failed", err)
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, rawURL, body)
	if err != nil {
		return dependency("storage.client.request", "Storage request failed", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return dependency("storage.client.http", "Storage request failed", err)
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		storageErr := decodeStorageError(res.Body)
		switch res.StatusCode {
		case http.StatusNotFound:
			return notFound(errorCode(storageErr), "%s", storageErr.Error())
		case http.StatusBadRequest:
			return invalid(errorCode(storageErr), "%s", storageErr.Error())
		default:
			return dependency(errorCode(storageErr), "Storage request failed", storageErr)
		}
	}
	if out == nil {
		return nil
	}
	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		return dependency("storage.client.decode", "Storage response decode failed", err)
	}
	return nil
}

func decodeStorageError(body io.Reader) error {
	var payload struct {
		Error struct {
			Title string `json:"title"`
			Code  string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(body).Decode(&payload); err != nil {
		return &Error{Kind: ErrDependencyFailed, Code: "storage.client.status", Message: "Storage request failed", Err: err}
	}
	if payload.Error.Title == "" {
		payload.Error.Title = "Storage request failed"
	}
	return &Error{Kind: ErrDependencyFailed, Code: payload.Error.Code, Message: payload.Error.Title}
}

func invalid(code, format string, args ...any) error {
	return &Error{Kind: ErrInvalid, Code: code, Message: fmt.Sprintf(format, args...)}
}

func notFound(code, format string, args ...any) error {
	return &Error{Kind: ErrNotFound, Code: code, Message: fmt.Sprintf(format, args...)}
}

func dependency(code, message string, err error) error {
	return &Error{Kind: ErrDependencyFailed, Code: code, Message: message, Err: err}
}

func errorCode(err error) string {
	var clientErr *Error
	if errors.As(err, &clientErr) && clientErr.Code != "" {
		return clientErr.Code
	}
	return "storage.client.request_failed"
}

func safePathPart(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || value == "." || value == ".." || strings.Contains(value, "/") || strings.Contains(value, "\\") {
		return false
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			continue
		}
		return false
	}
	return !strings.Contains(value, "..")
}
