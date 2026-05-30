---
id: "idea-mgmt-research-note"
type: "research_note"
title: "Idea Management Research Note"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Idea Management Research Note

## Question

How should Workbench model idea capture, refinement, relation, and promotion
so raw ideas remain separate from committed work while still connecting to
artifacts, work items, opportunities, knowledge, and memory?

## Sources

- `README.md`: establishes current Workbench as a local stdio MCP context and
  artifact kernel with file-backed artifacts.
- `docs/how-to/epic-branch-workflow.md`: establishes current epic packet
  expectations, the RFC drill-hub role, and the current-docs research base.
- `docs/reference/artifact-conventions.md`: defines artifact frontmatter,
  supported contract types, deterministic validation, and the required
  human-in-the-loop index for RFCs.
- MCP Resources specification, version 2025-06-18:
  <https://modelcontextprotocol.io/specification/2025-06-18/server/resources>.
  Relevant because idea records are primarily context that clients may list,
  read, search, subscribe to, or select.
- MCP Tools specification, version 2025-06-18:
  <https://modelcontextprotocol.io/specification/2025-06-18/server/tools>.
  Relevant because capture, refinement, relation, lifecycle transition, and
  promotion are model-callable actions that need schemas and user control.
- MCP Elicitation specification, version 2025-06-18:
  <https://modelcontextprotocol.io/specification/2025-06-18/client/elicitation>.
  Relevant because unresolved idea details may require structured user input
  while respecting consent and avoiding sensitive information requests.
- GitHub "Communicating on GitHub":
  <https://docs.github.com/en/get-started/using-github/communicating-on-github>.
  Relevant because it distinguishes open-form ideas and discussions from
  specific repository issues and pull requests.
- GitHub Discussions quickstart:
  <https://docs.github.com/en/discussions/quickstart>. Relevant because it
  frames discussions as transparent conversation that need not be tracked as
  project work, and it describes labels, categories, and eventual movement to
  issues when scope is clear.
- GitHub "Creating an issue":
  <https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/creating-an-issue>.
  Relevant because it documents creating an issue from a discussion without
  deleting the original discussion, preserving the separation between source
  conversation and tracked work.
- GitHub issue dependencies:
  <https://docs.github.com/en/enterprise-cloud@latest/issues/tracking-your-work-with-issues/using-issues/creating-issue-dependencies>.
  Relevant because promoted work may need explicit blocking relationships, but
  raw idea relations should not automatically become work dependencies.

## Findings

The repository's current foundation supports a small, deterministic contract
surface: context, typed Markdown artifacts, artifact resources, and explicit
artifact tools. That suggests ideas should start as a separate context layer
that can later promote into existing typed artifacts or future target systems.

MCP resources are a good fit for idea discovery because they are read-oriented
context and support list/read/template patterns. MCP tools are a fit for
mutating operations such as capture, refinement, relation changes, lifecycle
transitions, and promotion, but the tools specification emphasizes schemas,
validation, and human-visible control for operations. MCP elicitation provides
a structured path for asking a user for missing details, which matches the
RFC's human nudge pattern.

GitHub's current collaboration model separates open-form discussion from
specific tracked work. Discussions support brainstorming and broad
conversation, while issues are better for specific tasks, enhancements, bugs,
or planned work. GitHub can create an issue from a discussion while retaining
the source discussion, and its issue hierarchy and dependency features show
that work relationships are distinct from discussion or idea relationships.

Together, these sources support a Workbench idea model with cheap capture,
explicit lifecycle, typed relations, and promotion that creates or links a
separate committed target instead of converting the raw idea in place.

## Implications

The first implementation should expose ideas as context resources before
optimizing around complex planning. Promotion should be explicit, audited by
backlinks, and initially narrow enough to create typed artifacts before trying
to write into future work, opportunity, knowledge, or memory systems. The RFC
should remain the living drill hub, with implementation plans and requirements
linked from it so action work stays separate from raw idea capture.

## Open Questions

The research leaves the RFC-indexed nudges `IM-HIL-001`, `IM-HIL-002`, and
`IM-HIL-003` for promotion scope, lifecycle status approval, and raw-capture
edit policy.
