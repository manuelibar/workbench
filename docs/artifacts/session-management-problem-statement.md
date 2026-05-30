---
id: "session-management-problem-statement"
type: "problem_statement"
title: "Session Management Problem Statement"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Session Management Problem Statement

## Context

Workbench currently keeps live context in memory with two fields: `focus` and
`artifact_id`. It persists typed Markdown artifacts under `docs/artifacts/`,
but it does not persist the identity, lifecycle, or event history of the agent
work session that produced those artifact updates.

MCP also uses the word session for client-server protocol lifecycles and, for
Streamable HTTP, optional `MCP-Session-Id` transport state. Workbench currently
runs as a local stdio server, so this epic needs a Workbench-level durable
session record rather than a transport-only session mirror.

## Problem

Without durable Workbench sessions, an agent can lose the continuity needed to
resume work, explain why live context changed, distinguish an intentionally
closed thread from an interrupted one, or inspect a concise history of events
that led to the current artifact state.

If this epic stores too much, session history becomes a hidden transcript or
secret store. If it stores too little, resume semantics become unreliable and
future epics will invent incompatible session identifiers.

## Impact

Agents must reconstruct intent from chat history and scattered artifact edits.
Users cannot reliably ask what session they are in, whether it is safe to close
it, or what durable record will exist after the model context window is gone.
Future Workbench branches that need continuity, such as memory, knowledge,
work management, AFK, and skills, lack a common record to reference.

The lack of close semantics also makes cleanup ambiguous: deleting live context,
terminating the process, and declaring the durable work thread finished are
currently different events with no shared contract.

## Constraints

The first implementation should respect the current foundation: local stdio
MCP, no database requirement, deterministic context changes, file-backed
artifacts, and capability sync through `context`.

The design must not treat the current `context` document as durable truth.
Instead, it should record checkpoints and events that can be used to restore or
explain context through explicit resume behavior.

Session identifiers must be safe to store and pass through tools. Event history
must be bounded, inspectable, and redactable by default.

## Source References

- `README.md`
- `docs/reference/context-contract.md`
- `docs/explanation/context-window-thesis.md`
- [MCP lifecycle specification, version 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25/basic/lifecycle)
- [MCP transport specification, version 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25/basic/transports)
- [Research note](session-management-research-note.md)

## Open Questions

The unresolved decisions are indexed as `SM-HIL-001` through `SM-HIL-005` in
the RFC Human-in-the-loop Index.
