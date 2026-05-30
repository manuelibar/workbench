---
id: "namespace-mgmt-problem-statement"
type: "problem_statement"
title: "Namespace Management Problem Statement"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Namespace Management Problem Statement

## Context

Workbench `main` is currently a small stdio MCP kernel with in-memory context
and file-backed typed artifacts. The active context can carry a focus string and
an artifact selection, but it does not yet have a durable scope identity that
can say which project, repository, working area, or artifact group the agent is
operating inside.

Future epics will need to refer to the same real-world scopes: a local project,
a GitHub organization, a repository, a submodule path, a codespace, a Git
worktree, and a collection of artifacts. If each feature invents its own
identity model, cross-feature context will drift.

## Problem

Workbench needs one namespace primitive that can identify and relate scopes
without taking ownership of planning workflows. Today, "project", "repo",
"workspace", "local directory", "artifact packet", and "current context" can all
mean scope, but they do not yet share a durable representation.

The missing contract creates ambiguity in basic questions:

- Which namespace does the current context belong to?
- Is a project separate from a namespace, or a namespace with project semantics?
- Does a local worktree have identity beyond its path?
- How should a Git submodule be represented: as a nested repository namespace,
  a relationship, or both?
- How can artifacts be grouped by scope without implying backlog or board
  ownership?

## Impact

Without a namespace contract, later features can produce incompatible
identifiers, duplicate provider metadata, and unclear ownership boundaries.
Artifacts may be hard to filter by project or repository, local working areas
may be confused with durable source identity, and external GitHub concepts may
leak planning behavior into this epic.

The impact is highest for features that need progressive disclosure: an agent
should be able to select a namespace first, then see only the relevant context,
artifacts, and local areas for that scope.

## Constraints

- This bootstrap pass is docs-only and may only update `docs/artifacts/`.
- The packet must derive from current Workbench docs, current repository state,
  and targeted authoritative research.
- The model must not cite or depend on stale implementation snapshots.
- The scope primitive must support GitHub and Git evidence without making
  Workbench's identity equal to provider IDs, repository paths, or local paths.
- Boards, prioritization, WIP, and daily planning are explicitly out of scope.
- Artifact scoping must respect the current file-backed artifact kernel until a
  later implementation pass changes runtime behavior.

## Source References

- [Workbench README](../../README.md)
- [Epic branch workflow](../how-to/epic-branch-workflow.md)
- [Artifact conventions](../reference/artifact-conventions.md)
- [Namespace management research note](namespace-mgmt-research-note.md)

## Open Questions

The problem is clear enough for a draft contract. Decisions that affect the
implementation path are indexed in [namespace-mgmt-rfc.md](namespace-mgmt-rfc.md).
