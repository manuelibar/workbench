// Package onboarding embeds the agent-facing onboarding document
// (SKILL.md) so the workbench MCP can serve it as the
// `workbench://skill` resource.
package onboarding

import _ "embed"

//go:embed SKILL.md
var skillMarkdown string

// SkillMarkdown returns the embedded SKILL.md content. Callers must not
// modify the returned string.
func SkillMarkdown() string { return skillMarkdown }
