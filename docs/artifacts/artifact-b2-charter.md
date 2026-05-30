---
id: "artifact-b2-charter"
type: "charter"
title: "Artifact B2 Workflow Layer Charter"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Artifact B2 Workflow Layer Charter

## Mission

Define the second-stage artifact workflow layer for Workbench while preserving
the current file-backed artifact kernel. Artifact B2 turns durable Markdown
artifacts into governed work objects with explicit lifecycle state, lineage,
supersession, delete/archive semantics, human elicitation, sign-off, evolved
frontmatter, typed relationships, and review governance.

## Scope

In scope for this epic packet: the target contract, vocabulary, transition
rules, metadata model, human-in-the-loop questions, implementation sequencing,
and verification strategy for advanced artifact workflows.

In scope for later Artifact B2 implementation passes: additive artifact metadata,
relationship-aware reads and validation, archive and delete behavior, lineage
tracking, supersession flows, elicitation-aware nudges, sign-off recording, and
governance checks.

Out of scope for this bootstrap pass: changing runtime behavior, changing the
current artifact storage backend, adding a database, changing current MCP tool
names, or broadening Workbench beyond the artifact workflow layer.

## Stakeholders

The primary stakeholder is the local agent workflow using Workbench as a stdio
MCP coordination server. Secondary stakeholders are branch owners who need
self-contained epic packets, reviewers who need durable evidence and sign-off
records, and future runtime implementers who need a narrow contract that can be
validated against Markdown artifacts.

## Success Criteria

Artifact B2 succeeds when Workbench can describe and later implement advanced
artifact workflows without weakening the current kernel:

- Lifecycle state is explicit and reversible paths are distinguishable from
  destructive paths.
- Lineage and supersession make it clear which artifacts replace, derive from,
  or depend on one another.
- Human nudges, approvals, and sign-offs are traceable from the RFC hub.
- Frontmatter evolves additively so existing artifacts remain readable.
- Relationship and governance rules are specific enough to test before runtime
  work begins.

## Source References

- [README.md](../../README.md)
- [Epic Branch Workflow](../how-to/epic-branch-workflow.md)
- [Artifact Conventions](../reference/artifact-conventions.md)
- [Artifact B2 RFC](artifact-b2-rfc.md)

## Open Questions

Artifact-local open questions are intentionally routed through the
[Artifact B2 RFC](artifact-b2-rfc.md#human-in-the-loop-index) so the packet has
one living drill hub.
