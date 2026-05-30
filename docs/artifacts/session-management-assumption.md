---
id: "session-management-assumption"
type: "assumption"
title: "Session Management File-backed First Assumption"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Session Management File-backed First Assumption

## Statement

The first useful Workbench session implementation can be file-backed and local,
with a store interface that leaves room for SQLite or another indexed backend
only if event volume, query requirements, or concurrency prove that plain files
are insufficient.

## Evidence

The current Workbench foundation already uses file-backed Markdown artifacts
and explicitly avoids a database requirement. MCP Streamable HTTP makes
stateful transport sessions optional, and official SDK guidance treats
stateless mode as simpler when resumability is not needed.

The Go SDK also exposes an in-memory event store for resumability, which is a
useful warning: memory can prove behavior but cannot provide durable resume
after process restart. A Workbench session store should therefore persist
durable session records even if its first implementation is intentionally small.

## Validation Plan

Prototype a store with session metadata files and append-only JSONL event logs.
Verify that it can create, resume, close, list, and page events for many small
sessions without requiring a database.

Escalate to SQLite only if tests show that file locking, crash recovery,
pagination, or retention compaction become fragile enough to leak complexity
into the MCP tool handlers.

## Source References

- `README.md`
- `docs/reference/artifact-conventions.md`
- [MCP transport specification, version 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25/basic/transports)
- [TypeScript SDK server guide](https://github.com/modelcontextprotocol/typescript-sdk/blob/main/docs/server.md)
- [Go SDK package documentation](https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp)
- [Research note](session-management-research-note.md)

## Open Questions

`SM-HIL-004` in the RFC tracks the storage backend approval point.
