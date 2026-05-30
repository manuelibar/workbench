---
id: "namespace-mgmt-charter"
type: "charter"
title: "Namespace Management Epic Charter"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Namespace Management Epic Charter

## Mission

Define namespace as Workbench's core scope primitive so an agent can attach
context, artifacts, local working areas, and external source identities to a
stable scope before runtime implementation begins.

## Scope

In scope:

- Namespace identity, naming, aliases, kind taxonomy, and lifecycle states.
- Project-as-namespace modeling, where a project is a namespace role or kind
  rather than a separate root primitive.
- GitHub organization, repository, codespace/workspace, and repository
  contents modeling when those entities provide source identity.
- Git submodule and superproject relationships, including path, URL, and pinned
  commit metadata.
- Local working areas such as clones, Git worktrees, codespaces, and scratch
  directories when they bind agent activity to a namespace.
- Artifact and context scoping rules that explain how future Workbench tools
  should resolve the active namespace.
- Typed namespace relationships such as contains, maps-to, mounted-at,
  works-in, supersedes, and related-to.

Out of scope:

- Boards, backlog ordering, prioritization, WIP limits, daily planning, sprint
  rituals, or work item scheduling.
- Runtime implementation, database schema migration, MCP tool changes, or UI
  changes during this bootstrap pass.
- Advanced artifact lineage, archive, supersession, elicitation, sign-off, and
  review workflows.

## Stakeholders

The primary stakeholder is the local agent workflow using Workbench as a stdio
MCP context and artifact kernel. Secondary stakeholders are future epic owners
that need a shared scope vocabulary for backlog, notes, memory, GitHub, and
artifact features without making those features own namespace identity.

The namespace epic owner is accountable for the scope identity contract. Feature
epics that consume namespace identity remain accountable for their own behavior,
such as planning, note capture, or issue synchronization.

## Success Criteria

The epic is successful when:

- Workbench has a durable namespace contract that can represent local,
  artifact, project, GitHub, Git, and workspace scopes with one primitive.
- Project, repository, organization, submodule, and local working area concepts
  are modeled as namespace kinds or relationships instead of competing identity
  systems.
- Artifacts and current context can be scoped to a namespace without requiring
  boards, prioritization, WIP, or daily planning concepts.
- External provider metadata is captured as evidence or aliases, while
  Workbench keeps its own stable namespace identity.
- Human decisions that affect implementation are indexed in
  [namespace-mgmt-rfc.md](namespace-mgmt-rfc.md).

## Source References

- [Workbench README](../../README.md)
- [Epic branch workflow](../how-to/epic-branch-workflow.md)
- [Artifact conventions](../reference/artifact-conventions.md)
- [Namespace management RFC](namespace-mgmt-rfc.md)

## Open Questions

Human nudges are tracked in the RFC Human-in-the-loop Index:
`NM-D1`, `NM-D2`, and `NM-Q1`.
