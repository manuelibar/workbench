---
id: "artifact-b2-assumption"
type: "assumption"
title: "Artifact B2 Preserves Markdown as the Source of Truth"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Artifact B2 Preserves Markdown as the Source of Truth

## Statement

Artifact B2 can add advanced workflow semantics while keeping Markdown files in
`docs/artifacts/` as the durable source of truth. Any future runtime index,
cache, or resource view should be rebuildable from artifact files.

## Evidence

The current Workbench foundation stores typed artifacts as Markdown and exposes
them through artifact tools and resources. The artifact conventions require
stable frontmatter keys, deterministic validation, and non-empty required
sections. The epic branch workflow also requires self-contained kickoff packets
under `docs/artifacts/` before runtime work.

This evidence supports an additive approach: B2 can define lifecycle,
relationship, sign-off, and elicitation metadata in Markdown first, then let a
later implementation pass decide which parts become validated runtime fields.

## Validation Plan

Validate the assumption by designing B2 fixtures that represent archived,
superseded, deleted, sign-off-pending, and relationship-rich artifacts using
Markdown alone. The assumption holds if those fixtures can express the workflow
state unambiguously and can be parsed without losing current artifact
validation behavior.

## Source References

- [README.md](../../README.md)
- [Epic Branch Workflow](../how-to/epic-branch-workflow.md)
- [Artifact Conventions](../reference/artifact-conventions.md)
- [Artifact B2 Test Strategy](artifact-b2-test-strategy.md)

## Open Questions

If this assumption fails, `B2-C1` in the
[Artifact B2 RFC](artifact-b2-rfc.md#human-in-the-loop-index) becomes the
decision point for whether relationships need an index before runtime work.
