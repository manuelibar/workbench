---
id: "session-management-test-strategy"
type: "test_strategy"
title: "Session Management Test Strategy"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Session Management Test Strategy

## Scope

This strategy covers verification for durable Workbench session identity,
lifecycle transitions, resume and close semantics, event history, retention,
storage corruption behavior, and interaction with the existing `context`
capability sync contract.

This is an action artifact linked from the [Session Management RFC](session-management-rfc.md).

## Test Levels

Unit tests:

- Validate session ID generation and parsing.
- Validate allowed lifecycle transitions and rejected transitions.
- Verify event append ordering, monotonic event IDs, bounded payload size, and
  opaque pagination cursors.
- Verify checkpoint serialization for empty context, focus-only context,
  artifact-only context, and full context.
- Verify close is idempotent and terminal under the chosen policy.

Store tests:

- Use temporary directories for metadata and JSONL event logs.
- Simulate partial metadata writes, truncated event records, missing event
  files, and compaction.
- Verify retention deletes or compacts the intended records without changing
  active session state.

MCP integration tests:

- Begin a session and observe the current session surface.
- Select an artifact through `context`, then verify session events reflect the
  context checkpoint.
- Resume after clearing in-memory context and verify resume uses the same
  capability sync behavior as a direct `context` call.
- Close a session, repeat close, and verify resume behavior follows the RFC
  decision.

Manual checks:

- Inspect created session files to confirm they are readable and do not contain
  raw tool output or full transcripts.
- Run `go test ./...` and `go test -race ./...` before merging runtime work.

## Fixtures

Use temporary artifact directories with at least one valid artifact, one missing
artifact reference, and one invalid artifact ID. Use deterministic clocks for
session and event timestamps. Use small retention limits to force compaction in
tests without large files.

Create fixtures for corrupted JSONL lines, duplicate event IDs, closed session
records, and sessions whose latest checkpoint references an artifact that no
longer exists.

## Risks

The main blind spot is assuming file-backed behavior is reliable because unit
tests pass on one filesystem. Crash and partial-write tests must exercise the
specific atomic write pattern used by the implementation.

Another risk is testing resume as a direct state assignment instead of through
the `context` path. That would miss capability sync regressions, which are a
core Workbench contract.

Privacy tests can also drift if event schemas grow. Tests should fail when
events include raw message fields or payloads larger than the configured limit.

## Source References

- [Session Management RFC](session-management-rfc.md)
- [Initial implementation plan](session-management-initial-implementation-plan.md)
- [Risk: Session History Becomes Hidden Transcript Storage](session-management-risk.md)
- `docs/reference/context-contract.md`
- `internal/mcpserver/artifacts_test.go`
- `internal/mcpserver/mcp_integration_test.go`

## Open Questions

The exact assertions for close and retention will follow `SM-HIL-001` and
`SM-HIL-002` in the RFC Human-in-the-loop Index.
