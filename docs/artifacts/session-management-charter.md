---
id: "session-management-charter"
type: "charter"
title: "Session Management Charter"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Session Management Charter

## Mission

Define Workbench session management as a durable coordination layer that can
name, resume, close, and inspect a local agent work session without confusing
that record with the current in-memory `context` document or with an MCP
transport session.

## Scope

In scope: Workbench session identity, session lifecycle states, resume and
close semantics, event history, context checkpoints, storage retention
tradeoffs, and the rules that connect volatile `focus` and `artifact_id` values
to durable session records.

Out of scope: multi-user authorization, remote HTTP hosting, cloud sync,
knowledge base design, backlog ownership, memory summarization policy, and
Artifact B2 workflows beyond recording references to artifacts touched during a
session.

## Stakeholders

The branch owner is accountable for the epic packet and later implementation
contract. Local agent workflows are the primary user. Future epic branches for
memory, knowledge, work management, roles, namespaces, AFK, and skills are
downstream stakeholders because they may want to attach their own records to a
stable Workbench session identifier.

## Success Criteria

The epic succeeds when Workbench can create a durable session record, append a
bounded event history, resume a prior session through the same `context` sync
path used for live context changes, close a session with explicit terminal
semantics, and explain what was intentionally not persisted.

The implementation should preserve the `main` foundation: Workbench remains a
local stdio MCP server with a small default surface, file-backed artifacts, and
progressive disclosure of tools.

## Source References

- `README.md`
- `docs/how-to/epic-branch-workflow.md`
- `docs/reference/context-contract.md`
- `docs/reference/artifact-conventions.md`
- `docs/explanation/context-window-thesis.md`
- `docs/explanation/progressive-disclosure.md`
- [RFC: Session Management](session-management-rfc.md)

## Open Questions

Human nudges for this charter are tracked in the RFC Human-in-the-loop Index.
