---
id: "work-management-rfc"
type: "rfc"
title: "Workbench Work Management RFC"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Workbench Work Management RFC

## Summary

Define Workbench work management as a namespace-scoped layer for work items,
backlog ranking, boards, WIP visibility, prioritization, daily planning,
project/work views, active work environment selection, and the first-session
"arrive at work" brief.

This RFC is the living drill hub for the epic. Action artifacts currently
linked from this hub:

- [Initial Implementation Plan](work-management-initial-implementation-plan.md)
- [Test Strategy](work-management-test-strategy.md)

Supporting packet artifacts:

- [Charter](work-management-charter.md)
- [Problem Statement](work-management-problem-statement.md)
- [Concept Map](work-management-concept-map.md)
- [Assumption](work-management-assumption.md)
- [Risk](work-management-risk.md)
- [Research Note](work-management-research-note.md)

## Problem

Workbench has a context and artifact kernel but no contract for operational
work. A user can store artifacts, but cannot ask Workbench to show active WIP,
rank the backlog, inspect a board, plan the day, select a namespace-scoped work
environment, or arrive at work and get a concise summary of what needs
attention.

Adding these features without a shared contract would likely create drift:
boards could store independent card state, daily plans could copy work items,
and namespace selection could conflict with the namespace-management epic.

## Proposal

Create a work-management domain that owns work item state and derives backlog,
board, project, WIP, triage, and daily-plan views from that state.

Proposed model:

- Work items are namespace-scoped by reference to namespace-management identity.
- Work items carry status, priority, rank, blockers, due date, reminder,
  acceptance, and optional project metadata.
- Backlog order is explicit and separate from priority.
- Boards are saved views grouped by status or another future column field.
- WIP limits are advisory at first and visible in board and arrive-at-work
  summaries.
- Daily plans reference work items and record plan decisions without replacing
  canonical work item state.
- Active work environment is a single context object: namespace reference plus
  optional project, board/view, and plan date.
- The arrive-at-work brief gathers active environment, unread Workbench-native
  notifications, reminders, overdue/due items, stale WIP, blockers,
  over-limit WIP, priority candidates, and a proposed day plan.

Candidate MCP surface for later implementation:

- Resources for active environment, namespace backlog, namespace board, today's
  plan, and individual work item detail.
- Tools for selecting active work environment, capturing work, updating item
  metadata, moving board cards, revising the daily plan, and producing the
  arrive-at-work brief.
- Human review for mutating tools, with arrive-at-work defaulting to
  read-only/suggested output unless the caller explicitly accepts changes.

The first implementation should prefer local deterministic state. The exact
backing model remains an open decision because artifact-backed work items,
separate Markdown files, and a lightweight embedded store each have tradeoffs.

## Tradeoffs

Starting with local deterministic state keeps Workbench aligned with the current
kernel and easier to verify. It may delay advanced search, sync, and
multi-user reporting that a database-backed tracker could provide.

Keeping boards as views avoids drift but makes board customization depend on a
well-designed work item field model. A separate card model would be faster to
prototype but harder to reconcile with daily planning and arrive-at-work.

Advisory WIP limits are less strict than hard constraints. They fit early agent
workflow because the system can explain over-limit work without blocking urgent
human choices.

Putting active work environment in this epic gives work management the flow it
needs, but it requires a careful boundary with namespace-management so this
epic does not redefine namespace identity.

External product inspiration is useful for views and metadata, but Workbench
must remain an MCP-oriented local system rather than a hosted project tracker.

## Rollout

1. Use this packet as the contract base for the epic branch.
2. Resolve the Human-in-the-loop Index items that block runtime shape:
   namespace boundary, storage model, default workflow columns, notification
   sources, and plan acceptance semantics.
3. Draft a focused runtime spec for work item storage and MCP resources/tools.
4. Implement read-only work item listing and derived views first.
5. Add mutation tools for capture, update, board move, and plan update with
   reviewable outputs.
6. Add the arrive-at-work brief with deterministic clock fixtures and no
   external integrations.
7. Integrate namespace-management identity once its contract is available.
8. Defer external tracker sync, chat/calendar/email notifications, automation,
   and agent execution orchestration to later epics or RFCs.

## Open Questions

### Human-in-the-loop Index

| ID | Nudge | Type | Why it matters | Blocks | Default if unanswered |
|---|---|---|---|---|---|
| WM-HIL-001 | Decide whether the first arrive-at-work flow may mutate the daily plan automatically or must require explicit acceptance. | decision | The morning brief can either be a passive report or an opinionated planning action. Users need to know when state changes. | `work.arrive`, `work.plan.update`, first-session tests | Produce a suggested plan only; require explicit acceptance before changing plan state. |
| WM-HIL-002 | Choose the initial work item backing model: artifact-backed, separate Markdown files, or lightweight embedded store. | decision | Storage affects validation, diffs, ranking updates, query performance, and later sync. | Work item schema, mutation tools, fixtures | Start with a deterministic file-backed model and prove one fixture can drive all views before adding a database. |
| WM-HIL-003 | Approve default workflow columns and initial WIP behavior. | approval | Defaults shape user expectations and test fixtures even if workflows become configurable later. | Board defaults, WIP-limit tests, arrive-at-work over-WIP summary | Use captured, ready, in progress, blocked, review, done; advisory WIP limit of one for in progress. |
| WM-HIL-004 | Challenge the namespace boundary: active work environment may select a namespace reference but must not create or redefine namespace identity. | challenge | This prevents work management from becoming an incompatible namespace system. | Active environment contract, cross-epic integration | Consume namespace IDs and display metadata from namespace-management; mark unresolved references as repairable. |
| WM-HIL-005 | Decide initial notification and reminder sources for the arrive-at-work brief. | question | The brief is only useful if it includes the right signals without over-claiming unsupported integrations. | Notification model, reminder fixtures, first-session acceptance criteria | Use Workbench-native due dates, reminders, stale WIP, blocked items, plan drift, and over-WIP signals only. |
| WM-HIL-006 | Decide whether rank and priority are both required on work items or whether one can be optional at capture. | tradeoff | Strict metadata improves planning but slows quick capture and triage. | Capture tool schema, triage view, backlog sorting | Require rank when an item enters a backlog view; allow priority to be unset until triage. |

## Source References

- [Research Note](work-management-research-note.md)
- [Concept Map](work-management-concept-map.md)
- `README.md`
- `docs/how-to/epic-branch-workflow.md`
- `docs/reference/artifact-conventions.md`
- [GitHub Projects docs](https://docs.github.com/en/issues/planning-and-tracking-with-projects)
- [MCP latest specification](https://modelcontextprotocol.io/specification/latest)
- [Nulab Backlog project board docs](https://support.nulab.com/hc/en-us/articles/8732487653145-Backlog-101-Project-board)
- [BacklogMD](https://www.backlogmd.com/en)
- [Vibe Kanban docs](https://www.vibekanban.com/docs/core-features/monitoring-task-execution)
