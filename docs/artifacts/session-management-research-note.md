---
id: "session-management-research-note"
type: "research_note"
title: "Session Management Research Note"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Session Management Research Note

## Question

What should Workbench borrow from current MCP lifecycle, transport, task, and
SDK session patterns while designing a local durable session record that is
separate from volatile context?

## Sources

- `README.md`: establishes Workbench as a local stdio MCP context and artifact
  kernel with no database requirement.
- `docs/how-to/epic-branch-workflow.md`: requires self-contained epic packets,
  current docs, targeted research, and RFC human-in-the-loop indexing.
- `docs/reference/artifact-conventions.md`: defines typed artifact metadata and
  required validation behavior.
- `docs/reference/context-contract.md`: defines the current `context` tool and
  capability sync behavior.
- [MCP lifecycle specification, version 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25/basic/lifecycle):
  authoritative lifecycle phases, initialization, operation, shutdown, and
  timeout guidance.
- [MCP transport specification, version 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25/basic/transports):
  authoritative Streamable HTTP session ID, resumability, redelivery, and
  explicit HTTP DELETE close behavior.
- [MCP tasks specification, version 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25/basic/utilities/tasks):
  authoritative experimental model for durable task IDs, status transitions,
  TTL, polling, result retrieval, and cancellation.
- [MCP security best practices](https://modelcontextprotocol.io/docs/tutorials/security/security_best_practices):
  authoritative session hijacking guidance and warning that sessions must not
  be used as authentication.
- [TypeScript SDK server guide](https://github.com/modelcontextprotocol/typescript-sdk/blob/main/docs/server.md):
  official SDK guidance for stateful versus stateless Streamable HTTP,
  resumability through event stores, shutdown, and task stores.
- [Go SDK package documentation](https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp):
  official Go SDK API surface for client and server sessions, graceful close,
  wait semantics, and memory event store behavior.

## Findings

MCP lifecycle gives Workbench a useful distinction between protocol connection
state and application state. Initialization negotiates capabilities, operation
uses only negotiated capabilities, and shutdown is signaled by the underlying
transport. Workbench durable sessions should not assume that a closed stdio
connection equals a closed work session.

MCP Streamable HTTP shows a concrete session identity pattern: a server may
assign an `MCP-Session-Id`, clients must send it on later requests, missing IDs
can be rejected, terminated sessions return not found, and clients may send
HTTP DELETE to end a session. It also separates resumability from session
identity by using SSE event IDs and `Last-Event-ID` as stream cursors.

MCP tasks are a close analog for durable handles. A task has a unique ID, a
status lifecycle, timestamps, TTL, polling guidance, cancellation behavior, and
related-message metadata. Workbench sessions should copy the durable handle and
state-machine discipline, but not the exact task model because sessions are
user work threads rather than single deferred operation results.

The security guidance is directly relevant even for local-first Workbench:
session IDs must be non-deterministic, should not authenticate the caller by
themselves, and should be bound to other local authority if remote transports
are ever added. Event redelivery also creates injection risk if untrusted data
can be enqueued against another user's session.

Official SDK docs favor explicit tradeoffs. Stateful HTTP enables resumability
but requires session routing and shutdown cleanup. Stateless mode is simpler
when a server does not need per-client state. In-memory event stores are useful
for behavior but insufficient for Workbench's durable resume goal.

## Implications

Workbench should define a Workbench session record with its own ID, state, and
event history. It should avoid using MCP transport session IDs as durable
business identifiers, because stdio does not provide the same header-based
state and a process restart should not necessarily close the user's work
thread.

Resume should be explicit. A durable session checkpoint can restore `focus` and
`artifact_id`, but that restoration should flow through the existing `context`
mutation and capability sync path so clients observe the same contract as any
other context change.

Event history should use normalized, bounded payloads. The event log should be
good enough to inspect lifecycle, context, artifact, and sync events without
becoming a raw transcript or memory system.

The first storage design can remain file-backed if it provides atomic append,
bounded retention, clear corruption handling, and a migration path to later
indexed storage.

## Source References

- [RFC: Session Management](session-management-rfc.md)
- [Concept map](session-management-concept-map.md)
- [Risk: Session History Becomes Hidden Transcript Storage](session-management-risk.md)

## Open Questions

The research leaves five human decisions open, indexed in the RFC as
`SM-HIL-001` through `SM-HIL-005`.
