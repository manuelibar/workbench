---
id: "namespace-mgmt-initial-implementation-plan"
type: "implementation_plan"
title: "Namespace Management Initial Implementation Plan"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Namespace Management Initial Implementation Plan

## Objective

Sequence the first runtime implementation pass for namespace management after
the docs packet is accepted. The objective is to add namespace scope identity
without changing planning behavior or breaking the existing context and artifact
kernel.

This is an action artifact produced by
[namespace-mgmt-rfc.md](namespace-mgmt-rfc.md).

## Steps

1. Confirm the RFC Human-in-the-loop defaults or update the RFC with accepted
   decisions for `NM-D1`, `NM-D2`, and `NM-Q1`.
2. Define the namespace data contract for records, aliases, relationships, and
   lifecycle status.
3. Add a small namespace store behind the existing Workbench kernel boundary,
   using deterministic validation and no planning fields.
4. Add discovery/import code for current local evidence: repository root,
   worktree list output, remotes, current branch, and submodule metadata.
5. Add provider evidence shape for GitHub organizations, repositories,
   codespaces, and repository contents without requiring network access in the
   core namespace contract.
6. Extend context selection with namespace awareness while preserving current
   `focus` and `artifact_id` behavior.
7. Add artifact scoping metadata or index support that can attach artifacts to a
   namespace without invalidating existing artifacts.
8. Add MCP/resource guidance so agents can inspect namespaces and relationships
   through progressive disclosure.
9. Add contract tests for namespace validation, alias merge behavior,
   relationship traversal, submodule modeling, local worktree modeling, and
   artifact scoping.
10. Document migration and rollback notes before enabling namespace selection as
    a required part of any workflow.

## Verification

Verification should include:

- Unit tests for namespace record, alias, and relationship validation.
- Fixture tests for GitHub org/repo evidence, Git worktree output, and
  submodule forms.
- Artifact contract tests proving old artifacts remain valid and new scoped
  artifacts can be resolved.
- Context tests proving namespace selection composes with `focus` and
  `artifact_id`.
- Negative tests proving priority, board column, WIP, sprint, and daily plan
  fields are rejected from namespace core records.
- Manual drill through a project namespace with multiple repository/local-area
  relationships.

## Rollback

The first implementation should be introduced behind additive data structures
and optional context selection. Rollback should be possible by disabling
namespace-aware tool/resource exposure and leaving existing `focus`,
`artifact_id`, and file-backed artifacts untouched.

If namespace metadata is written to artifacts, rollback requires preserving the
files and ignoring namespace metadata rather than deleting artifacts. If a
separate index is used, rollback can remove or rebuild the index from source
artifacts and provider/local discovery evidence.

## Source References

- [Namespace management RFC](namespace-mgmt-rfc.md)
- [Namespace scope requirement](namespace-mgmt-requirement.md)
- [Namespace management concept map](namespace-mgmt-concept-map.md)

## Open Questions

This implementation plan is blocked only by the RFC-indexed nudges `NM-D1`,
`NM-D2`, and `NM-Q1`; documented defaults unblock an initial draft if humans do
not answer before implementation begins.
