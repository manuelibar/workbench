package mcpserver

import (
	"io"
	"log/slog"
	"testing"

	"github.com/manuelibar/workbench/internal/mcpserver/skills"
)

func TestFeedbackBecomesAskableKnowledge(t *testing.T) {
	s := New(slog.New(slog.NewTextHandler(io.Discard, nil)), NewMemProjectStore(), skills.NewEmbeddedRegistry())
	s.ingestFeedback("skill://go-coding-guidelines/SKILL.md", "Prefer gofmt", "Run gofmt before tests")
	_, out, err := s.handleAsk(nil, nil, askIn{Query: "gofmt"})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Results) != 1 || out.Results[0].Summary != "Prefer gofmt" {
		t.Fatalf("results = %#v", out.Results)
	}
}
