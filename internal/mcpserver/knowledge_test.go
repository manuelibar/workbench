package mcpserver

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/manuelibar/workbench/internal/mcpserver/skills"
)

func TestFeedbackBecomesQueryableKnowledge(t *testing.T) {
	s := New(slog.New(slog.NewTextHandler(io.Discard, nil)), NewMemProjectStore(), skills.NewEmbeddedRegistry())
	s.ingestFeedback("skill://go-coding-guidelines/SKILL.md", "Prefer gofmt", "Run gofmt before tests")
	_, out, err := s.handleQuery(nil, nil, queryIn{Query: "gofmt"})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Results) != 1 || out.Results[0].Summary != "Prefer gofmt" {
		t.Fatalf("results = %#v", out.Results)
	}
}

func TestQueryUsesKBRetrievalPrimitivesAndSynthesizesOnce(t *testing.T) {
	s := New(slog.New(slog.NewTextHandler(io.Discard, nil)), NewMemProjectStore(), skills.NewEmbeddedRegistry())

	tests := []struct {
		name          string
		criteria      string
		wantContent   int
		wantKnowledge int
	}{
		{name: "both", criteria: "agent memory", wantContent: 1, wantKnowledge: 1},
		{name: "content only", criteria: "raw source passage", wantContent: 1, wantKnowledge: 0},
		{name: "knowledge only", criteria: "concept framework", wantContent: 0, wantKnowledge: 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			retriever := &fakeKBRetriever{}
			synth := &fakeQuerySynthesizer{result: querySynthesisResult{Answer: "grounded answer"}}
			s.SetKBRetriever(retriever)
			s.SetQuerySynthesizer(synth)

			_, out, err := s.handleQuery(context.Background(), nil, queryIn{
				Criteria: tc.criteria,
				Scope:    queryScope{NamespaceID: "acme", ProjectID: "platform", Role: "coder"},
				Limit:    7,
			})
			if err != nil {
				t.Fatal(err)
			}
			if out.Answer != "grounded answer" {
				t.Fatalf("answer = %q", out.Answer)
			}
			if retriever.contentCalls != tc.wantContent || retriever.knowledgeCalls != tc.wantKnowledge {
				t.Fatalf("calls content=%d knowledge=%d", retriever.contentCalls, retriever.knowledgeCalls)
			}
			if synth.calls != 1 {
				t.Fatalf("synth calls = %d", synth.calls)
			}
			if synth.last.Query != tc.criteria || synth.last.Limit != 7 {
				t.Fatalf("synth request = %#v", synth.last)
			}
		})
	}
}

func TestQueryPublishesAdHocSkillResources(t *testing.T) {
	s := New(slog.New(slog.NewTextHandler(io.Discard, nil)), NewMemProjectStore(), skills.NewEmbeddedRegistry())
	s.SetKBRetriever(&fakeKBRetriever{})
	s.SetQuerySynthesizer(&fakeQuerySynthesizer{result: querySynthesisResult{Resources: []querySkillResource{{
		URI:      "skill://agent-memory/SKILL.md",
		MIMEType: "text/markdown",
		Text:     "# Agent Memory\n",
	}}}})

	_, out, err := s.handleQuery(context.Background(), nil, queryIn{Query: "agent memory"})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Resources) != 1 || out.Resources[0] != "skill://agent-memory/SKILL.md" {
		t.Fatalf("resources = %#v", out.Resources)
	}

	res, err := s.handleAdHocSkillResource(context.Background(), &mcp.ReadResourceRequest{Params: &mcp.ReadResourceParams{URI: "skill://agent-memory/SKILL.md"}})
	if err != nil {
		t.Fatal(err)
	}
	if got := res.Contents[0].Text; got != "# Agent Memory\n" {
		t.Fatalf("resource text = %q", got)
	}
}

type fakeKBRetriever struct {
	contentCalls   int
	knowledgeCalls int
}

func (f *fakeKBRetriever) SearchContent(_ context.Context, req kbContentSearchRequest) (kbContentSearchResponse, error) {
	f.contentCalls++
	return kbContentSearchResponse{
		Criteria: req.Criteria,
		Matches: []kbContentMatch{{
			ID:            "chunk_01J_0007",
			DocumentURI:   "kb:document:01J",
			DocumentTitle: "Memory Frameworks for Agents",
			Text:          "Agent memory systems usually separate working memory.",
			Score:         0.92,
			TokenCount:    8,
		}},
		Stats: kbSearchStats{MatchCount: 1, TokenCount: 8},
	}, nil
}

func (f *fakeKBRetriever) QueryKnowledge(_ context.Context, req kbKnowledgeQueryRequest) (kbKnowledgeQueryResponse, error) {
	f.knowledgeCalls++
	return kbKnowledgeQueryResponse{
		Criteria: req.Criteria,
		Matches: []kbKnowledgeMatch{{
			ID:      "concept_agent_memory",
			Kind:    "concept",
			Title:   "Agent Memory",
			Summary: "Agent memory preserves working context and durable knowledge.",
			Score:   0.91,
		}},
		Stats: kbKnowledgeStats{MatchCount: 1},
	}, nil
}

type fakeQuerySynthesizer struct {
	calls  int
	last   querySynthesisRequest
	result querySynthesisResult
}

func (f *fakeQuerySynthesizer) SynthesizeQuery(_ context.Context, req querySynthesisRequest) (querySynthesisResult, error) {
	f.calls++
	f.last = req
	return f.result, nil
}
