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

type askScope struct {
	NamespaceID string `json:"namespace_id,omitempty"`
	ProjectID   string `json:"project_id,omitempty"`
	Role        string `json:"role,omitempty"`
}

type askIn struct {
	Query    string   `json:"query,omitempty"`
	Criteria string   `json:"criteria,omitempty"`
	Scope    askScope `json:"scope,omitempty"`
	Limit    int      `json:"limit,omitempty"`
}

type askOut struct {
	Criteria  string          `json:"criteria,omitempty"`
	Answer    string          `json:"answer,omitempty"`
	Resources []string        `json:"resources,omitempty"`
	Retrieval askRetrievalOut `json:"retrieval,omitempty"`
	Results   []knowledgeWire `json:"results,omitempty"`
}

type askRetrievalOut struct {
	Used           []string                  `json:"used,omitempty"`
	ContentSearch  *kbContentSearchResponse  `json:"content_search,omitempty"`
	KnowledgeQuery *kbKnowledgeQueryResponse `json:"knowledge_query,omitempty"`
}

type KBRetriever interface {
	SearchContent(ctx context.Context, req kbContentSearchRequest) (kbContentSearchResponse, error)
	QueryKnowledge(ctx context.Context, req kbKnowledgeQueryRequest) (kbKnowledgeQueryResponse, error)
}

type AskSynthesizer interface {
	SynthesizeAsk(ctx context.Context, req askSynthesisRequest) (askSynthesisResult, error)
}

type askSynthesisRequest struct {
	Criteria  string                    `json:"criteria"`
	Scope     askScope                  `json:"scope,omitempty"`
	Limit     int                       `json:"limit"`
	Content   *kbContentSearchResponse  `json:"content_search,omitempty"`
	Knowledge *kbKnowledgeQueryResponse `json:"knowledge_query,omitempty"`
}

type askSynthesisResult struct {
	Answer    string             `json:"answer,omitempty"`
	Resources []askSkillResource `json:"resources,omitempty"`
}

type askSkillResource struct {
	URI      string `json:"uri"`
	MIMEType string `json:"mime_type,omitempty"`
	Text     string `json:"text,omitempty"`
	Content  string `json:"content,omitempty"`
}

func (r askSkillResource) body() string {
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

func (s *Server) SetAskSynthesizer(synth AskSynthesizer) {
	s.mu.Lock()
	s.askSynth = synth
	s.mu.Unlock()
}

func (s *Server) ingestFeedback(uri, summary, details string) KnowledgeItem {
	item := KnowledgeItem{ID: uuid.New(), Kind: "feedback", URI: uri, Summary: summary, Details: details, CreatedAt: time.Now().UTC()}
	s.mu.Lock()
	s.knowledge[item.ID] = item
	s.mu.Unlock()
	return item
}

func (s *Server) handleAsk(ctx context.Context, _ *mcp.CallToolRequest, in askIn) (*mcp.CallToolResult, askOut, error) {
	criteria := strings.TrimSpace(in.Criteria)
	if criteria == "" {
		criteria = strings.TrimSpace(in.Query)
	}

	s.mu.Lock()
	retriever := s.kbRetriever
	synth := s.askSynth
	s.mu.Unlock()

	if retriever == nil {
		return s.handleLocalAsk(criteria)
	}
	if synth == nil {
		return nil, askOut{}, fmt.Errorf("ask synthesizer is required when KB retrieval is configured")
	}
	return s.handleKBAsk(ctx, in, criteria, retriever, synth)
}

func (s *Server) handleLocalAsk(q string) (*mcp.CallToolResult, askOut, error) {
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
	return nil, askOut{Criteria: q, Results: out}, nil
}

func (s *Server) handleKBAsk(ctx context.Context, in askIn, criteria string, retriever KBRetriever, synth AskSynthesizer) (*mcp.CallToolResult, askOut, error) {
	if criteria == "" {
		return nil, askOut{}, fmt.Errorf("criteria is required")
	}
	limit := normalizeAskLimit(in.Limit)
	scope := s.resolveAskScope(in.Scope)
	useContent, useKnowledge := decideRetrieval(criteria)
	retrieval := askRetrievalOut{}
	reqScope := kbScope{NamespaceID: scope.NamespaceID, ProjectID: scope.ProjectID, Role: scope.Role}

	var content *kbContentSearchResponse
	if useContent {
		resp, err := retriever.SearchContent(ctx, kbContentSearchRequest{Criteria: criteria, Scope: reqScope, Limit: limit})
		if err != nil {
			return nil, askOut{}, err
		}
		content = &resp
		retrieval.ContentSearch = &resp
		retrieval.Used = append(retrieval.Used, "content/search")
	}

	var knowledge *kbKnowledgeQueryResponse
	if useKnowledge {
		resp, err := retriever.QueryKnowledge(ctx, kbKnowledgeQueryRequest{Criteria: criteria, Scope: reqScope, Limit: limit})
		if err != nil {
			return nil, askOut{}, err
		}
		knowledge = &resp
		retrieval.KnowledgeQuery = &resp
		retrieval.Used = append(retrieval.Used, "knowledge/query")
	}

	synthResult, err := synth.SynthesizeAsk(ctx, askSynthesisRequest{Criteria: criteria, Scope: scope, Limit: limit, Content: content, Knowledge: knowledge})
	if err != nil {
		return nil, askOut{}, err
	}
	uris := s.publishAskResources(synthResult.Resources)
	return nil, askOut{Criteria: criteria, Answer: synthResult.Answer, Resources: uris, Retrieval: retrieval}, nil
}

func (s *Server) resolveAskScope(in askScope) askScope {
	if in.NamespaceID != "" || in.ProjectID != "" || in.Role != "" {
		return in
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	out := askScope{}
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

func decideRetrieval(criteria string) (content bool, knowledge bool) {
	lower := strings.ToLower(criteria)
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

func normalizeAskLimit(limit int) int {
	if limit <= 0 {
		return 20
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func (s *Server) publishAskResources(resources []askSkillResource) []string {
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
		_, existed := s.askResources[resource.URI]
		s.askResources[resource.URI] = resource
		s.mu.Unlock()
		if existed {
			continue
		}
		s.sdkServer.AddResource(&mcp.Resource{
			URI:         resource.URI,
			Name:        resource.URI,
			Description: "Ad hoc resource generated by Workbench ask",
			MIMEType:    mimeType,
		}, s.handleAdHocSkillResource)
	}
	return uris
}

func (s *Server) handleAdHocSkillResource(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	s.mu.Lock()
	resource, ok := s.askResources[req.Params.URI]
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
