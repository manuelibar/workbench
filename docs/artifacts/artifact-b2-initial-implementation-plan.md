---
id: "artifact-b2-initial-implementation-plan"
type: "implementation_plan"
title: "Artifact B2 Initial Implementation Plan"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Artifact B2 Initial Implementation Plan

## Objective

Sequence the first runtime implementation pass for Artifact B2 after the RFC
questions are resolved or their defaults are accepted. The objective is to add
advanced artifact workflow semantics without breaking current Markdown artifact
validation or changing unrelated Workbench behavior.

## Steps

1. Confirm the RFC defaults or decisions for lifecycle states, additive
   frontmatter fields, sign-off gates, elicitation mode, and relationship
   rollout.
2. Add Markdown fixtures for active, archived, superseded, tombstoned,
   sign-off-pending, and relationship-rich artifacts.
3. Extend parsing to preserve optional B2 frontmatter fields without requiring
   them on existing artifacts.
4. Add validation helpers for lifecycle values, successor/predecessor links,
   relationship edge shape, and sign-off records.
5. Add list/read behavior for archived and superseded artifacts, keeping
   existing default reads stable.
6. Add archive, supersession, and tombstone mutation paths only after validators
   and audit fields exist.
7. Evaluate MCP elicitation integration separately, using capability detection
   and sensitive-data boundaries from the MCP elicitation spec.

## Verification

Verification should include `go test ./...`, fixture validation for every B2
lifecycle state, backward compatibility checks for current foundation artifacts,
and negative tests for invalid transitions, missing successors, cyclic
supersession, malformed relationship edges, and missing sign-off on governed
operations.

## Rollback

Rollback should be possible by reverting the implementation commit because B2
metadata is additive. If runtime parsing ships before mutation tools, existing
artifacts continue to validate with the current required frontmatter. If a
mutation tool ships with a defect, disable or revert that tool before changing
the stored Markdown so archived or tombstoned records remain readable.

## Source References

- [Artifact B2 RFC](artifact-b2-rfc.md)
- [Artifact B2 Concept Map](artifact-b2-concept-map.md)
- [Artifact B2 Test Strategy](artifact-b2-test-strategy.md)
- [Artifact Conventions](../reference/artifact-conventions.md)

## Open Questions

This plan depends on the RFC Human-in-the-loop Index entries `B2-D1`, `B2-D2`,
`B2-Q1`, `B2-T1`, and `B2-C1`.
