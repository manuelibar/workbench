package onboarding_test

import (
	"strings"
	"testing"

	"github.com/manuelibar/workbench/internal/onboarding"
)

func TestSkillMarkdown_NonEmpty(t *testing.T) {
	t.Parallel()
	body := onboarding.SkillMarkdown()
	if len(body) == 0 {
		t.Fatal("SkillMarkdown is empty; SKILL.md was not embedded")
	}
	for _, want := range []string{
		"Workbench MCP",
		"refresh",
		"WorkSession",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("SKILL.md missing %q", want)
		}
	}
}
