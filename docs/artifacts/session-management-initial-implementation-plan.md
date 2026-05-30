---
id: "session-management-initial-implementation-plan"
type: "implementation_plan"
title: "Session Management Initial Implementation Plan"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Session Management Initial Implementation Plan

## Objective

Implement the smallest durable Workbench session layer that can create, inspect,
resume, and close local work sessions while preserving the current context and
artifact kernel boundaries.

This plan is an action artifact produced by the [Session Management RFC](session-management-rfc.md).

## Steps

1. Define internal session domain types: `SessionID`, `SessionState`,
   `SessionRecord`, `ContextCheckpoint`, `SessionEvent`, and pagination cursor.
2. Add a store interface with create, get, list, append event, update
   checkpoint, close, and compact methods.
3. Implement the first file-backed store behind that interface, using atomic
   metadata writes and append-only event logs.
4. Add lifecycle rules for active, suspended, and closed states, including
   idempotent close and explicit errors for non-resumable sessions.
5. Route resume through the existing `context` mutation path so capability sync
   behavior stays consistent.
6. Add MCP tools for the smallest useful surface: begin, current, list, get,
   resume, close, and event pagination.
7. Add resource exposure for current session and selected session only after
   the tool surface is stable.
8. Document configuration, retention defaults, and privacy boundaries in
   foundation docs after runtime behavior lands.

## Verification

Run unit tests for store behavior, lifecycle transitions, event ordering,
checkpoint serialization, retention compaction, and close idempotency.

Run MCP integration tests that begin a session, select an artifact through
`context`, append session events, restart the store, resume the session, and
verify the context result includes expected capability sync behavior.

Run race tests around concurrent event append and close attempts if the store
uses locks or background compaction.

## Rollback

Keep all runtime changes behind newly introduced session tools and resources.
If the first implementation is not stable, remove the session tool
registration and leave the store package unused while preserving the docs
packet for the next implementation pass.

Because the session store is additive and outside `docs/artifacts/`, rollback
should not require artifact migration. If a later migration changes the storage
layout, ship a one-way importer only after the old format remains readable for
one release window.

## Source References

- [Session Management RFC](session-management-rfc.md)
- [Concept map](session-management-concept-map.md)
- [Test strategy](session-management-test-strategy.md)
- `docs/reference/context-contract.md`
- `internal/mcpserver/contracts.go`

## Open Questions

The implementation sequence depends on the RFC decisions for retention,
payload policy, storage backend, and resume default.
