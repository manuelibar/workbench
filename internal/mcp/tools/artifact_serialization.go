package tools

import (
	"maps"

	"github.com/manuelibar/workbench/internal/artifacts"
)

type artifactSummaryPayload struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Title   string `json:"title"`
	Status  string `json:"status"`
	Created string `json:"created"`
	Updated string `json:"updated"`
	Path    string `json:"path"`
}

type artifactPayload struct {
	artifactSummaryPayload
	Markdown string            `json:"markdown"`
	Sections map[string]string `json:"sections,omitempty"`
}

func artifactSummaryPayloadsFrom(summaries []artifacts.Summary) []artifactSummaryPayload {
	out := make([]artifactSummaryPayload, 0, len(summaries))
	for _, summary := range summaries {
		out = append(out, artifactSummaryPayloadFrom(summary))
	}
	return out
}

func artifactSummaryPayloadFrom(summary artifacts.Summary) artifactSummaryPayload {
	return artifactSummaryPayload{
		ID:      summary.ID,
		Type:    summary.Type,
		Title:   summary.Title,
		Status:  summary.Status,
		Created: summary.Created,
		Updated: summary.Updated,
		Path:    summary.Path,
	}
}

func artifactPayloadFrom(artifact artifacts.Artifact) artifactPayload {
	return artifactPayload{
		artifactSummaryPayload: artifactSummaryPayloadFrom(artifact.Summary),
		Markdown:               artifact.Markdown,
		Sections:               maps.Clone(artifact.Sections),
	}
}
