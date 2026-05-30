---
id: "work-management-concept-map"
type: "spec"
title: "Workbench Work Management Concept Map"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Workbench Work Management Concept Map

## Context

This concept map names the domain objects and relationships owned by the
work-management epic. It is a taxonomy for later specs rather than a final data
schema.

The boundary is deliberately narrow: work management owns work items, views,
flow state, ranking, daily planning, and active work selection. Namespace
management owns namespace identity. Sessions, memory, AFK, skills, and external
integrations may attach to work items later.

## Design

Core concepts:

- `Work item`: The durable unit of work. It has a stable ID, title, body,
  namespace reference, type, status, priority, rank, optional project reference,
  optional parent/child links, blockers, due date, reminders, acceptance
  criteria, and audit metadata.
- `Namespace reference`: A pointer to a namespace identity owned elsewhere. It
  scopes work items and views but does not duplicate namespace lifecycle rules.
- `Project`: A named grouping of work inside one namespace. It can provide a
  release, initiative, or product view but should not be required for every
  work item.
- `Backlog`: An ordered set of candidate work items in a namespace or project.
  Backlog rank is explicit and distinct from priority.
- `Priority`: A decision signal such as urgent, high, normal, low, or numeric
  value. Priority explains importance; rank explains ordering within a view.
- `Status`: The workflow state of a work item. The initial portable vocabulary
  should cover captured, ready, in progress, blocked, review, and done.
- `Board`: A saved view that groups work items into columns by status or another
  single-select field. A board may have advisory WIP limits per column.
- `WIP limit`: A visible threshold for an active column or lane. Initial
  behavior should warn and surface excess WIP before preventing movement.
- `Work view`: A saved query and layout over work items. Examples are backlog,
  board, project, namespace, active WIP, blocked, triage, today, and stale.
- `Daily plan`: A dated plan containing selected work items, commitments,
  reminders, deferrals, and carry-over notes for a namespace or active work
  environment.
- `Notification`: A Workbench-native signal that something changed or needs
  attention. Initial sources are stale WIP, due reminders, blocked work, and
  plan drift.
- `Reminder`: A user- or agent-created future nudge attached to a work item,
  project, daily plan, or namespace-scoped work view.
- `Arrive-at-work brief`: The first-session-of-day synthesis of active
  namespace, notifications, reminders, WIP, blockers, priority candidates, and
  proposed day plan.
- `Active work environment`: The selected namespace plus optional active
  project, board, view, and day-plan date. It is the work-management lens over
  namespace identity.

Relationships:

- A work item belongs to exactly one namespace reference.
- A work item may belong to zero or more projects, but each project belongs to
  one namespace reference.
- A board, backlog, daily plan, and saved work view are scoped by namespace and
  may optionally narrow to a project.
- A daily plan references existing work items and can also contain ad hoc plan
  notes that must be promoted to work items before they become backlog state.
- Notifications and reminders point to the object that caused them and should
  be dismissible or snoozable without losing the underlying work state.
- Active work environment selection is context state, not a new namespace.

## Interfaces

Candidate MCP resources for later runtime work:

- `workbench:///work/environment`: The active work environment and selected
  namespace reference.
- `workbench:///work/namespaces/{namespace_id}/backlog`: Ranked backlog view.
- `workbench:///work/namespaces/{namespace_id}/board/{board_id}`: Board view
  with columns, WIP counts, and visible cards.
- `workbench:///work/namespaces/{namespace_id}/today`: Current day plan and
  arrive-at-work brief.
- `workbench:///work/items/{work_item_id}`: Full work item detail.

Candidate MCP tools for later runtime work:

- `work.environment.select`: Select namespace reference and optional project or
  view for the active work environment.
- `work.arrive`: Produce the first-session-of-day brief without mutating state
  unless the caller explicitly accepts plan changes.
- `work.item.capture`: Add a work item to a namespace backlog.
- `work.item.update`: Change status, priority, rank, blockers, due date, or
  reminder metadata.
- `work.board.move`: Move one or more items between board columns and report
  WIP-limit effects.
- `work.plan.update`: Create or revise the daily plan.

These names are placeholders for contract discussion, not implemented APIs.

## Edge Cases

- A work item references a namespace ID that namespace-management no longer
  resolves. Work management should mark the item as orphaned and keep it
  readable until namespace-management defines repair behavior.
- A board column exceeds its WIP limit. Initial behavior should surface the
  violation in views and the arrive-at-work brief rather than silently blocking
  status changes.
- A work item is both blocked and due today. The daily brief should preserve
  both facts and avoid presenting it as ordinary available work.
- A reminder points to a completed item. The brief should show it as a
  completed-item reminder and allow dismissal or conversion to follow-up.
- A daily plan references a work item that moved to done after the plan was
  created. The plan should show drift and suggest removal or replacement.
- A quick-captured item lacks type, priority, or project. It should enter
  triage without forcing unnecessary metadata.
- Active environment selection is missing. The first-session flow should ask for
  or infer a namespace reference according to the cross-epic contract.

## Test Plan

Later implementation should verify:

- Work items are always namespace-scoped and can be listed by namespace,
  project, status, priority, and due/reminder fields.
- Backlog rank changes do not overwrite priority and priority changes do not
  reorder the backlog unless explicitly requested.
- Board moves update status consistently and expose advisory WIP-limit
  violations.
- Daily plans remain references over work items and report drift when referenced
  item state changes.
- The arrive-at-work brief is deterministic for a fixed clock, active
  environment, and work item fixture.
- Missing namespace identity is handled as an explicit unresolved dependency,
  not by fabricating namespace data inside work management.

## Source References

- [RFC: Work Management](work-management-rfc.md)
- [Test Strategy](work-management-test-strategy.md)
- GitHub Projects board and custom-field docs linked from the research note.
- Model Context Protocol resources, tools, and elicitation docs linked from the
  research note.

## Open Questions

Open human nudges are centralized in the RFC's Human-in-the-loop Index.
