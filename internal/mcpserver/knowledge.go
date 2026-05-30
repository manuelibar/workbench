package mcpserver

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type knowledgeWire struct{ ID, Kind, URI, Summary, Details, CreatedAt string }

type queryScope struct {
	NamespaceID string `json:"namespace_id,omitempty"`
	ProjectID   string `json:"project_id,omitempty"`
	Role        string `json:"role,omitempty"`
}

type queryIn struct {
	Query    string     `json:"query,omitempty"`
	Criteria string     `json:"criteria,omitempty"`
	Scope    queryScope `json:"scope,omitempty"`
	Limit    int        `json:"limit,omitempty"`
}

type queryOut struct {
	Query     string            `json:"query,omitempty"`
	Answer    string            `json:"answer,omitempty"`
	Resources []string          `json:"resources,omitempty"`
	Retrieval queryRetrievalOut `json:"retrieval,omitempty"`
	Results   []knowledgeWire   `json:"results,omitempty"`
}

type queryRetrievalOut struct {
	Used           []string                  `json:"used,omitempty"`
	ContentSearch  *kbContentSearchResponse  `json:"content_search,omitempty"`
	KnowledgeQuery *kbKnowledgeQueryResponse `json:"knowledge_query,omitempty"`
}

type KBRetriever interface {
	SearchContent(ctx context.Context, req kbContentSearchRequest) (kbContentSearchResponse, error)
	QueryKnowledge(ctx context.Context, req kbKnowledgeQueryRequest) (kbKnowledgeQueryResponse, error)
}

type QuerySynthesizer interface {
	SynthesizeQuery(ctx context.Context, req querySynthesisRequest) (querySynthesisResult, error)
}

type querySynthesisRequest struct {
	Query     string                    `json:"query"`
	Scope     queryScope                `json:"scope,omitempty"`
	Limit     int                       `json:"limit"`
	Content   *kbContentSearchResponse  `json:"content_search,omitempty"`
	Knowledge *kbKnowledgeQueryResponse `json:"knowledge_query,omitempty"`
}

type querySynthesisResult struct {
	Answer    string               `json:"answer,omitempty"`
	Resources []querySkillResource `json:"resources,omitempty"`
}

type querySkillResource struct {
	URI      string `json:"uri"`
	MIMEType string `json:"mime_type,omitempty"`
	Text     string `json:"text,omitempty"`
	Content  string `json:"content,omitempty"`
}

func (r querySkillResource) body() string {
	if r.Text != "" {
		return r.Text
	}
	return r.Content
}

func (s *Server) SetKBRetriever(retriever KBRetriever) {
	s.mu.Lock()
	s.kbRetriever = retriever
	s.mu.Unlock()
}

func (s *Server) SetQuerySynthesizer(synth QuerySynthesizer) {
	s.mu.Lock()
	s.querySynth = synth
	s.mu.Unlock()
}

func (s *Server) ingestFeedback(uri, summary, details string) KnowledgeItem {
	item := KnowledgeItem{ID: uuid.New(), Kind: "feedback", URI: uri, Summary: summary, Details: details, CreatedAt: time.Now().UTC()}
	s.mu.Lock()
	s.knowledge[item.ID] = item
	s.mu.Unlock()
	return item
}

func (s *Server) handleQuery(ctx context.Context, _ *mcp.CallToolRequest, in queryIn) (*mcp.CallToolResult, queryOut, error) {
	queryText := strings.TrimSpace(in.Query)
	if queryText == "" {
		queryText = strings.TrimSpace(in.Criteria)
	}

	s.mu.Lock()
	retriever := s.kbRetriever
	synth := s.querySynth
	s.mu.Unlock()

	if retriever == nil {
		return s.handleLocalQuery(queryText)
	}
	if synth == nil {
		return nil, queryOut{}, fmt.Errorf("query synthesizer is required when KB retrieval is configured")
	}
	return s.handleKBQuery(ctx, in, queryText, retriever, synth)
}

func (s *Server) handleLocalQuery(q string) (*mcp.CallToolResult, queryOut, error) {
	lower := strings.ToLower(strings.TrimSpace(q))
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []knowledgeWire{}
	for _, item := range s.knowledge {
		haystack := strings.ToLower(item.Summary + " " + item.Details + " " + item.URI)
		if lower == "" || strings.Contains(haystack, lower) {
			out = append(out, knowledgeToWire(item))
		}
	}
	return nil, queryOut{Query: q, Results: out}, nil
}

