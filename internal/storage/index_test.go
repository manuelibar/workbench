package storage

import (
	"strings"
	"testing"
	"time"
)

func TestIndexMarkdownPreservesFrontmatterAndBuildsFinalByteOffsets(t *testing.T) {
	ref := ResourceRef{
		OrgID:        "acme",
		ProjectID:    "workbench",
		ResourceType: "artifacts",
		ResourceID:   "artifact-1",
	}
	indexed, err := IndexMarkdown(ref, `---
id: artifact-1
type: rfc
title: Storage RFC
---

# Storage RFC

Summary body.

## Security

Security body.
`, "text/markdown", time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(indexed.Markdown, "---\n") {
		t.Fatalf("indexed markdown missing frontmatter:\n%s", indexed.Markdown)
	}
	if !strings.Contains(indexed.Markdown, "id: artifact-1") ||
		!strings.Contains(indexed.Markdown, "type: rfc") {
		t.Fatalf("artifact frontmatter was not preserved:\n%s", indexed.Markdown)
	}
	stats, err := ParseStats(indexed.Markdown[:min(len(indexed.Markdown), 8192)])
	if err != nil {
		t.Fatal(err)
	}
	if stats.Resource.ByteLength != len(indexed.Markdown) {
		t.Fatalf("byte length = %d, want %d", stats.Resource.ByteLength, len(indexed.Markdown))
	}
	if got := len(stats.Index.Sections); got != 2 {
		t.Fatalf("section count = %d, want 2", got)
	}
	if err := validateIndexOffsets(indexed.Markdown, stats.Index.Sections); err != nil {
		t.Fatal(err)
	}
}
