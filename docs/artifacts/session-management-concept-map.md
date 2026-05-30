---
id: "session-management-concept-map"
type: "spec"
title: "Session Management Concept Map"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Session Management Concept Map

## Context

This taxonomy separates three similarly named layers:

- MCP transport session: protocol connection state negotiated during MCP
  initialization, with transport-specific close behavior.
- Workbench live context: the current in-memory `focus` and `artifact_id`
  fields exposed through the `context` tool and `workbench:///context`.
- Workbench durable session record: the future persisted record owned by this
  epic, used to resume, close, and inspect a local work thread.

Only the durable Workbench session record is in scope for this epic. It may
observe live context and MCP lifecycle events, but it should not impersonate
either layer.

## Design

Core terms:

| Concept | Meaning | Persistence |
|---|---|---|
| `session_id` | Stable Workbench identifier for one durable work thread. | Persisted. |
| `session_state` | Lifecycle state such as `active`, `suspended`, or `closed`. | Persisted. |
| `context_checkpoint` | Latest restorable `focus` and `artifact_id` values plus generation metadata. | Persisted as a snapshot, not as live context. |
| `session_event` | Append-only normalized event describing an important session change. | Persisted with retention policy. |
| `resume` | Explicit operation that selects a durable session and proposes or applies its last context checkpoint through the normal `context` path. | Recorded as an event. |
| `close` | Explicit operation that marks a session terminal or non-resumable according to the chosen close policy. | Recorded as an event. |
| `artifact_touch` | Event reference to an artifact read, selected, created, validated, or updated during a session. | Persisted by reference, not by copying artifact content. |

Proposed lifecycle:

| State | Enters when | Allows | Exits when |
|---|---|---|---|
| `active` | A session is created or resumed. | Context checkpoints and events. | Suspended or closed. |
| `suspended` | The process exits, the user switches sessions, or an explicit pause occurs. | Read-only inspection and resume. | Resumed or closed. |
| `closed` | The user closes the work thread or retention policy closes stale work. | Read-only inspection. | Reopen policy is an explicit follow-up decision. |

Event classes:

| Event class | Examples | Payload posture |
|---|---|---|
| `lifecycle` | `created`, `resumed`, `suspended`, `closed` | Full structured metadata. |
| `context` | `focus_changed`, `artifact_selected`, `context_cleared` | New values plus prior hash or generation. |
| `artifact` | `artifact_created`, `artifact_updated`, `artifact_validated` | Artifact ID, type, section keys, validation summary. |
| `capability` | `capability_sync_started`, `capability_sync_timeout` | Category and generation data. |
| `note` | Human or agent summary checkpoints. | Short text, redacted by caller. |

## Interfaces

Likely future MCP tools:

| Tool | Purpose |
|---|---|
| `session.begin` | Create a durable Workbench session record and make it active. |
| `session.current` | Return the active session summary and last context checkpoint. |
| `session.list` | List recent sessions with state, timestamps, labels, and artifact references. |
| `session.get` | Read one session record and bounded event history. |
| `session.resume` | Select a durable session and restore or preview its checkpoint through `context`. |
| `session.close` | Mark a session closed and define whether current live context is cleared. |
| `session.events` | Page through normalized events with opaque cursors. |

Likely future resources:

| Resource | Purpose |
|---|---|
| `workbench:///session` | Current durable session summary, if one is active. |
| `workbench:///sessions/{id}` | Full durable session record and recent event window. |

Storage shape should be hidden behind a store interface. A conservative first
cut can use one metadata file per session plus append-only JSONL event files in
a configurable session directory outside `docs/artifacts/`.

## Edge Cases

Resume must handle a missing artifact referenced by the last checkpoint,
stale capability sync state, closed sessions, corrupted event files, duplicate
resume requests, and process crashes between context mutation and event append.

Close must distinguish clearing live context from closing the durable record.
It must also be idempotent so repeated close calls do not create contradictory
session state.

Event history must survive ordinary process restarts but remain bounded.
History pagination should use opaque cursors rather than exposing file offsets
as public contract.

## Test Plan

Unit tests should cover lifecycle transitions, ID validation, event append and
read ordering, bounded retention, resume from checkpoint, close idempotency,
and context checkpoint serialization.

Integration tests should exercise MCP tool behavior around `session.begin`,
`session.resume`, `session.close`, and capability sync after resume. Crash and
corruption tests should verify that Workbench can report partial history
without silently applying stale context.

## Source References

- [RFC: Session Management](session-management-rfc.md)
- [Implementation plan](session-management-initial-implementation-plan.md)
- [Test strategy](session-management-test-strategy.md)
- `docs/reference/context-contract.md`
- [MCP transport specification, version 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25/basic/transports)
- [MCP tasks specification, version 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25/basic/utilities/tasks)

## Open Questions

The taxonomy defaults are provisional until the RFC resolves `SM-HIL-001`
through `SM-HIL-005`.