func (s *Server) handleKBQuery(ctx context.Context, in queryIn, queryText string, retriever KBRetriever, synth QuerySynthesizer) (*mcp.CallToolResult, queryOut, error) {
	if queryText == "" {
		return nil, queryOut{}, fmt.Errorf("query is required")
	}
	limit := normalizeQueryLimit(in.Limit)
	scope := s.resolveQueryScope(in.Scope)
	useContent, useKnowledge := decideRetrieval(queryText)
	retrieval := queryRetrievalOut{}
	reqScope := kbScope{NamespaceID: scope.NamespaceID, ProjectID: scope.ProjectID, Role: scope.Role}

	var content *kbContentSearchResponse
	if useContent {
		resp, err := retriever.SearchContent(ctx, kbContentSearchRequest{Criteria: queryText, Scope: reqScope, Limit: limit})
		if err != nil {
			return nil, queryOut{}, err
		}
		content = &resp
		retrieval.ContentSearch = &resp
		retrieval.Used = append(retrieval.Used, "content/search")
	}

	var knowledge *kbKnowledgeQueryResponse
	if useKnowledge {
		resp, err := retriever.QueryKnowledge(ctx, kbKnowledgeQueryRequest{Criteria: queryText, Scope: reqScope, Limit: limit})
		if err != nil {
			return nil, queryOut{}, err
		}
		knowledge = &resp
		retrieval.KnowledgeQuery = &resp
		retrieval.Used = append(retrieval.Used, "knowledge/query")
	}

	synthResult, err := synth.SynthesizeQuery(ctx, querySynthesisRequest{Query: queryText, Scope: scope, Limit: limit, Content: content, Knowledge: knowledge})
	if err != nil {
		return nil, queryOut{}, err
	}
	uris := s.publishQueryResources(synthResult.Resources)
	return nil, queryOut{Query: queryText, Answer: synthResult.Answer, Resources: uris, Retrieval: retrieval}, nil
}

func (s *Server) resolveQueryScope(in queryScope) queryScope {
	if in.NamespaceID != "" || in.ProjectID != "" || in.Role != "" {
		return in
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	out := queryScope{}
	if s.sel.NamespaceID != nil {
		out.NamespaceID = s.sel.NamespaceID.String()
	}
	if s.sel.ProjectID != nil {
		out.ProjectID = s.sel.ProjectID.String()
	}
	if s.sel.RoleID != nil {
		if role, ok := s.roles[*s.sel.RoleID]; ok {
			out.Role = role.Name
		} else {
			out.Role = s.sel.RoleID.String()
		}
	}
	return out
}

func decideRetrieval(queryText string) (content bool, knowledge bool) {
	lower := strings.ToLower(queryText)
	rawSignals := []string{"quote", "source", "evidence", "passage", "raw", "content", "document", "chunk"}
	knowledgeSignals := []string{"concept", "relationship", "framework", "synthesize", "knowledge", "why", "how"}
	raw := containsAny(lower, rawSignals)
	know := containsAny(lower, knowledgeSignals)
	switch {
	case raw && !know:
		return true, false
	case know && !raw:
		return false, true
	default:
		return true, true
	}
}

func containsAny(s string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(s, needle) {
			return true
		}
	}
	return false
}

func normalizeQueryLimit(limit int) int {
	if limit <= 0 {
		return 20
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func (s *Server) publishQueryResources(resources []querySkillResource) []string {
	uris := []string{}
	for _, resource := range resources {
		if resource.URI == "" {
			continue
		}
		uris = append(uris, resource.URI)
		if resource.body() == "" {
			continue
		}
		mimeType := resource.MIMEType
		if mimeType == "" {
			mimeType = "text/markdown"
		}
		resource.MIMEType = mimeType
		s.mu.Lock()
		_, existed := s.queryResources[resource.URI]
		s.queryResources[resource.URI] = resource
		s.mu.Unlock()
		if existed {
			continue
		}
		s.sdkServer.AddResource(&mcp.Resource{
			URI:         resource.URI,
			Name:        resource.URI,
			Description: "Ad hoc resource generated by Workbench query",
			MIMEType:    mimeType,
		}, s.handleAdHocSkillResource)
	}
	return uris
}

func (s *Server) handleAdHocSkillResource(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	s.mu.Lock()
	resource, ok := s.queryResources[req.Params.URI]
	s.mu.Unlock()
	if !ok {
		return nil, mcp.ResourceNotFoundError(req.Params.URI)
	}
	mimeType := resource.MIMEType
	if mimeType == "" {
		mimeType = "text/markdown"
	}
	return &mcp.ReadResourceResult{Contents: []*mcp.ResourceContents{{
		URI:      req.Params.URI,
		MIMEType: mimeType,
		Text:     resource.body(),
	}}}, nil
}

func (s *Server) handleKnowledgeResource(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []knowledgeWire{}
	for _, item := range s.knowledge {
		out = append(out, knowledgeToWire(item))
	}
	return jsonResource(req.Params.URI, map[string]any{"knowledge": out})
}

func knowledgeToWire(k KnowledgeItem) knowledgeWire {
	return knowledgeWire{ID: k.ID.String(), Kind: k.Kind, URI: k.URI, Summary: k.Summary, Details: k.Details, CreatedAt: k.CreatedAt.Format(time.RFC3339)}
}
