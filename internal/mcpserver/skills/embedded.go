package skills

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed seeds
var seedFS embed.FS

// EmbeddedRegistry is a SkillRegistry backed by the embedded seed bundles.
type EmbeddedRegistry struct {
	bundles []Bundle
}

// NewEmbeddedRegistry loads the seed bundles and returns an EmbeddedRegistry.
func NewEmbeddedRegistry() *EmbeddedRegistry {
	return &EmbeddedRegistry{bundles: buildSeedBundles()}
}

func (r *EmbeddedRegistry) All() []Bundle { return r.bundles }

func (r *EmbeddedRegistry) Get(name string) (Bundle, bool) {
	for _, b := range r.bundles {
		if b.Name == name {
			return b, true
		}
	}
	return Bundle{}, false
}

// For returns bundles visible given the selection state.
// Project-scoped skills are only surfaced when a project is selected.
func (r *EmbeddedRegistry) For(hasProject bool) []Bundle {
	var out []Bundle
	for _, b := range r.bundles {
		if isProjectScoped(b.Name) && !hasProject {
			continue
		}
		out = append(out, b)
	}
	return out
}

func isProjectScoped(name string) bool {
	switch name {
	case "workbench-system-prompt", "go-coding-guidelines":
		return true
	default:
		return false
	}
}

func buildSeedBundles() []Bundle {
	return []Bundle{
		{
			Name:        "workbench-orient",
			Description: "Orientation guide: how to use this workbench MCP server",
			Version:     "0.1.0",
			Files: []File{{
				RelPath:  "SKILL.md",
				MIMEType: "text/markdown",
				Content:  func(_ ProjectContext) []byte { return mustReadSeed("workbench-orient/SKILL.md") },
			}},
		},
		{
			Name:        "workbench-system-prompt",
			Description: "Active project context: name, description, and system prompt",
			Version:     "0.1.0",
			Files: []File{{
				RelPath:  "SKILL.md",
				MIMEType: "text/markdown",
				Content:  renderProjectContext,
			}},
		},
		{
			Name:        "go-coding-guidelines",
			Description: "Go coding guidelines for agents working in Go projects",
			Version:     "0.1.0",
			Files: []File{{
				RelPath:  "SKILL.md",
				MIMEType: "text/markdown",
				Content:  func(_ ProjectContext) []byte { return mustReadSeed("go-coding-guidelines/SKILL.md") },
			}},
		},
	}
}

func mustReadSeed(rel string) []byte {
	b, err := seedFS.ReadFile("seeds/" + rel)
	if err != nil {
		panic(fmt.Sprintf("skills: missing embedded seed %q: %v", rel, err))
	}
	return b
}

func renderProjectContext(ctx ProjectContext) []byte {
	var sb strings.Builder

	sb.WriteString("# Project: ")
	sb.WriteString(ctx.Name)
	sb.WriteString("\n\n")

	if strings.TrimSpace(ctx.Description) != "" {
		sb.WriteString(ctx.Description)
		sb.WriteString("\n\n")
	}

	sb.WriteString("## System Prompt\n\n")
	if strings.TrimSpace(ctx.SystemPrompt) != "" {
		sb.WriteString(ctx.SystemPrompt)
	} else {
		sb.WriteString("_No system prompt configured for this project._")
	}
	sb.WriteString("\n\n")

	sb.WriteString("---\n")
	sb.WriteString("_Re-call `refresh()` after context compaction to reload this context._\n")

	return []byte(sb.String())
}
