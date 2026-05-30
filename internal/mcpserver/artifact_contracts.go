package mcpserver

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/manuelibar/workbench/internal/domain"
)

const artifactTimeFormat = time.RFC3339

var requiredArtifactFrontmatter = []string{"id", "type", "title", "status", "created", "updated"}

// ArtifactSectionSpec describes one section in a typed artifact contract.
type ArtifactSectionSpec struct {
	Key      string `json:"key"`
	Title    string `json:"title"`
	Prompt   string `json:"prompt,omitempty"`
	Required bool   `json:"required"`
}

// ArtifactContract is the deterministic registry entry for one artifact type.
type ArtifactContract struct {
	Type             string                `json:"type"`
	Title            string                `json:"title"`
	Purpose          string                `json:"purpose"`
	RequiredSections []ArtifactSectionSpec `json:"required_sections"`
	OptionalSections []ArtifactSectionSpec `json:"optional_sections,omitempty"`
}

type artifactSectionContent struct {
	Key      string `json:"key"`
	Title    string `json:"title"`
	Body     string `json:"body,omitempty"`
	Required bool   `json:"required"`
}

var artifactContracts = []ArtifactContract{
	contract("rfc", "RFC", "Propose and evaluate a significant change before implementation.",
		req("summary", "Summary", "Summarize the proposal and intended outcome."),
		req("problem", "Problem", "Describe the problem, context, and why it matters."),
		req("proposal", "Proposal", "Describe the proposed approach and key mechanics."),
		req("tradeoffs", "Tradeoffs", "List alternatives, risks, and explicit tradeoffs."),
		req("rollout", "Rollout", "Describe rollout, migration, compatibility, and operational impact."),
		req("open_questions", "Open Questions", "List unresolved questions and decisions needed.")),
	contract("adr", "ADR", "Record a durable architecture decision and its consequences.",
		req("context", "Context", "Describe the forces and constraints that led to the decision."),
		req("decision", "Decision", "State the decision plainly."),
		req("consequences", "Consequences", "Describe positive, negative, and neutral consequences."),
		req("alternatives", "Alternatives", "List meaningful alternatives and why they were not chosen.")),
	contract("prd", "PRD", "Define product intent, user value, and acceptance boundaries.",
		req("problem", "Problem", "Describe the customer or business problem."),
		req("goals", "Goals", "List measurable product goals."),
		req("non_goals", "Non-goals", "State what is intentionally out of scope."),
		req("users", "Users", "Identify primary users and stakeholders."),
		req("requirements", "Requirements", "Capture functional requirements."),
		req("success_metrics", "Success Metrics", "Define how success will be measured.")),
	contract("requirement", "Requirement", "Capture one specific required capability or property.",
		req("statement", "Statement", "State the requirement unambiguously."),
		req("rationale", "Rationale", "Explain why the requirement exists."),
		req("acceptance_criteria", "Acceptance Criteria", "Define observable pass/fail criteria.")),
	contract("spec", "Spec", "Define a concrete technical design and verification path.",
		req("context", "Context", "Summarize the current system and constraints."),
		req("design", "Design", "Describe the proposed implementation design."),
		req("interfaces", "Interfaces", "Describe APIs, schemas, files, or contracts affected."),
		req("edge_cases", "Edge Cases", "List expected edge cases and failure behavior."),
		req("test_plan", "Test Plan", "Describe how the design will be verified.")),
	contract("research_note", "Research Note", "Capture sourced findings and their implications.",
		req("question", "Question", "State the research question."),
		req("sources", "Sources", "List source references and their relevance."),
		req("findings", "Findings", "Summarize the evidence."),
		req("implications", "Implications", "Explain what the findings change or enable.")),
	contract("risk", "Risk", "Track a risk with mitigation and ownership.",
		req("description", "Description", "Describe the risk event or condition."),
		req("impact", "Impact", "Describe expected impact if the risk materializes."),
		req("likelihood", "Likelihood", "Estimate likelihood and confidence."),
		req("mitigation", "Mitigation", "Describe mitigation and contingency actions."),
		req("owner", "Owner", "Name the accountable owner or role.")),
	contract("assumption", "Assumption", "Make an assumption explicit and testable.",
		req("statement", "Statement", "State the assumption."),
		req("evidence", "Evidence", "Describe available supporting or opposing evidence."),
		req("validation_plan", "Validation Plan", "Describe how the assumption will be validated.")),
	contract("constraint", "Constraint", "Record a design or process constraint.",
		req("statement", "Statement", "State the constraint."),
		req("source", "Source", "Identify where the constraint comes from."),
		req("impact", "Impact", "Explain how the constraint affects choices.")),
	contract("acceptance_contract", "Acceptance Contract", "Define agreed acceptance boundaries for a deliverable.",
		req("scope", "Scope", "Define what the contract covers."),
		req("criteria", "Criteria", "List acceptance criteria."),
		req("verification", "Verification", "Describe how criteria will be verified.")),
	contract("test_strategy", "Test Strategy", "Plan verification across levels and risks.",
		req("scope", "Scope", "Define what is covered by the strategy."),
		req("test_levels", "Test Levels", "Describe unit, integration, system, or manual coverage."),
		req("fixtures", "Fixtures", "List required fixtures, data, or environments."),
		req("risks", "Risks", "Identify test risks and blind spots.")),
	contract("implementation_plan", "Implementation Plan", "Sequence implementation work with verification and rollback.",
		req("objective", "Objective", "State the implementation objective."),
		req("steps", "Steps", "List ordered implementation steps."),
		req("verification", "Verification", "Describe checks before completion."),
		req("rollback", "Rollback", "Describe how to revert or recover.")),
	contract("runbook", "Runbook", "Document an operational procedure.",
		req("scope", "Scope", "Define when the runbook applies."),
		req("prerequisites", "Prerequisites", "List required access, state, and tools."),
		req("procedure", "Procedure", "List execution steps."),
		req("verification", "Verification", "Describe how to confirm success."),
		req("escalation", "Escalation", "Describe escalation triggers and contacts.")),
	contract("postmortem", "Postmortem", "Analyze an incident and corrective actions.",
		req("incident", "Incident", "Summarize the incident."),
		req("impact", "Impact", "Describe user, business, or system impact."),
		req("timeline", "Timeline", "Record the important sequence of events."),
		req("root_cause", "Root Cause", "Identify root causes and contributing factors."),
		req("corrective_actions", "Corrective Actions", "List owner-backed corrective actions.")),
	contract("retrospective", "Retrospective", "Reflect on an iteration and produce actions.",
		req("context", "Context", "Describe the period, project, or event reviewed."),
		req("what_worked", "What Worked", "List effective practices or outcomes."),
		req("what_did_not", "What Did Not", "List problems and friction."),
		req("actions", "Actions", "List follow-up actions.")),
	contract("iteration_log", "Iteration Log", "Record what changed in an iteration and what comes next.",
		req("goal", "Goal", "State the iteration goal."),
		req("changes", "Changes", "Summarize changes made."),
		req("results", "Results", "Capture outcomes and evidence."),
		req("next", "Next", "List next steps.")),
	contract("charter", "Charter", "Define mission, scope, stakeholders, and success.",
		req("mission", "Mission", "State the mission."),
		req("scope", "Scope", "Define in-scope and out-of-scope boundaries."),
		req("stakeholders", "Stakeholders", "Identify stakeholders and responsibilities."),
		req("success_criteria", "Success Criteria", "Define success criteria.")),
	contract("problem_statement", "Problem Statement", "Frame a problem before solutioning.",
		req("context", "Context", "Describe the background and current situation."),
		req("problem", "Problem", "State the problem precisely."),
		req("impact", "Impact", "Explain who or what is affected."),
		req("constraints", "Constraints", "List meaningful constraints.")),
	contract("opportunity", "Opportunity", "Frame an opportunity and its value.",
		req("context", "Context", "Describe relevant context."),
		req("opportunity", "Opportunity", "State the opportunity."),
		req("value", "Value", "Describe expected value."),
		req("risks", "Risks", "List risks and uncertainties.")),
	contract("decision_record", "Decision Record", "Record a decision outside the heavier ADR format.",
		req("decision", "Decision", "State the decision."),
		req("rationale", "Rationale", "Explain why it was chosen."),
		req("alternatives", "Alternatives", "List alternatives considered."),
		req("follow_up", "Follow-up", "List follow-up work or review dates.")),
}

