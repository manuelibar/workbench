---
id: "work-management-assumption"
type: "assumption"
title: "Work Management Starts as Local Deterministic State"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Work Management Starts as Local Deterministic State

## Statement

The first work-management implementation should start with local deterministic
state and MCP-facing contracts rather than an external service, hosted project
tracker, or database-backed sync system.

This assumption does not require every future work item to be a typed artifact.
It means the initial behavior should preserve Workbench's local-first kernel
properties until the data model, namespace dependency, and mutation semantics
are proven.

## Evidence

Current Workbench `main` is intentionally a stdio MCP server with no HTTP
listener and no database requirement. The repository already stores typed
Markdown artifacts under `docs/artifacts/`, and the epic workflow says runtime
ports should wait until the kickoff packet establishes the contract.

External research supports starting with flexible views and metadata rather
than a heavy process model. GitHub Projects exposes table, board, and roadmap
views over the same underlying items; Backlog boards update issue status from
card movement; BacklogMD demonstrates a plain-Markdown backlog protocol for
agent-readable work. These are product signals, not binding implementation
choices.

## Validation Plan

Validate this assumption during the first implementation drill by:

- Comparing a file-backed work item model, artifact-backed work item model, and
  lightweight embedded store against the RFC's requirements.
- Proving that one namespace-scoped fixture can drive backlog, board, WIP, and
  daily-plan views without duplicating state.
- Confirming with namespace-management that the active environment can point to
  a namespace reference without redefining namespace identity.
- Revisiting the storage decision before external integrations, multi-user
  workflows, or high-volume work item history are added.

## Source References

- [RFC: Work Management](work-management-rfc.md)
- [Research Note](work-management-research-note.md)
- `README.md`
- `docs/how-to/epic-branch-workflow.md`

## Open Questions

Open human nudges are centralized in the RFC's Human-in-the-loop Index.
