---
id: "namespace-mgmt-requirement"
type: "requirement"
title: "Namespace Is the Required Scope Reference"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Namespace Is the Required Scope Reference

## Statement

Every future Workbench feature that needs durable scope identity must reference
a namespace ID rather than inventing a separate project, repository, workspace,
artifact group, or local-area identity.

This is an action artifact produced by
[namespace-mgmt-rfc.md](namespace-mgmt-rfc.md).

## Rationale

Workbench's value depends on progressive disclosure of relevant context and
artifacts. A single namespace reference gives consuming features one stable
scope handle while still allowing provider aliases, local paths, and
relationship metadata to change.

This requirement prevents duplicate identity systems across planning, notes,
memory, GitHub integration, artifact scoping, and local working area discovery.
It also keeps namespace management focused on identity instead of absorbing
feature-specific planning behavior.

## Acceptance Criteria

- A feature that scopes data to a project, repository, workspace, local area, or
  artifact group stores or resolves a namespace ID.
- GitHub org/repo IDs, `owner/repo` strings, codespace IDs, Git remote URLs,
  submodule paths, and filesystem paths are stored as aliases or evidence, not
  as the sole Workbench identity.
- A project can be represented as a namespace and can relate to multiple
  repositories or local areas.
- Existing unscoped artifacts remain readable during migration.
- The namespace core model contains no priority, board, WIP, sprint, or daily
  planning fields.
- Submodule modeling preserves both child repository identity and superproject
  mount metadata.

## Source References

- [Namespace management RFC](namespace-mgmt-rfc.md)
- [Namespace management concept map](namespace-mgmt-concept-map.md)
- [Namespace management risk](namespace-mgmt-risk.md)

## Open Questions

Acceptance depends on the RFC-indexed defaults for `NM-D1`, `NM-D2`, and
`NM-Q1`.