func contract(typ, title, purpose string, required ...ArtifactSectionSpec) ArtifactContract {
	requiredKeys := map[string]bool{}
	for _, spec := range required {
		requiredKeys[spec.Key] = true
	}
	optional := []ArtifactSectionSpec{
		opt("source_refs", "Source References", "List notes, issues, docs, commits, or external sources this artifact derives from."),
		opt("open_questions", "Open Questions", "Track unresolved questions that do not block the current draft."),
	}
	filteredOptional := optional[:0]
	for _, spec := range optional {
		if !requiredKeys[spec.Key] {
			filteredOptional = append(filteredOptional, spec)
		}
	}
	return ArtifactContract{
		Type:             typ,
		Title:            title,
		Purpose:          purpose,
		RequiredSections: required,
		OptionalSections: filteredOptional,
	}
}

func req(key, title, prompt string) ArtifactSectionSpec {
	return ArtifactSectionSpec{Key: key, Title: title, Prompt: prompt, Required: true}
}

func opt(key, title, prompt string) ArtifactSectionSpec {
	return ArtifactSectionSpec{Key: key, Title: title, Prompt: prompt}
}

func normalizeArtifactType(typ string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(typ), " ", "_"))
}

func artifactContractFor(typ string) (ArtifactContract, bool) {
	typ = normalizeArtifactType(typ)
	for _, c := range artifactContracts {
		if c.Type == typ {
			return c, true
		}
	}
	return ArtifactContract{}, false
}

