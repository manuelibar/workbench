---
id: "session-management-rfc"
type: "rfc"
title: "Session Management RFC"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Session Management RFC

## Summary

Create a durable Workbench session layer that records a local work thread with
a stable ID, lifecycle state, context checkpoints, and bounded event history.
Keep this layer distinct from volatile live context and from MCP transport
sessions.

This RFC is the living drill hub for the session management epic. Action
artifacts created from it:

- [Initial implementation plan](session-management-initial-implementation-plan.md)
- [Test strategy](session-management-test-strategy.md)

Supporting packet artifacts:

- [Charter](session-management-charter.md)
- [Problem statement](session-management-problem-statement.md)
- [Concept map](session-management-concept-map.md)
- [Assumption](session-management-assumption.md)
- [Risk](session-management-risk.md)
- [Research note](session-management-research-note.md)

## Problem

Workbench has live context and durable artifacts, but no durable record for the
session that connects them. A model context window can disappear, a process can
restart, or a user can return later with only artifacts and chat history as
evidence.

At the same time, simply persisting live context would be the wrong boundary.
`focus` and `artifact_id` are volatile steering fields. They need checkpoints
and provenance, not silent conversion into durable truth.

The term session is also overloaded. MCP transport sessions describe protocol
connections. Workbench sessions should describe durable local work threads.

## Proposal

Add a Workbench durable session concept in a later runtime pass:

- Generate a stable `session_id` for each Workbench work thread.
- Track lifecycle state as `active`, `suspended`, or `closed`.
- Maintain a latest context checkpoint containing restorable `focus` and
  `artifact_id` values plus generation metadata.
- Append normalized events for lifecycle, context, artifact, capability sync,
  and short note checkpoints.
- Expose a small MCP surface for begin, current, list, get, resume, close, and
  event pagination.
- Add session resources only when progressive disclosure rules make them useful.

Resume behavior:

1. `session.resume` reads the durable session and verifies it is resumable.
2. It checks whether the checkpoint references an existing artifact.
3. It applies or previews the checkpoint through the existing `context` mutation
   path.
4. It records a `session.resumed` event and any context events created by the
   patch.

Close behavior:

1. `session.close` marks the durable record `closed` using an idempotent update.
2. It records who or what initiated the close and whether live context was
   cleared.
3. It keeps read-only history available until retention policy deletes or
   compacts it.

Storage behavior:

Start with a store interface. The default implementation can be file-backed:
one session metadata file plus one append-only event log per session in a
configurable session directory outside `docs/artifacts/`. SQLite remains a
future option if pagination, concurrent writers, or retention compaction become
too complex for plain files.

## Tradeoffs

File-backed storage matches Workbench's current local-first posture and keeps
the first implementation easy to inspect. It makes indexed queries and
concurrent access harder than a database would.

Normalized events protect privacy and keep history concise. They make forensic
debugging less complete than raw MCP message capture.

Explicit resume avoids surprising context changes. It adds one more tool call
before an agent can continue old work.

Closed sessions remaining readable helps accountability and recovery. It means
retention policy must be explicit so closed sessions do not accumulate forever.

## Rollout

1. Finalize the human-in-the-loop decisions in this RFC.
2. Implement the session store and domain model behind internal interfaces.
3. Add focused unit tests for IDs, lifecycle transitions, event append, resume
   checkpoints, close idempotency, and retention.
4. Add MCP tool handlers behind the smallest useful visible surface.
5. Add integration tests proving that resume goes through `context` and emits
   the expected capability sync behavior.
6. Update foundation docs after the runtime behavior lands.

No runtime behavior changes are part of this bootstrap commit.

## Open Questions

### Human-in-the-loop Index

| ID | Nudge | Type | Why it matters | Blocks | Default if unanswered |
|---|---|---|---|---|---|
| SM-HIL-001 | Decide whether closed sessions can ever be reopened or whether resume must create a successor session. | decision | This defines close semantics and whether history has a branching model. | `session.close`, `session.resume`, lifecycle tests. | Closed sessions are read-only and non-resumable; resuming closed work creates a new session linked to the closed one. |
| SM-HIL-002 | Choose the default retention window for session events and context checkpoints. | decision | Retention controls privacy, disk growth, and how far back resume inspection can go. | Store schema, compaction tests, user-visible history docs. | Keep full normalized events for 30 days, retain final session metadata and closing summary until explicit cleanup. |
| SM-HIL-003 | Confirm whether raw MCP messages are excluded from session history by default. | approval | This is the main privacy and security boundary for event history. | Event schema, redaction tests, risk mitigation. | Exclude raw MCP messages and full tool outputs; store normalized event summaries only. |
| SM-HIL-004 | Approve file-backed storage as the first implementation backend. | approval | The backend shapes atomicity, pagination, corruption recovery, and future migration work. | Store interface, fixtures, implementation sequencing. | Use metadata JSON plus append-only JSONL events in a configurable local session directory. |
| SM-HIL-005 | Decide whether `session.resume` applies context immediately or defaults to preview mode. | tradeoff | Immediate apply is ergonomic; preview is safer when stale checkpoints reference outdated artifacts. | Resume tool contract, capability sync behavior, integration tests. | Apply immediately when checkpoint artifacts exist; otherwise return a preview with a recovery warning. |

## Source References

- `README.md`
- `docs/how-to/epic-branch-workflow.md`
- `docs/reference/artifact-conventions.md`
- `docs/reference/context-contract.md`
- [MCP lifecycle specification, version 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25/basic/lifecycle)
- [MCP transport specification, version 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25/basic/transports)
- [MCP tasks specification, version 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25/basic/utilities/tasks)
- [MCP security best practices](https://modelcontextprotocol.io/docs/tutorials/security/security_best_practices)
- [TypeScript SDK server guide](https://github.com/modelcontextprotocol/typescript-sdk/blob/main/docs/server.md)
- [Go SDK package documentation](https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp)
