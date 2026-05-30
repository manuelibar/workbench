---
id: "namespace-mgmt-research-note"
type: "research_note"
title: "Namespace Management Research Note"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Namespace Management Research Note

## Question

What current Workbench, GitHub, and Git facts should shape a docs-only
namespace management contract for scope identity, project-as-namespace
modeling, GitHub org/repo/workspace/submodule modeling, local working areas,
artifact/context scoping, and namespace relationships?

## Sources

- [Workbench README](../../README.md): current foundation scope and MCP surface.
- [Epic branch workflow](../how-to/epic-branch-workflow.md): bootstrap packet
  expectations and RFC drill-hub requirements.
- [Artifact conventions](../reference/artifact-conventions.md): frontmatter,
  artifact types, required sections, and human-nudge index rules.
- [GitHub REST repositories](https://docs.github.com/en/rest/repos/repos):
  repository endpoints, organization repository listing, repository metadata,
  and current REST API versioning.
- [GitHub REST organizations](https://docs.github.com/en/rest/orgs/orgs):
  organization metadata, repository URLs, members, and permissions context.
- [GitHub REST repository contents](https://docs.github.com/en/rest/repos/contents):
  content entries, submodule content fields, and default-branch lookup behavior.
- [GitHub GraphQL repositories](https://docs.github.com/en/graphql/reference/repos):
  Repository object identity, owner/name lookup, node IDs, and relationships.
- [GitHub GraphQL organizations](https://docs.github.com/en/graphql/reference/orgs):
  Organization object identity, login lookup, node IDs, and repository owner
  relationships.
- [GitHub GraphQL Projects](https://docs.github.com/en/graphql/reference/projects):
  Projects v2 as adjacent planning data that should remain a dependency, not
  namespace-owned behavior.
- [GitHub Codespaces lifecycle](https://docs.github.com/en/codespaces/about-codespaces/understanding-the-codespace-lifecycle):
  hosted workspace lifecycle, repository/branch attachment, persistence, and
  deletion behavior.
- [GitHub REST Codespaces](https://docs.github.com/en/rest/codespaces/codespaces):
  codespaces as authenticated user-owned work environments tied to repositories.
- [Git submodules](https://git-scm.com/docs/gitsubmodules): submodule,
  superproject, gitlink, `.gitmodules`, and active/deinitialized forms.
- [Git submodule command](https://git-scm.com/docs/git-submodule): submodule
  initialization, update, and inspection surface.
- [Git modules file](https://git-scm.com/docs/gitmodules): `.gitmodules`
  path, URL, and branch configuration.
- [Git worktree](https://git-scm.com/docs/git-worktree): local worktree
  identity, linked worktrees, porcelain output, lock/prune state, and shared
  repository metadata.

## Findings

Workbench is intentionally small today. It exposes context and typed artifacts,
with context currently carrying `focus` and `artifact_id`. Namespace management
therefore needs to define an additive scope contract before changing runtime
tools.

GitHub's current REST API is versioned and its repository APIs expose owner,
name, node ID, clone URLs, visibility, default branch, topics, repository
feature flags, and organization repository listings. GitHub organizations act
as repository owners with members and settings. These are strong external
aliases and evidence for namespace discovery, but they should not replace
Workbench-owned namespace IDs.

GitHub GraphQL exposes Organization and Repository as node-backed objects and
Projects v2 as a separate planning-oriented graph. This supports keeping
organization and repository identity in namespace management while treating
Projects, issues, and planning workflows as downstream consumers.

GitHub Codespaces are work environments associated with repositories, branches,
and authenticated users. Their lifecycle includes creation, stop/start,
rebuild, timeout, and deletion. That makes them better modeled as workspace or
local-area namespaces related to a repository namespace, not as the repository
identity itself.

Git submodules have their own repository history and are embedded in a
superproject. The superproject records a gitlink commit and `.gitmodules`
metadata such as path and URL. Submodule state can be initialized,
deinitialized, deleted, or active. The namespace model should therefore capture
both a child repository namespace and a `mounted_at` relationship with path,
URL, and pinned commit metadata.

Git worktrees let one repository have a main worktree and zero or more linked
worktrees. Worktree state includes path, checked-out revision, branch or
detached HEAD, lock state, and prune state. A worktree is a local working area
for a repository namespace, and its local path should not be the durable
identity of the project or repository.

## Implications

The namespace contract should:

- Use Workbench-owned stable namespace IDs with provider IDs, `owner/repo`,
  local paths, and Git metadata recorded as aliases or evidence.
- Model project as a namespace kind or role so project scope can exist without
  GitHub and can span more than one repository.
- Model GitHub orgs and repos as namespace kinds, with org-to-repo containment
  and provider aliases.
- Model codespaces, clones, and worktrees as workspace/local-area namespaces
  related to repository or project namespaces.
- Model submodules as both namespaces and explicit relationships from a
  superproject namespace to the child repository namespace.
- Add artifact/context scoping as namespace selection behavior without forcing
  planning concepts into the namespace model.
- Keep GitHub Projects, issues, boards, WIP, and daily planning out of the
  namespace core.

## Source References

- [Namespace management concept map](namespace-mgmt-concept-map.md)
- [Namespace management RFC](namespace-mgmt-rfc.md)

## Open Questions

The research supports the default model. `NM-D1`, `NM-D2`, and `NM-Q1` remain
human nudges because they set implementation defaults.
