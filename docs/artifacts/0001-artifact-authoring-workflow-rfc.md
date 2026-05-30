# RFC 0001: Artifact Authoring Workflow

Status: accepted

## Summary

Workbench should author durable project and process assets as typed,
versioned artifacts inside the existing MCP server. The workflow extends the
current `artifact.*` tools, artifact persistence tables, MCP resources,
elicitation support, and refresh-driven tool visibility. It does not introduce
a standalone `artifactly` CLI.

## Rules

Conventional repository documents remain conventional files:

- `README.md`
- `CONTRIBUTING.md`
- `SECURITY.md`
- `.github/*`
- repository-level RFCs such as this document

Project and process assets live as Workbench artifacts:

- RFCs, ADRs, PRDs, requirements, specs, research notes
- risks, assumptions, constraints
- acceptance contracts, test strategies, implementation plans
- runbooks, postmortems, retrospectives, iteration logs
- charters, problem statements, opportunities, decision records

There are no generated `docs/catalog` views. Discovery comes from artifact
metadata, artifact resources such as `workbench:///artifacts/{id}`, and MCP
tools such as `artifact.list`, `artifact.guidance`, and `refresh`.

## Public Interface

`refresh` remains the single selection mutation verb. It accepts:

- `namespace_id`
- `project_id`
- `artifact_id`
- `blueprint_id`
- `mode_name`
- `focus`
- `clear`
- `clear_artifact`
- `clear_focus`

Selecting an artifact resolves and persists its project and namespace, clears
blueprint and mode, refreshes the visible tool surface, and returns
`artifact_id` plus `focus` in the selection result.

`focus` is a persisted steering string for the active work session. The agent
uses it to shape the next authoring step; Workbench stores it and returns it
with selection state.

## Authoring Surface

When a project is selected, `artifact.begin` creates a draft from a registered
contract type and title, writes version 1, selects the artifact, and returns
guidance.

When an artifact is selected:

- `artifact.guidance` returns the contract, missing pieces, focus, and next
  deterministic authoring step.
- `artifact.validate` checks the body convention and required sections.
- `artifact.elicit` asks the human for missing information via MCP
  elicitation and appends a new artifact version only when accepted.

Workbench does not call an LLM internally. The agent performs inference and
drafting; Workbench provides contracts, validation, elicitation, state, and
guidance.

## Body Convention

The existing database model is retained:

- normalized fields live in `content_jsonb`
- a Markdown projection lives in `content_text`

The Markdown projection starts with YAML frontmatter.

Required frontmatter:

- `id`
- `type`
- `title`
- `status`
- `created`
- `updated`

Optional frontmatter:

- `owners`
- `tags`
- `parents`
- `source_refs`
- `supersedes`
- `superseded_by`

The initial registry includes: `rfc`, `adr`, `prd`, `requirement`, `spec`,
`research_note`, `risk`, `assumption`, `constraint`, `acceptance_contract`,
`test_strategy`, `implementation_plan`, `runbook`, `postmortem`,
`retrospective`, `iteration_log`, `charter`, `problem_statement`,
`opportunity`, and `decision_record`.
