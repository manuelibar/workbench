---
id: "work-management-problem-statement"
type: "problem_statement"
title: "Workbench Work Management Problem Statement"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Workbench Work Management Problem Statement

## Context

Workbench `main` is currently a small stdio MCP context and artifact kernel. It
stores typed Markdown artifacts, exposes the current context, and intentionally
defers backlog, project, namespace, session, memory, AFK, and workflow systems
to epic branches.

That foundation is a good base for work management because it already treats
context as explicit state and artifacts as durable contracts. It does not yet
offer a way to capture work, decide what is next, inspect active WIP, plan the
day, or switch the active working scope for an agent-assisted session.

## Problem

Without a work-management contract, Workbench cannot answer basic operational
questions for a returning user or agent:

- Which namespace or working environment is active?
- What items are in the backlog, ready, in progress, blocked, under review, or
  done?
- Which items are highest priority, due, stale, blocked, or over WIP?
- What should be considered during the first session of the day?
- Which project or work view should the agent load without flooding context?

If these concepts are introduced piecemeal by later runtime work, boards, daily
planning, prioritization, notifications, and namespace scoping may become
separate partial systems with conflicting sources of truth.

## Impact

The lack of a work-management layer affects:

- Humans who arrive at work and need an immediate summary of WIP, reminders,
  priorities, and a proposed day plan.
- Agents that need scoped, structured work context instead of a broad artifact
  or repository scan.
- Future epics that need to attach sessions, memory, AFK follow-up,
  notifications, or external integrations to specific work items.
- Namespace management, because work needs namespace references but should not
  redefine namespace identity.

The most serious product risk is losing trust: if Workbench shows a board but
the day plan and reminders are derived from different state, the user cannot
tell which view is authoritative.

## Constraints

- This bootstrap pass is docs-only and must not introduce runtime behavior.
- Work management depends on namespace-management identity but does not own
  namespace creation, naming, hierarchy, or lifecycle.
- Current Workbench has no HTTP listener and no database requirement; the first
  contract should preserve local, deterministic operation unless later research
  justifies a heavier store.
- MCP tools that mutate work state must make changes reviewable and must leave
  room for human consent, especially for plan changes, status transitions, and
  active environment selection.
- External products are inspiration only. They should inform views, fields, and
  workflow mechanics without turning Workbench into a clone of any one tracker.

## Source References

- [Charter](work-management-charter.md)
- [Research Note](work-management-research-note.md)
- `README.md`
- `docs/how-to/epic-branch-workflow.md`
- `docs/reference/artifact-conventions.md`

## Open Questions

Open human nudges are centralized in the RFC's Human-in-the-loop Index.
