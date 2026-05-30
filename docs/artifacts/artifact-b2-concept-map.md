---
id: "artifact-b2-concept-map"
type: "spec"
title: "Artifact B2 Concept Map"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Artifact B2 Concept Map

## Context

The current artifact kernel treats each artifact as one typed Markdown file with
required frontmatter and type-specific required sections. Artifact B2 adds a
workflow layer around those files. This concept map is a taxonomy, not a runtime
change.

## Design

Artifact B2 groups the advanced workflow layer into eight related concepts:

| Concept | Meaning | Expected B2 responsibility |
|---|---|---|
| Lifecycle | The current workflow state of an artifact. | Define valid states and transitions for active, archived, superseded, and deleted-like records. |
| Archive | A reversible or non-destructive removal from active work. | Preserve readable content while hiding or de-prioritizing the artifact in active lists. |
| Delete | A destructive or tombstoning operation. | Require explicit governance and define whether content is retained, redacted, or removed. |
| Lineage | The chain of derivation between artifacts. | Track source, successor, predecessor, and generated-from relationships. |
| Supersession | A replacement relationship that changes which artifact is current. | Mark old and new artifacts consistently and preserve reviewer navigation. |
| Elicitation | A structured human nudge needed to continue work. | Represent questions, decisions, approvals, challenges, and tradeoffs in artifacts before runtime elicitation exists. |
| Sign-off | A recorded human or role approval. | Capture who approved what, when, and under which artifact version or state. |
| Governance | The rules that make workflow transitions reviewable. | Define required reasons, sign-offs, validation checks, and audit evidence. |

The concepts deliberately overlap. For example, supersession is both a
relationship and a lifecycle transition; delete is both lifecycle and governance;
elicitation may produce sign-off, a decision record, or a revised plan.

## Interfaces

Artifact B2 should evolve the artifact contract through additive fields and
sections that remain readable as Markdown:

| Interface area | Candidate contract shape | Notes |
|---|---|---|
| Frontmatter lifecycle | `lifecycle`, `archived_at`, `deleted_at`, `superseded_by` | Names and enum values are pending `B2-D1`. |
| Frontmatter lineage | `supersedes`, `derived_from`, `parent_artifacts` | Prefer artifact ID arrays over prose-only links. |
| Frontmatter governance | `review_state`, `signoffs`, `governance_policy` | Must not make current required frontmatter invalid. |
| Body relationships | `## Relationships` tables | Useful for rich edge labels before validation supports frontmatter arrays. |
| Body elicitation | `## Human Nudges` or RFC index rows | The RFC remains the packet-level index. |
| Packet grouping | Optional semantic directories under `docs/artifacts/` | Pending `B2-Q2`; current kernel treats artifacts as flat files. |
| MCP resources | `workbench:///artifacts/{id}` and selected artifact resources | Existing resource shape remains the baseline for reads. |
| MCP tools | Existing `artifact.*` tools, with later additive workflow tools if justified | Sensitive transitions should keep human confirmation explicit. |

Relationship vocabulary should start small:

| Relationship | Direction | Meaning |
|---|---|---|
| `supersedes` | new artifact to older artifact | The new artifact replaces the older one as current. |
| `superseded_by` | older artifact to new artifact | The older artifact points to its replacement. |
| `depends_on` | consumer to prerequisite | The artifact should not be completed until the prerequisite is resolved. |
| `implements` | action artifact to requirement or RFC | The action artifact carries out the referenced intent. |
| `tests` | test artifact to target artifact | The test strategy or evidence verifies the target. |
| `mitigates` | plan or constraint to risk | The artifact reduces a named risk. |
| `derived_from` | artifact to source evidence | The artifact is based on another artifact or source. |
| `related_to` | either direction | A weak edge used only when no stronger relation fits. |

## Edge Cases

The B2 contract must define behavior for stale links, missing successors,
cycles in supersession, multiple successors, attempted deletion of an artifact
that is still depended on, conflicting sign-offs, unanswered human nudges, and
legacy artifacts that only have the current required frontmatter. It should also
specify whether archived artifacts appear in `artifact.list` by default and how
resource reads behave for tombstoned artifacts.

## Test Plan

Verification should start with docs fixtures and contract validation before any
runtime change. Tests should cover frontmatter parsing for additive fields,
relationship normalization, transition validation, RFC human-nudge indexing,
sign-off recording, and backward compatibility with the three current foundation
artifacts. It should also cover packet subdirectories if B2 allows semantic
grouping, including recursive discovery, duplicate filenames in different
directories, and stable artifact IDs that do not change when files move.

## Source References

- [Artifact Conventions](../reference/artifact-conventions.md)
- [Context Contract](../reference/context-contract.md)
- [MCP Resources, 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25/server/resources)
- [MCP Tools, 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25/server/tools)
- [Artifact B2 RFC](artifact-b2-rfc.md)

## Open Questions

The relationship, lifecycle, and packet grouping decisions are tracked as
`B2-D1`, `B2-D2`, `B2-C1`, and `B2-Q2` in the
[Artifact B2 RFC](artifact-b2-rfc.md#human-in-the-loop-index).
