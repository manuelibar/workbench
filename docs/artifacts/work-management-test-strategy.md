---
id: "work-management-test-strategy"
type: "test_strategy"
title: "Work Management Test Strategy"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Work Management Test Strategy

## Scope

This strategy covers future runtime verification for namespace-scoped work
items, backlog rank, board views, WIP limits, priorities, reminders,
notifications, daily plans, active work environment selection, and the
arrive-at-work brief.

It links back to the living drill hub: [Work Management RFC](work-management-rfc.md).

Out of scope for the first test pass are external tracker sync, email/chat
notification ingestion, calendar integrations, multi-user authorization, and
automatic agent execution from work items.

## Test Levels

Unit tests should cover:

- Work item validation, including namespace reference presence, status values,
  priority values, rank uniqueness within a backlog view, and reminder dates.
- Query functions for backlog, board, active WIP, blocked, stale, due, and
  today views.
- WIP-limit calculations and advisory violation reporting.
- Daily-plan drift detection when referenced work items move or complete.
- Active work environment selection and clearing semantics.

Integration tests should cover:

- MCP resources for active environment, backlog, board, today, and work item
  detail.
- MCP tools for capture, update, board move, plan update, and arrive-at-work
  aggregation.
- Capability list changes if work-management tools or resources are gated by
  context.
- Namespace-management integration through a fake namespace provider or
  contract fixture.

System or acceptance tests should cover:

- First session of the day with reminders, stale WIP, blocked work, an over-WIP
  column, and a proposed day plan.
- Quick capture into triage followed by ranking and prioritization.
- Moving an item from ready to in progress and then to review without losing
  backlog rank or daily-plan references.
- Missing namespace reference repair behavior.

## Fixtures

Minimum deterministic fixture set:

- Fixed clock: `2026-05-30T09:00:00Z`.
- Namespace `ns-product` supplied by a fake namespace-management provider.
- Project `project-alpha` inside `ns-product`.
- Work items covering captured, ready, in progress, blocked, review, and done.
- Two in-progress items to trigger an advisory WIP-limit violation.
- One blocked item with blocker metadata.
- One due-today item and one overdue reminder.
- One daily plan that references an item later marked done to test drift.
- One orphaned work item with an unresolved namespace reference.

Fixtures should be small enough to inspect by hand and rich enough to drive
backlog, board, WIP, due, blocked, stale, and today views from the same data.

## Risks

- If fixtures are too happy-path, the arrive-at-work brief may look correct
  while missing stale WIP, blockers, or plan drift.
- If tests assert a storage implementation too early, they may freeze the wrong
  backing model before the RFC storage decision is resolved.
- If namespace-management is mocked too loosely, work management may drift into
  owning namespace identity by accident.
- If mutation tests only inspect tool responses, board, backlog, and daily-plan
  views may silently diverge.
- If time is not injectable, first-session behavior and reminder tests will be
  flaky.

## Source References

- [RFC: Work Management](work-management-rfc.md)
- [Concept Map](work-management-concept-map.md)
- [Initial Implementation Plan](work-management-initial-implementation-plan.md)
- [Risk](work-management-risk.md)

## Open Questions

Open human nudges are centralized in the RFC's Human-in-the-loop Index.
