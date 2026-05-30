---
id: "namespace-mgmt-concept-map"
type: "spec"
title: "Namespace Management Concept Map"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Namespace Management Concept Map

## Context

This concept map defines taxonomy for the namespace management epic. It is a
spec artifact because later implementation work can translate the taxonomy into
storage, context selection, validation, and relationship APIs.

The design starts from the current Workbench kernel: context is small,
artifacts are typed Markdown files, and epic packets define contracts before
runtime changes. Namespace management adds scope identity to that foundation.

## Design

### Core Terms

| Term | Definition | Notes |
|---|---|---|
| Namespace | Durable Workbench scope identity. | The primitive every scoped feature should reference. |
| Namespace kind | Classification of a namespace. | Examples: `project`, `github_org`, `github_repo`, `workspace`, `local_area`, `git_submodule`, `artifact_scope`. |
| Alias | External or human-facing identifier for a namespace. | Examples: GitHub node ID, `owner/repo`, filesystem path, display slug. |
| Relationship | Typed directed edge between namespaces. | Relationships carry source, confidence, and metadata. |
| Project namespace | A namespace with project semantics. | Project is not a separate identity root. |
| Workspace namespace | A working environment bound to a namespace. | Includes local clones, Git worktrees, and GitHub Codespaces. |
| Artifact scope | Namespace used to group artifacts and context. | Does not imply board, backlog, or planning behavior. |

### Namespace Kinds

| Kind | Represents | Stable evidence | Volatile evidence |
|---|---|---|---|
| `project` | User-recognized body of work. | Workbench namespace ID, title, aliases. | Current focus string, active local path. |
| `github_org` | GitHub organization account. | GitHub login, node ID, REST URL. | Membership counts, settings snapshots. |
| `github_repo` | GitHub repository. | `owner/name`, repository node ID, clone URLs. | Default branch, visibility, topics, permissions. |
| `workspace` | Concrete work environment. | Workbench ID, provider ID when present. | Running/stopped state, machine, container status. |
| `local_area` | Local checkout, worktree, or scratch directory. | Root path plus repository evidence at discovery time. | Current branch, dirty state, lock/prune state. |
| `git_submodule` | Repository mounted in a superproject path. | Superproject namespace, path, URL, gitlink commit. | Initialized/deinitialized working directory state. |
| `artifact_scope` | Grouping boundary for artifacts and context. | Artifact IDs and namespace relationship. | Current selected artifact. |

### Relationship Types

| Relationship | From | To | Meaning |
|---|---|---|---|
| `contains` | Parent namespace | Child namespace | Hierarchical containment, such as org to repo. |
| `maps_to` | Workbench namespace | Provider entity | External identity evidence or alias mapping. |
| `mounted_at` | Superproject repo | Submodule namespace | Child repo is mounted at a path and commit. |
| `works_in` | Workspace or local area | Project or repo namespace | Work happens in that scope. |
| `scopes` | Namespace | Artifact or context scope | Artifacts or context belong to this namespace. |
| `related_to` | Namespace | Namespace | Non-hierarchical relation with a reason. |
| `supersedes` | Namespace | Namespace | New namespace replaces an older local identity. |

### Boundary Rules

- A namespace is identity; boards and planning state are consumers, not part of
  this model.
- Provider IDs are aliases or evidence, not the Workbench primary key.
- Paths identify local areas only with supporting Git evidence. A path move
  should update the local area metadata without changing project or repository
  identity.
- A Git submodule should be represented as both a namespace and a relationship:
  the namespace identifies the mounted repository scope, and `mounted_at`
  records the superproject path and pinned commit.
- Artifact scoping starts as metadata and selection behavior. It should not
  require changing every existing artifact in one migration.

## Interfaces

The future runtime contract should expose these conceptual interfaces:

- Namespace record: `id`, `kind`, `title`, `status`, `created_at`,
  `updated_at`, `description`, and optional `metadata`.
- Namespace alias: `namespace_id`, `provider`, `key`, `value`,
  `observed_at`, and optional `source_url`.
- Namespace relationship: `id`, `from_namespace_id`, `to_namespace_id`,
  `type`, `metadata`, `source`, `created_at`, and `updated_at`.
- Context selection: a future `namespace_id` context field or equivalent
  namespace selector that composes with `focus` and `artifact_id`.
- Artifact scoping: frontmatter or index metadata that can attach an artifact
  to one namespace without breaking current artifact validation.

The first implementation pass should prefer append-only discovery and explicit
selection over implicit inheritance.

## Edge Cases

- One GitHub repository can appear in multiple local areas, worktrees, or
  codespaces.
- One local directory can be moved, deleted, pruned, or recreated while the
  project namespace remains valid.
- Submodules can be initialized, deinitialized, deleted, or pinned to commits
  that are not the remote default branch.
- A GitHub repository can transfer between owners; Workbench should preserve
  namespace identity and update aliases or relationships.
- A project can span multiple repositories without becoming a board or planning
  object.
- An artifact may initially have no namespace; migration should support
  unscoped legacy artifacts.
- Provider access can be partial. Namespace records must tolerate public-only
  metadata and permission-limited snapshots.

## Test Plan

Validate the taxonomy in later implementation work with fixtures covering:

- A single local project with one repository namespace.
- A GitHub organization containing multiple repository namespaces.
- Multiple Git worktrees for the same repository namespace.
- A superproject with one initialized submodule and one deinitialized
  submodule.
- A project namespace that spans multiple repositories.
- Existing artifacts with no namespace metadata and new artifacts scoped to a
  namespace.

Contract tests should assert that planning fields such as priority, WIP, board
column, and daily plan do not appear in namespace records or relationships.

## Source References

- [Namespace management RFC](namespace-mgmt-rfc.md)
- [Namespace management research note](namespace-mgmt-research-note.md)
- [GitHub REST repositories](https://docs.github.com/en/rest/repos/repos)
- [GitHub REST repository contents](https://docs.github.com/en/rest/repos/contents)
- [Git submodules](https://git-scm.com/docs/gitsubmodules)
- [Git worktree](https://git-scm.com/docs/git-worktree)

## Open Questions

Open implementation decisions are indexed in
[namespace-mgmt-rfc.md](namespace-mgmt-rfc.md): `NM-D1`, `NM-D2`, and `NM-Q1`.
