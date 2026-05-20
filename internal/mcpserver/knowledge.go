package mcpserver

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type knowledgeWire struct{ ID, Kind, URI, Summary, Details, CreatedAt string }
type askIn struct {
	Query string `json:"query"`
}
type askOut struct {
	Results []knowledgeWire `json:"results"`
}

func (s *Server) ingestFeedback(uri, summary, details string) KnowledgeItem {
	item := KnowledgeItem{ID: uuid.New(), Kind: "feedback", URI: uri, Summary: summary, Details: details, CreatedAt: time.Now().UTC()}
	s.mu.Lock()
	s.knowledge[item.ID] = item
	s.mu.Unlock()
	return item
}

func (s *Server) handleAsk(_ context.Context, _ *mcp.CallToolRequest, in askIn) (*mcp.CallToolResult, askOut, error) {
	q := strings.ToLower(strings.TrimSpace(in.Query))
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []knowledgeWire{}
	for _, item := range s.knowledge {
		haystack := strings.ToLower(item.Summary + " " + item.Details + " " + item.URI)
		if q == "" || strings.Contains(haystack, q) {
			out = append(out, knowledgeToWire(item))
		}
	}
	return nil, askOut{Results: out}, nil
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
