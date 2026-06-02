package tools

import (
	"context"

	"github.com/manuelibar/workbench/internal/errs"
)

type contextualizeTool struct{}

type CapabilityKind string

const (
	CapabilityTool     CapabilityKind = "tool"
	CapabilityResource CapabilityKind = "resource"
	CapabilityPrompt   CapabilityKind = "prompt"
)

type CapabilitySyncStatus struct {
	Generation int64    `json:"generation"`
	Status     string   `json:"status"`
	Required   []string `json:"required"`
	Observed   []string `json:"observed"`
	TimedOut   bool     `json:"timed_out"`
}

type CapabilitySummary struct {
	ID          string         `json:"-"`
	Kind        CapabilityKind `json:"kind"`
	Name        string         `json:"name,omitempty"`
	URI         string         `json:"uri,omitempty"`
	Description string         `json:"description"`
	Group       string         `json:"group"`
	Active      bool           `json:"active"`
}

type CapabilitySurface struct {
	Tools     []CapabilitySummary `json:"tools"`
	Resources []CapabilitySummary `json:"resources"`
	Prompts   []CapabilitySummary `json:"prompts"`
}

type ContextualizeResult struct {
	ContextDocument      string               `json:"context_document"`
	Focus                *string              `json:"focus,omitempty"`
	ArtifactID           *string              `json:"artifact_id,omitempty"`
	Sync                 CapabilitySyncStatus `json:"sync"`
	FallbackCapabilities *CapabilitySurface   `json:"fallback_capabilities,omitempty"`
}

func init() {
	register[map[string]any, ContextualizeResult](contextualizeTool{})
}

func (contextualizeTool) Name() string {
	return "contextualize"
}

func (contextualizeTool) Group() string {
	return ""
}

func (contextualizeTool) Description() string {
	return "Read or patch focus/artifact context. Optional inputs: omit focus or artifact_id to preserve it, set a string to update it, or set null to clear it."
}

func (contextualizeTool) Handle(ctx context.Context, host Host, args map[string]any) (ContextualizeResult, error) {
	attrs := map[string]any{"tool": "contextualize"}
	result, err := host.ApplyContextPatch(ctx, args)
	if err != nil {
		return ContextualizeResult{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	return result, nil
}
