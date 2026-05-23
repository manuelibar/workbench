package mcpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type kbScope struct {
	NamespaceID string `json:"namespace_id,omitempty"`
	ProjectID   string `json:"project_id,omitempty"`
	Role        string `json:"role,omitempty"`
}

type kbContentSearchRequest struct {
	Criteria string  `json:"criteria"`
	Scope    kbScope `json:"scope,omitempty"`
	Limit    int     `json:"limit,omitempty"`
}

type kbKnowledgeQueryRequest struct {
	Criteria string  `json:"criteria"`
	Scope    kbScope `json:"scope,omitempty"`
	Limit    int     `json:"limit,omitempty"`
}

type kbContentSearchResponse struct {
	Criteria string           `json:"criteria"`
	Matches  []kbContentMatch `json:"matches"`
	Stats    kbSearchStats    `json:"stats"`
}

type kbKnowledgeQueryResponse struct {
	Criteria string             `json:"criteria"`
	Matches  []kbKnowledgeMatch `json:"matches"`
	Stats    kbKnowledgeStats   `json:"stats"`
}

type kbContentMatch struct {
	ID            string    `json:"id"`
	DocumentURI   string    `json:"document_uri"`
	DocumentTitle string    `json:"document_title"`
	Text          string    `json:"text"`
	Score         float64   `json:"score"`
	TokenCount    int       `json:"token_count"`
	Locator       kbLocator `json:"locator,omitempty"`
}

type kbKnowledgeMatch struct {
	ID       string            `json:"id"`
	Kind     string            `json:"kind"`
	Title    string            `json:"title"`
	Summary  string            `json:"summary"`
	Score    float64           `json:"score"`
	Evidence []kbEvidenceChunk `json:"evidence,omitempty"`
}

type kbEvidenceChunk struct {
	ChunkID       string    `json:"chunk_id"`
	DocumentURI   string    `json:"document_uri"`
	DocumentTitle string    `json:"document_title"`
	Text          string    `json:"text"`
	Score         float64   `json:"score"`
	TokenCount    int       `json:"token_count"`
	Locator       kbLocator `json:"locator,omitempty"`
}

type kbLocator struct {
	PageStart int `json:"page_start,omitempty"`
	PageEnd   int `json:"page_end,omitempty"`
	LineStart int `json:"line_start,omitempty"`
	LineEnd   int `json:"line_end,omitempty"`
}

type kbSearchStats struct {
	MatchCount int `json:"match_count"`
	TokenCount int `json:"token_count"`
}

type kbKnowledgeStats struct {
	MatchCount    int `json:"match_count"`
	EvidenceCount int `json:"evidence_count"`
	TokenCount    int `json:"token_count"`
}

type HTTPKBRetriever struct {
	BaseURL string
	Client  *http.Client
}

func NewHTTPKBRetriever(baseURL string) *HTTPKBRetriever {
	return &HTTPKBRetriever{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Client:  &http.Client{Timeout: 20 * time.Second},
	}
}

func (c *HTTPKBRetriever) SearchContent(ctx context.Context, req kbContentSearchRequest) (kbContentSearchResponse, error) {
	var out kbContentSearchResponse
	err := c.postJSON(ctx, "/content/search", req, &out)
	return out, err
}

func (c *HTTPKBRetriever) QueryKnowledge(ctx context.Context, req kbKnowledgeQueryRequest) (kbKnowledgeQueryResponse, error) {
	var out kbKnowledgeQueryResponse
	err := c.postJSON(ctx, "/knowledge/query", req, &out)
	return out, err
}

func (c *HTTPKBRetriever) postJSON(ctx context.Context, path string, in, out any) error {
	if c.BaseURL == "" {
		return fmt.Errorf("KB base URL is required")
	}
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(in); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+path, &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := c.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		var problem map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&problem)
		if problem["error"] != "" {
			return fmt.Errorf("KB %s failed: %s", path, problem["error"])
		}
		return fmt.Errorf("KB %s failed with status %d", path, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
