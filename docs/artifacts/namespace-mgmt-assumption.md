---
id: "namespace-mgmt-assumption"
type: "assumption"
title: "Project Is a Namespace, Not a Competing Primitive"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Project Is a Namespace, Not a Competing Primitive

## Statement

Workbench should treat a project as a namespace kind or namespace role, not as
a separate identity primitive. Features may add project-specific behavior later,
but they should reference namespace identity instead of inventing parallel
project IDs.

## Evidence

The current Workbench foundation keeps `main` small and defers feature systems
to epic branches. The namespace epic owns scope identity, while future planning
or backlog epics own their own behavior. GitHub and Git evidence also separates
identity from workflow: organizations, repositories, worktrees, codespaces, and
submodules can all identify scopes without implying a board or daily plan.

Treating project as a namespace kind keeps artifact/context scoping consistent
across local and provider-backed scopes. It also allows a project to span
multiple repositories or local areas through relationships instead of forcing a
single repository-shaped project model.

## Validation Plan

Validate this assumption during implementation design by checking that:

- Namespace records can represent a standalone project with no GitHub provider.
- A project namespace can relate to one or more repository namespaces.
- Artifact and context scoping can target the project namespace directly.
- Planning-specific fields remain absent from namespace records.
- Consumers that need project behavior can layer it on top of namespace
  references.

If those checks require a separate project identity table, record the reason as
a decision before runtime work proceeds.

## Source References

- [Namespace management charter](namespace-mgmt-charter.md)
- [Namespace management concept map](namespace-mgmt-concept-map.md)
- [Namespace management RFC](namespace-mgmt-rfc.md)

## Open Questions

This assumption depends on `NM-D2` in the RFC Human-in-the-loop Index.
