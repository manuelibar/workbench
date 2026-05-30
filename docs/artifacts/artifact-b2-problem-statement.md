---
id: "artifact-b2-problem-statement"
type: "problem_statement"
title: "Artifact B2 Workflow Problem Statement"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Artifact B2 Workflow Problem Statement

## Context

Workbench `main` is currently a small stdio MCP kernel with a deterministic
`context` tool and file-backed typed Markdown artifacts under `docs/artifacts/`.
The foundation supports creation, listing, reading, validation, guidance, and
section updates. It intentionally leaves advanced artifact lifecycle behavior to
Artifact B2.

## Problem

Artifacts can exist and validate today, but they cannot yet express the
workflow semantics needed for long-running branch work. A reviewer cannot tell
from the contract alone whether an artifact is current, archived, superseded,
deleted, awaiting sign-off, blocked on human input, or related to another
artifact in a typed way. Without a B2 contract, later runtime changes would need
to guess at governance rules and could make incompatible metadata choices.

## Impact

The gap affects agents, branch owners, and reviewers:

- Agents lack a stable way to decide whether to update, supersede, archive, or
  preserve an artifact.
- Branch owners lack a compact evidence trail for sign-off and human decisions.
- Reviewers lack relationship and lineage cues that distinguish source evidence,
  action plans, assumptions, risks, and final decisions.
- Runtime implementers lack a validated target for frontmatter evolution and
  workflow transition checks.

## Constraints

Artifact B2 must keep this bootstrap pass docs-only and must preserve the
current flat Markdown artifact model. It should prefer additive frontmatter and
body conventions over breaking schema changes. It must align with the existing
artifact contract registry, keep RFCs as living drill hubs, and avoid requiring
clients to support MCP elicitation before Workbench can represent human nudges
as ordinary artifacts.

## Source References

- [README.md](../../README.md)
- [Artifact Conventions](../reference/artifact-conventions.md)
- [Artifact B2 Concept Map](artifact-b2-concept-map.md)
- [Artifact B2 Risk](artifact-b2-risk.md)

## Open Questions

The blocking questions for this problem are indexed in the
[Artifact B2 RFC](artifact-b2-rfc.md#human-in-the-loop-index).
