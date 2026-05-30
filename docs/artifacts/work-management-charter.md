---
id: "work-management-charter"
type: "charter"
title: "Workbench Work Management Charter"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Workbench Work Management Charter

## Mission

Define the Workbench work-management contract for backlog capture, project and
work views, boards, WIP-aware flow, prioritization, daily planning, and the
first-session-of-day "arrive at work" flow for a local agent-assisted work
environment.

The epic should make Workbench useful as the place an agent can ask "what work
is active, what matters next, and what should happen today?" without absorbing
namespace identity, long-term memory, external notification integrations, or
runtime implementation details before the contract is clear.

## Scope

In scope:

- Namespace-scoped work items that reference namespace identities owned by the
  namespace-management epic.
- Backlog items, ranked backlog order, priority fields, work item status, WIP
  columns, WIP limits, blockers, due dates, reminders, and acceptance criteria.
- Board, backlog, project, namespace, WIP, triage, and daily-plan views.
- Agile-style workflows that are configurable enough for kanban-style flow,
  iteration planning, project releases, and agent task attempts.
- Selection of the active working environment for work-management purposes,
  using namespace identity rather than defining it.
- The "arrive at work" flow that summarizes notifications, reminders, WIP,
  blockers, stale work, priorities, and a proposed day plan.
- MCP-facing resources and tools that expose work state with deterministic,
  reviewable behavior.

Out of scope:

- Creating, editing, merging, deleting, or otherwise defining namespace
  identity.
- User identity, team membership, role permissions, or account management.
- External issue tracker synchronization, chat notifications, calendar sync, or
  email ingestion in the initial contract.
- Automatic work execution, branch orchestration, code changes, or test running
  on behalf of a work item.
- Artifact B2 lineage, sign-off, archive, and supersession workflows.

## Stakeholders

The primary stakeholder is the local agent workflow using Workbench as a stdio
MCP server. Secondary stakeholders are humans who want a reliable morning brief,
project/work views that survive context switches, and predictable controls over
what the agent sees or changes.

The namespace-management epic is a dependency because this epic needs stable
namespace identifiers and display metadata. Artifact B2, sessions, memory, AFK,
and future external-integration epics are adjacent systems that may consume work
state later but should not shape the initial ownership boundary.

## Success Criteria

The epic is successful when:

- Work items can be captured, ranked, prioritized, filtered, and moved through a
  configurable workflow within a namespace scope.
- Boards and daily plans are views over the same work item state rather than
  separate sources of truth.
- WIP limits are visible and testable, with advisory behavior before any hard
  blocking policy is introduced.
- The first session of the day produces a deterministic brief that a human or
  agent can inspect before choosing work.
- The active work environment references namespace-management identity without
  duplicating namespace ownership.
- MCP resources and tools can expose work state without hiding destructive
  changes or skipping human-in-the-loop controls.

## Source References

- [RFC: Work Management](work-management-rfc.md)
- [Problem Statement](work-management-problem-statement.md)
- [Concept Map](work-management-concept-map.md)
- `README.md`
- `docs/how-to/epic-branch-workflow.md`
- `docs/reference/artifact-conventions.md`

## Open Questions

Open human nudges are centralized in the RFC's Human-in-the-loop Index.
