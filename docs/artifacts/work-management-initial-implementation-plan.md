---
id: "work-management-initial-implementation-plan"
type: "implementation_plan"
title: "Work Management Initial Implementation Plan"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Work Management Initial Implementation Plan

## Objective

Sequence the first runtime drill for Workbench work management after the RFC
questions that affect storage and namespace integration are answered.

This plan links back to the living drill hub: [Work Management RFC](work-management-rfc.md).

## Steps

1. Resolve or explicitly default the RFC Human-in-the-loop Index items for
   storage, namespace boundary, default workflow columns, notification sources,
   and arrive-at-work mutation semantics.
2. Draft a focused work item storage spec that defines IDs, namespace
   references, required fields, optional fields, ranking, reminders, and audit
   metadata.
3. Add read-only fixtures for one namespace, one project, one backlog, one
   board, several work items, stale WIP, blocked work, due reminders, and a
   daily plan with drift.
4. Implement read-only listing/query logic for namespace backlog, work item
   detail, WIP, blocked, stale, due, and today views.
5. Expose MCP resources for active work environment, namespace backlog, board,
   today, and work item detail.
6. Add mutation tools in small slices: capture item, update item metadata,
   move board card, update plan, and select active work environment.
7. Add `work.arrive` as a deterministic aggregation over the active environment
   and work item fixtures, with suggested plan output separated from accepted
   plan mutation.
8. Integrate namespace-management identity only through its published contract.
9. Add user-facing validation and repair behavior for unresolved namespace
   references, invalid ranks, over-WIP columns, and stale plan references.

## Verification

Before considering the first implementation complete:

- Run the repository test suite.
- Validate that work item fixtures produce consistent backlog, board, WIP,
  blocked, due, and today views.
- Verify that board moves update the same work item state used by backlog and
  daily-plan views.
- Verify that `work.arrive` is deterministic under a fixed clock and does not
  mutate plan state unless explicit acceptance is supplied.
- Verify that unresolved namespace references are visible and repairable.
- Verify that MCP resource/tool list changes follow the existing context
  synchronization expectations.

## Rollback

Keep the first runtime slices independently revertable:

- Revert MCP resource exposure without deleting stored work items.
- Revert mutation tools while preserving read-only views.
- Revert `work.arrive` aggregation independently from work item storage.
- If the storage model proves wrong, write a migration or exporter before
  changing the canonical representation.

Because this bootstrap packet is docs-only, rollback for this artifact is a
single docs revert before runtime work begins.

## Source References

- [RFC: Work Management](work-management-rfc.md)
- [Concept Map](work-management-concept-map.md)
- [Test Strategy](work-management-test-strategy.md)
- [Assumption](work-management-assumption.md)

## Open Questions

Open human nudges are centralized in the RFC's Human-in-the-loop Index.