func artifactContractTypes() []string {
	out := make([]string, len(artifactContracts))
	for i, c := range artifactContracts {
		out[i] = c.Type
	}
	return out
}

func newArtifactContent(artifactID uuid.UUID, typ, title, status, focus string, now time.Time, c ArtifactContract) map[string]any {
	sections := make([]artifactSectionContent, 0, len(c.RequiredSections)+len(c.OptionalSections))
	for _, spec := range c.RequiredSections {
		sections = append(sections, artifactSectionContent{Key: spec.Key, Title: spec.Title, Required: true})
	}
	for _, spec := range c.OptionalSections {
		sections = append(sections, artifactSectionContent{Key: spec.Key, Title: spec.Title})
	}
	return map[string]any{
		"id":       artifactID.String(),
		"type":     normalizeArtifactType(typ),
		"title":    strings.TrimSpace(title),
		"status":   status,
		"created":  now.UTC().Format(artifactTimeFormat),
		"updated":  now.UTC().Format(artifactTimeFormat),
		"focus":    strings.TrimSpace(focus),
		"sections": sections,
	}
}

func artifactMarkdownProjection(content map[string]any, c ArtifactContract) string {
	var b strings.Builder
	b.WriteString("---\n")
	for _, key := range requiredArtifactFrontmatter {
		b.WriteString(key)
		b.WriteString(": ")
		b.WriteString(yamlString(contentString(content, key)))
		b.WriteByte('\n')
	}
	for _, key := range []string{"owners", "tags", "parents", "source_refs", "supersedes", "superseded_by"} {
		values := stringSliceFromAny(content[key])
		if len(values) == 0 {
			continue
		}
		b.WriteString(key)
		b.WriteString(": ")
		b.WriteString(yamlStringList(values))
		b.WriteByte('\n')
	}
	b.WriteString("---\n\n")

	title := contentString(content, "title")
	if title == "" {
		title = c.Title
	}
	b.WriteString("# ")
	b.WriteString(title)
	b.WriteString("\n\n")

	if focus := contentString(content, "focus"); focus != "" {
		b.WriteString("Focus: ")
		b.WriteString(focus)
		b.WriteString("\n\n")
	}

	sections := artifactSectionsFromContent(content, c)
	for _, section := range sections {
		b.WriteString("## ")
		b.WriteString(section.Title)
		b.WriteString("\n\n")
		if strings.TrimSpace(section.Body) != "" {
			b.WriteString(strings.TrimSpace(section.Body))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func artifactSectionsFromContent(content map[string]any, c ArtifactContract) []artifactSectionContent {
	raw := content["sections"]
	sections := decodeArtifactSections(raw)
	if len(sections) == 0 {
		for _, spec := range c.RequiredSections {
			sections = append(sections, artifactSectionContent{Key: spec.Key, Title: spec.Title, Required: true})
		}
		for _, spec := range c.OptionalSections {
			sections = append(sections, artifactSectionContent{Key: spec.Key, Title: spec.Title})
		}
		return sections
	}

	byKey := map[string]int{}
	for i := range sections {
		sections[i].Key = strings.TrimSpace(sections[i].Key)
		sections[i].Title = strings.TrimSpace(sections[i].Title)
		if sections[i].Key != "" {
			byKey[sections[i].Key] = i
		}
	}
	for _, spec := range c.RequiredSections {
		if idx, ok := byKey[spec.Key]; ok {
			sections[idx].Title = spec.Title
			sections[idx].Required = true
			continue
		}
		sections = append(sections, artifactSectionContent{Key: spec.Key, Title: spec.Title, Required: true})
	}
	for _, spec := range c.OptionalSections {
		if _, ok := byKey[spec.Key]; ok {
			continue
		}
		sections = append(sections, artifactSectionContent{Key: spec.Key, Title: spec.Title})
	}
	return sections
}

func decodeArtifactSections(raw any) []artifactSectionContent {
	switch v := raw.(type) {
	case []artifactSectionContent:
		return append([]artifactSectionContent(nil), v...)
	case []map[string]any:
		out := make([]artifactSectionContent, 0, len(v))
		for _, item := range v {
			out = append(out, artifactSectionContent{
				Key:      contentString(item, "key"),
				Title:    contentString(item, "title"),
				Body:     contentString(item, "body"),
				Required: contentBool(item, "required"),
			})
		}
		return out
	case []any:
		out := make([]artifactSectionContent, 0, len(v))
		for _, item := range v {
			if m, ok := item.(map[string]any); ok {
				out = append(out, artifactSectionContent{
					Key:      contentString(m, "key"),
					Title:    contentString(m, "title"),
					Body:     contentString(m, "body"),
					Required: contentBool(m, "required"),
				})
			}
		}
		return out
	default:
		return nil
	}
}

func artifactContentWithSection(a domain.Artifact, content map[string]any, c ArtifactContract, sectionKey, body, focus string, now time.Time) map[string]any {
	out := cloneContent(content)
	if out["id"] == nil {
		out["id"] = a.ID.String()
	}
	if out["type"] == nil {
		out["type"] = a.Type
	}
	if out["status"] == nil {
		out["status"] = a.Status
	}
	if out["title"] == nil {
		out["title"] = fmt.Sprintf("%s %s", c.Title, a.ID.String()[:8])
	}
	if out["created"] == nil {
		out["created"] = now.UTC().Format(artifactTimeFormat)
	}
	out["updated"] = now.UTC().Format(artifactTimeFormat)
	if strings.TrimSpace(focus) != "" {
		out["focus"] = strings.TrimSpace(focus)
	}
	sections := artifactSectionsFromContent(out, c)
	found := false
	for i := range sections {
		if sections[i].Key == sectionKey {
			sections[i].Body = strings.TrimSpace(body)
			found = true
			break
		}
	}
	if !found {
		sections = append(sections, artifactSectionContent{Key: sectionKey, Title: sectionKey, Body: strings.TrimSpace(body), Required: true})
	}
	out["sections"] = sections
	return out
}

func cloneContent(content map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range content {
		out[k] = v
	}
	return out
}

func contentString(content map[string]any, key string) string {
	switch v := content[key].(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	case nil:
		return ""
	default:
		return fmt.Sprint(v)
	}
}

func contentBool(content map[string]any, key string) bool {
	v, _ := content[key].(bool)
	return v
}

func stringSliceFromAny(raw any) []string {
	switch v := raw.(type) {
	case []string:
		return append([]string(nil), v...)
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

func yamlString(v string) string {
	return strconv.Quote(v)
}

func yamlStringList(values []string) string {
	raw, err := json.Marshal(values)
	if err != nil {
		return "[]"
	}
	return string(raw)
}

func frontmatterKeys(text string) map[string]bool {
	keys := map[string]bool{}
	text = strings.ReplaceAll(text, "\r\n", "\n")
	if !strings.HasPrefix(text, "---\n") {
		return keys
	}
	lines := strings.Split(text, "\n")
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "---" {
			break
		}
		key, _, ok := strings.Cut(line, ":")
		if ok {
			keys[strings.TrimSpace(key)] = true
		}
	}
	return keys
}

func missingFrontmatterKeys(text string) []string {
	keys := frontmatterKeys(text)
	missing := make([]string, 0, len(requiredArtifactFrontmatter))
	for _, key := range requiredArtifactFrontmatter {
		if !keys[key] {
			missing = append(missing, key)
		}
	}
	return missing
}

func sectionBodyMissing(body string) bool {
	body = strings.TrimSpace(body)
	switch strings.ToLower(body) {
	case "", "todo", "tbd", "n/a":
		return true
	default:
		return false
	}
}
