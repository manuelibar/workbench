---
id: "artifact-b2-risk"
type: "risk"
title: "Artifact B2 Workflow Semantics Can Become Destructive or Ambiguous"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Artifact B2 Workflow Semantics Can Become Destructive or Ambiguous

## Description

Artifact B2 introduces lifecycle transitions that can hide, replace, or remove
artifact content. If delete/archive behavior, supersession, lineage, and
sign-off are underspecified, agents may treat a recoverable archive as a
destructive delete, overwrite useful history, or mark an artifact as superseded
without a clear replacement.

## Impact

The main impact is loss of reviewer trust. Ambiguous lifecycle state can cause
agents to skip current evidence, cite stale work, or remove context that should
remain auditable. A destructive delete path without explicit approval can also
make it impossible to reconstruct why an epic chose one plan over another.

## Likelihood

Medium. The current kernel is intentionally small and does not yet encode these
workflow distinctions. The risk is manageable if B2 requires additive metadata,
transition reasons, relationship checks, and sign-off gates before runtime
mutations are implemented.

## Mitigation

Mitigate the risk by defining archive as the default non-destructive removal
path, treating hard delete as exceptional, requiring explicit successor links
for supersession, and recording sign-off for destructive or currentness-changing
operations. The implementation plan should add validation before adding runtime
mutation tools.

## Owner

The Artifact B2 branch owner owns this risk until the RFC is accepted. Runtime
owners inherit it when implementing lifecycle mutation tools or validators.

## Source References

- [Artifact B2 RFC](artifact-b2-rfc.md)
- [Artifact B2 Concept Map](artifact-b2-concept-map.md)
- [MCP Security and Trust & Safety, 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25#security-and-trust--safety)
- [MCP Tools Security Considerations, 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25/server/tools#security-considerations)

## Open Questions

The risk is reduced by resolving `B2-D1`, `B2-Q1`, and `B2-T1` in the
[Artifact B2 RFC](artifact-b2-rfc.md#human-in-the-loop-index).
