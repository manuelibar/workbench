---
id: "work-management-research-note"
type: "research_note"
title: "Workbench Work Management Research Note"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Workbench Work Management Research Note

## Question

What current local repository constraints and external work-management patterns
should shape a self-contained Workbench work-management contract for backlog,
boards, WIP, prioritization, daily planning, namespace-scoped work, active work
environment selection, and the first-session-of-day flow?

## Sources

Local sources:

- `README.md`: Defines Workbench as a local stdio MCP server with current
  context and typed Markdown artifacts, no HTTP listener, and no database
  requirement.
- `docs/how-to/epic-branch-workflow.md`: Requires epic packets to be
  self-contained, use current docs and targeted research, avoid archive
  citations, and make the RFC the living drill hub.
- `docs/reference/artifact-conventions.md`: Defines required artifact
  frontmatter, supported contract types, required section validation, and the
  Human-in-the-loop Index requirement.

External sources:

- [GitHub Docs: Planning and tracking with Projects](https://docs.github.com/en/issues/planning-and-tracking-with-projects):
  GitHub Projects supports table, board, and roadmap views over issues, pull
  requests, and draft ideas with custom fields and filters.
- [GitHub Docs: About Projects](https://docs.github.com/en/enterprise-cloud@latest/issues/planning-and-tracking-with-projects/learning-about-projects/about-projects?apiVersion=2022-11-28):
  Projects sync metadata both ways with issues and pull requests, support saved
  views, custom fields, automation, charts, templates, and status updates.
- [GitHub Docs: Customizing the board layout](https://docs.github.com/en/issues/planning-and-tracking-with-projects/customizing-views-in-your-project/customizing-the-board-layout):
  Boards can use a status or other single-select/iteration field as columns,
  drag cards between columns, and show advisory column limits.
- [GitHub Docs: Filtering projects](https://docs.github.com/en/issues/planning-and-tracking-with-projects/customizing-views-in-your-project/filtering-projects):
  Saved views can filter by fields, assignee, state, type, date, iteration, and
  missing metadata; adding items in filtered views can apply filtered metadata.
- [GitHub Docs: About tasklists](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/about-tasklists):
  Task breakdown is useful, but GitHub now points users toward sub-issues for
  dedicated related-work tracking.
- [Model Context Protocol latest specification](https://modelcontextprotocol.io/specification/latest):
  MCP separates resources, prompts, and tools, and emphasizes user consent,
  privacy, tool safety, and clear review of operations.
- [MCP Resources](https://modelcontextprotocol.io/specification/2025-11-25/server/resources):
  Resources are application-driven context with optional list-changed and
  subscription notifications.
- [MCP Tools](https://modelcontextprotocol.io/specification/2025-11-25/server/tools):
  Tools are model-controlled operations; clients should keep humans able to
  review and deny tool invocations.
- [MCP Elicitation](https://modelcontextprotocol.io/specification/2025-11-25/client/elicitation):
  Servers can request structured user input while clients retain control over
  interaction and data sharing.
- [Nulab Backlog: Project board](https://support.nulab.com/hc/en-us/articles/8732487653145-Backlog-101-Project-board):
  Product inspiration for boards as status columns, drag-to-update behavior,
  filtering, due/overdue cues, and priority ordering within columns.
- [Nulab Backlog: Project settings](https://support.nulab.com/hc/en-us/articles/8616310170009-Backlog-101-Project-settings):
  Product inspiration for custom statuses, categories, milestones, custom
  fields, templates, and integrations as project-level configuration.
- [BacklogMD](https://www.backlogmd.com/en):
  Product inspiration for an agent-readable backlog protocol in plain Markdown.
- [Vibe Kanban: Monitoring task execution](https://www.vibekanban.com/docs/core-features/monitoring-task-execution)
  and [Creating workspaces](https://www.vibekanban.com/docs/workspaces/creating-workspaces):
  Product inspiration for isolated workspaces, branch/worktree association,
  task execution monitoring, cleanup hooks, approval moments, and workspace
  notes.

## Findings

The local Workbench foundation favors a local-first, deterministic contract.
Because `main` is a stdio MCP kernel with file-backed artifacts, the initial
work-management design should avoid assuming a web UI, database, or external
sync service.

Strong external patterns converge on "one item state, many views." GitHub
Projects and Backlog both let table/board/project views reflect the same
underlying work item metadata. This argues against separate board-card and
daily-plan stores that drift away from work items.

Boards should be configurable views, not the entire work model. GitHub board
columns can be based on status or another select field; Backlog uses status
columns. For Workbench, the portable default should be status columns, with room
for later custom column fields.

WIP limits are best introduced as visible/advisory constraints. GitHub column
limits communicate desired limits and highlight excess rather than preventing
all overflow. That fits Workbench's early contract because an agent can surface
over-WIP conditions without blocking legitimate urgent work.

Priority and rank need separate concepts. Product trackers often support
custom priority fields, while Backlog also lets teams order cards within a
column. Workbench should keep "importance" and "ordering in a view" distinct so
daily planning can explain why something is selected.

The first-session-of-day flow should be an aggregation view with explicit human
control. MCP guidance makes tool mutation and elicitation sensitive: the arrive
flow can gather WIP, reminders, blockers, and priorities automatically, but
accepting or changing the day plan should be explicit and reviewable.

Namespace identity must remain a foreign dependency. Work management can own
the active work environment lens, but it should treat namespace IDs and
metadata as namespace-management inputs.

## Implications

The RFC should propose:

- A namespace-scoped work item model with priority, rank, status, blockers,
  reminders, due dates, acceptance criteria, and project references.
- Saved work views over work items for backlog, board, project, WIP, triage,
  blocked, stale, and today.
- A daily plan that references work items and records plan-specific decisions
  without copying the canonical item state.
- An arrive-at-work tool/resource contract that is deterministic for a fixed
  clock and active work environment.
- Human-in-the-loop controls for storage choice, WIP default policy, active
  namespace boundary, and plan acceptance.
- Tests that prove a single fixture can drive backlog, board, WIP, and daily
  plan views without divergent state.

## Source References

- [RFC: Work Management](work-management-rfc.md)
- [Concept Map](work-management-concept-map.md)
- [Initial Implementation Plan](work-management-initial-implementation-plan.md)
- [Test Strategy](work-management-test-strategy.md)

## Open Questions

Open human nudges are centralized in the RFC's Human-in-the-loop Index.
