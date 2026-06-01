package artifacts

import (
	"sort"
	"strings"
)

// SectionSpec describes one section required or suggested by a contract.
type SectionSpec struct {
	Key      string
	Title    string
	Prompt   string
	Required bool
}

// Contract is one deterministic typed artifact contract.
type Contract struct {
	Type             string
	Title            string
	Purpose          string
	RequiredSections []SectionSpec
	OptionalSections []SectionSpec
}

type Registry struct {
	byType map[string]Contract
}

func NewRegistry() Registry {
	contracts := []Contract{
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
	reg := Registry{byType: map[string]Contract{}}
	for _, c := range contracts {
		reg.byType[c.Type] = c
	}
	return reg
}

func (r Registry) Get(typ string) (Contract, bool) {
	c, ok := r.byType[normalizeArtifactType(typ)]
	return c, ok
}

func (r Registry) Types() []string {
	types := make([]string, 0, len(r.byType))
	for typ := range r.byType {
		types = append(types, typ)
	}
	sort.Strings(types)
	return types
}

func contract(typ, title, purpose string, required ...SectionSpec) Contract {
	requiredKeys := map[string]bool{}
	for _, spec := range required {
		requiredKeys[spec.Key] = true
	}
	optional := []SectionSpec{
		opt("source_refs", "Source References", "List notes, issues, docs, commits, or external sources this artifact derives from."),
		opt("open_questions", "Open Questions", "Track unresolved questions that do not block the current draft."),
	}
	filtered := optional[:0]
	for _, spec := range optional {
		if !requiredKeys[spec.Key] {
			filtered = append(filtered, spec)
		}
	}
	return Contract{
		Type:             typ,
		Title:            title,
		Purpose:          purpose,
		RequiredSections: required,
		OptionalSections: filtered,
	}
}

func req(key, title, prompt string) SectionSpec {
	return SectionSpec{Key: key, Title: title, Prompt: prompt, Required: true}
}

func opt(key, title, prompt string) SectionSpec {
	return SectionSpec{Key: key, Title: title, Prompt: prompt}
}

func normalizeArtifactType(typ string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(typ), " ", "_"))
}

func normalizeSectionKey(key string) string {
	key = strings.ToLower(strings.TrimSpace(key))
	key = strings.ReplaceAll(key, "-", "_")
	key = strings.ReplaceAll(key, " ", "_")
	var b strings.Builder
	lastUnderscore := false
	for _, r := range key {
		ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if ok {
			b.WriteRune(r)
			lastUnderscore = false
			continue
		}
		if r == '_' && !lastUnderscore {
			b.WriteByte('_')
			lastUnderscore = true
		}
	}
	return strings.Trim(b.String(), "_")
}

func titleFromSectionKey(key string) string {
	parts := strings.Split(normalizeSectionKey(key), "_")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}
