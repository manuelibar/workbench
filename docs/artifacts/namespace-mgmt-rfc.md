---
id: "namespace-mgmt-rfc"
type: "rfc"
title: "Namespace as Workbench Scope Identity"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Namespace as Workbench Scope Identity

## Summary

Introduce namespace as the durable Workbench scope primitive for project,
GitHub, Git, local working area, artifact, and context identity. A namespace is
not a planning object. It is the identity layer that later planning, notes,
memory, GitHub, and artifact features can reference.

This RFC is the living drill hub for namespace management. Stable packet
artifacts:

- [Charter](namespace-mgmt-charter.md)
- [Problem Statement](namespace-mgmt-problem-statement.md)
- [Concept Map](namespace-mgmt-concept-map.md)
- [Assumption](namespace-mgmt-assumption.md)
- [Risk](namespace-mgmt-risk.md)
- [Research Note](namespace-mgmt-research-note.md)

Action artifacts:

- [Initial Implementation Plan](namespace-mgmt-initial-implementation-plan.md)
- [Namespace Scope Requirement](namespace-mgmt-requirement.md)

## Problem

Workbench can currently select a focus string and artifact, but it lacks a
stable scope identity that can bind context, artifacts, local work areas, and
external provider evidence. Future epics need to refer to the same project,
repository, organization, worktree, codespace, or submodule without each epic
inventing its own identity system.

The problem is scope identity, not planning. Namespace management must avoid
owning boards, prioritization, WIP, daily planning, or backlog semantics.

## Proposal

Define `namespace` as the single core scope primitive.

Core proposal:

- A namespace has a Workbench-owned stable ID, kind, title, status, metadata,
  aliases, and relationships.
- A project is represented as a namespace kind or role, not as a separate
  identity root.
- GitHub organizations and repositories are namespace kinds with provider
  aliases such as login, `owner/repo`, REST URL, and GraphQL node ID.
- GitHub Codespaces, local clones, Git worktrees, and scratch directories are
  workspace or local-area namespaces related to the durable project or
  repository namespace.
- Git submodules are represented as child repository namespaces plus a
  relationship from the superproject that records mount path, `.gitmodules`
  URL, and pinned gitlink commit.
- Artifact and context scoping should reference namespace identity while
  preserving existing artifact files until an implementation pass migrates or
  indexes them.
- Namespace relationships are explicit typed edges; default behavior should not
  silently inherit artifacts or context from parents.

Initial namespace kinds:

| Kind | Purpose |
|---|---|
| `project` | User-recognized body of work or product area. |
| `github_org` | GitHub organization account. |
| `github_repo` | GitHub repository identity. |
| `workspace` | Hosted or editor-backed working environment, including codespaces. |
| `local_area` | Local checkout, Git worktree, or scratch directory. |
| `git_submodule` | Repository mounted inside a superproject path. |
| `artifact_scope` | Artifact grouping and context boundary. |

Initial relationship types:

| Type | Purpose |
|---|---|
| `contains` | Parent scope contains child scope. |
| `maps_to` | Workbench namespace maps to provider evidence. |
| `mounted_at` | Repository is mounted inside a superproject. |
| `works_in` | Workspace or local area is used to work in another namespace. |
| `scopes` | Namespace scopes an artifact or context selection. |
| `related_to` | Non-hierarchical relation with an explicit reason. |
| `supersedes` | Namespace replaces an older namespace identity. |

## Tradeoffs

Using Workbench-owned IDs avoids binding identity to provider IDs or filesystem
paths, but it requires alias management and merge behavior when discovery finds
the same entity through multiple routes.

Treating project as a namespace keeps the identity model small and composable,
but it means project-specific behavior must be layered by other epics instead
of being embedded in the namespace record.

Explicit relationships avoid surprising inherited context, but consumers may
need a relationship traversal API before they can answer questions such as
"show artifacts for this project and its repositories."

Modeling submodules as both namespaces and relationships is more verbose than a
single nested record, but it preserves the child repository identity and the
superproject mount evidence separately.

## Rollout

1. Land this docs-only bootstrap packet on `epic/namespace-mgmt`.
2. Use [namespace-mgmt-initial-implementation-plan.md](namespace-mgmt-initial-implementation-plan.md)
   as the first implementation guide after packet review.
3. Confirm the open human nudges or apply their documented defaults.
4. Draft runtime data contracts for namespace records, aliases, relationships,
   context selection, and artifact scoping.
5. Implement discovery/index behavior in small steps that preserve current
   `context` and artifact behavior until namespace selection is ready.
6. Add validation and contract tests that keep planning fields out of namespace
   records.
7. Hand stable namespace IDs to consuming epics as a dependency.

## Open Questions

These are the current human nudges for the packet. Defaults are intentionally
implementation-ready so the epic can proceed if no answer is provided.

### Human-in-the-loop Index

| ID | Nudge | Type | Why it matters | Blocks | Default if unanswered |
|---|---|---|---|---|---|
| NM-D1 | Should primary namespace IDs be opaque Workbench IDs or human-readable slugs? | decision | The choice affects merges, provider renames, path moves, URLs, and future API ergonomics. | Runtime schema and migration design. | Use opaque stable Workbench IDs and store slugs/provider keys as aliases. |
| NM-D2 | Should `project` be a namespace kind/role instead of a separate primitive? | decision | This decides whether project scope composes with repo, local, artifact, and context scope through one identity system. | Project-as-namespace schema and consuming epic contracts. | Treat project as a namespace kind or role, not a separate identity root. |
| NM-Q1 | Should child namespaces inherit artifact/context scope from parents by default? | tradeoff | Automatic inheritance is convenient, but can expose unrelated artifacts or context in multi-repo and submodule setups. | Context/artifact scoping behavior. | No automatic inheritance; require explicit relationship traversal by consumers. |

## Source References

- [Workbench README](../../README.md)
- [Epic branch workflow](../how-to/epic-branch-workflow.md)
- [Artifact conventions](../reference/artifact-conventions.md)
- [Namespace management research note](namespace-mgmt-research-note.md)
- [GitHub REST repositories](https://docs.github.com/en/rest/repos/repos)
- [GitHub REST organizations](https://docs.github.com/en/rest/orgs/orgs)
- [GitHub REST repository contents](https://docs.github.com/en/rest/repos/contents)
- [GitHub GraphQL repositories](https://docs.github.com/en/graphql/reference/repos)
- [GitHub GraphQL organizations](https://docs.github.com/en/graphql/reference/orgs)
- [GitHub GraphQL Projects](https://docs.github.com/en/graphql/reference/projects)
- [GitHub Codespaces lifecycle](https://docs.github.com/en/codespaces/about-codespaces/understanding-the-codespace-lifecycle)
- [GitHub REST Codespaces](https://docs.github.com/en/rest/codespaces/codespaces)
- [Git submodules](https://git-scm.com/docs/gitsubmodules)
- [Git modules file](https://git-scm.com/docs/gitmodules)
- [Git worktree](https://git-scm.com/docs/git-worktree)
