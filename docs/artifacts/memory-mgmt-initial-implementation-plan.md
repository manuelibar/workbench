---
id: "memory-mgmt-initial-implementation-plan"
type: "implementation_plan"
title: "Memory Management Initial Implementation Plan"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Memory Management Initial Implementation Plan

## Objective

Implement the first local, docs-backed-compatible memory management capability
for Workbench after the RFC decisions are resolved. The first pass should prove
the memory contract with deterministic tools, resources, and tests before any
semantic retrieval or remote storage is introduced.

This action artifact was created from [memory-mgmt-rfc.md](memory-mgmt-rfc.md)
and should be updated as the RFC drills into runtime design.

## Steps

1. Confirm RFC decisions for write consent, scope labels, correction model,
   forgetting tombstones, and sensitivity policy.
2. Define the `MemoryRecord` schema with id, kind, scope, source, source ref,
   confidence, sensitivity, state, timestamps, and supersession links.
3. Add a local memory store interface with file-backed implementation under the
   runtime codebase in a later implementation pass.
4. Add deterministic validation for record metadata and sensitivity rejection.
5. Add `memory.remember` with explicit consent checks and structured success or
   rejection results.
6. Add `memory.recall` with filters for query, scope, kind, confidence, state,
   and limit. Return records with provenance and conflict-relevant metadata.
7. Add `memory.correct` to supersede records and keep correction provenance.
8. Add `memory.forget` to remove content from active recall and write the
   approved deletion state.
9. Add memory inspection resources or resource templates so users can inspect
   memory without injecting all records into live context.
10. Integrate memory capability visibility with the existing context and
    capability sync model only after tool and resource contracts are stable.
11. Update user-facing docs for the final runtime surface.

## Verification

- Run existing Go tests for the Workbench kernel after runtime changes.
- Add unit tests for schema validation, sensitivity rejection, scope matching,
  correction supersession, and forgetting behavior.
- Add integration tests for MCP tool calls and resource inspection.
- Add conflict tests where memory disagrees with current user instruction,
  selected artifact, sourced knowledge, session state, and another memory
  record.
- Manually inspect sample memory records to confirm no hidden prompt text or
  secrets are written.

## Rollback

The first implementation should be isolated behind memory-specific tools,
resources, and storage. Rollback should remove the memory capability from the
capability catalog and leave existing context and artifact behavior unchanged.

If memory files are created before rollback, the runtime should preserve them
without exposing them, or run an explicit user-approved export/delete path
depending on the reason for rollback.

## Source References

- [memory-mgmt-rfc.md](memory-mgmt-rfc.md)
- [memory-mgmt-concept-map.md](memory-mgmt-concept-map.md)
- [memory-mgmt-test-strategy.md](memory-mgmt-test-strategy.md)
- [docs/reference/context-contract.md](../reference/context-contract.md)

## Open Questions

No implementation-plan-specific nudges are open. Packet-level human nudges are
indexed in [memory-mgmt-rfc.md](memory-mgmt-rfc.md).
