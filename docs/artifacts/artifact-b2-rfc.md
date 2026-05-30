---
id: "artifact-b2-rfc"
type: "rfc"
title: "Artifact B2 Advanced Workflow Layer RFC"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Artifact B2 Advanced Workflow Layer RFC

## Summary

Artifact B2 is the living drill hub for Workbench's second-stage artifact
workflow layer. It defines the target contract for advanced lifecycle state,
delete/archive behavior, lineage, supersession, elicitation, sign-off,
frontmatter evolution, artifact relationships, and governance while keeping
this pass docs-only.

Action artifacts created from this RFC:

- [Artifact B2 Initial Implementation Plan](artifact-b2-initial-implementation-plan.md)
- [Artifact B2 Test Strategy](artifact-b2-test-strategy.md)

## Problem

The current artifact kernel can create, read, list, validate, guide, and update
typed Markdown artifacts, but it does not distinguish active work from archived,
deleted, superseded, blocked, or signed-off work. It also does not define typed
relationships beyond ordinary links. Without a B2 contract, future runtime work
could introduce destructive actions or relationship metadata inconsistently.

## Proposal

Define B2 as an additive workflow layer over the existing Markdown artifact
contract:

- Preserve the existing required frontmatter and section contracts.
- Add lifecycle vocabulary for active, archived, superseded, and delete-like
  states after `B2-D1` is resolved.
- Treat archive as the default non-destructive removal path and hard delete as
  exceptional governance-gated behavior.
- Represent lineage and supersession with stable artifact IDs and directional
  relationships.
- Decide whether future packets may live in semantic subdirectories while
  preserving artifact IDs as path-independent durable identity.
- Add an explicit relationship vocabulary before adding broad relationship
  search or graph behavior.
- Represent human nudges in the RFC Human-in-the-loop Index first; later map
  eligible nudges to MCP elicitation when client support and sensitive-data
  rules are clear.
- Record sign-off as structured governance evidence tied to a transition,
  artifact ID, actor or role, timestamp, and decision.
- Evolve frontmatter additively so current artifacts remain valid and readable.

The first implementation pass after this packet should create fixtures and
validators before adding mutation tools. Runtime behavior should change only
after the packet's lifecycle, sign-off, and relationship decisions are settled.

## Tradeoffs

Using Markdown as the source of truth keeps the workflow local, reviewable, and
consistent with the current kernel, but rich relationship queries may need a
derived index later. Additive frontmatter protects current artifacts, but it may
leave some governance data in body tables until the parser and validator evolve.

Deferring MCP elicitation avoids assuming client support too early, but it
means early B2 workflows rely on artifact-indexed nudges rather than interactive
server requests. Requiring sign-off for destructive transitions adds friction,
but it protects the audit trail and matches MCP's emphasis on user consent and
clear authorization for sensitive operations.

## Rollout

Rollout should proceed in gated docs-to-runtime stages:

1. Bootstrap this packet and commit it without runtime changes.
2. Resolve or accept defaults for the indexed human nudges.
3. Add Markdown fixtures for lifecycle, relationship, supersession,
   elicitation, and sign-off cases.
4. Add parser and validator support for additive metadata while preserving
   existing artifact validity.
5. Add read/list filtering semantics for archived and superseded artifacts.
6. Add mutation tools only after governance checks and tests exist.
7. Add MCP elicitation integration only when client capability detection and
   sensitive-data handling are explicit.

## Open Questions

### Human-in-the-loop Index

| ID | Nudge | Type | Why it matters | Blocks | Default if unanswered |
|---|---|---|---|---|---|
| B2-D1 | Choose the lifecycle state vocabulary and decide whether delete is tombstone-first or hard-delete-capable. | decision | State names become durable frontmatter and validation values. | Lifecycle fixtures, archive and delete validators, mutation tool design. | Use `active`, `archived`, `superseded`, and `tombstoned`; require separate approval for physical removal. |
| B2-D2 | Approve the first additive frontmatter fields for lifecycle, lineage, sign-off, and relationships. | decision | Field names are expensive to rename once artifacts exist. | Parser changes, migration guidance, validation fixtures. | Add optional fields only: `lifecycle`, `supersedes`, `superseded_by`, `related_artifacts`, `signoffs`, `workflow_reason`. |
| B2-Q1 | Identify which transitions require human sign-off. | question | Governance needs a clear boundary between routine cleanup and currentness-changing operations. | Sign-off schema, archive and delete behavior, supersession workflow. | Require sign-off for hard delete, tombstone, and supersession; allow ordinary archive with a reason. |
| B2-T1 | Decide when to use artifact-indexed human nudges versus MCP elicitation. | tradeoff | Runtime elicitation depends on client support and has sensitive-data constraints. | Elicitation UX, tool design, human-nudge validation. | Keep packet nudges in artifacts first; add MCP elicitation only for supported non-sensitive form requests or URL-mode flows. |
| B2-C1 | Challenge whether typed artifact relationships should ship before a derived relationship index exists. | challenge | Relationships without query support may still help reviewers but can drift without validation. | Relationship rollout, list/read behavior, test fixture scope. | Store explicit typed edges in Markdown first and defer graph indexing. |
| B2-Q2 | Decide whether artifact packets may use semantic subdirectories as packet prefixes. | question | Nested packet directories would reduce long repeated filename prefixes, but the current kernel lists flat Markdown files only. | Artifact discovery, path-to-ID rules, packet grouping, and migration from flat files. | Keep files flat until B2 adds recursive discovery; allow directory grouping only when artifact ID remains independent from the path. |

## Source References

- [README.md](../../README.md)
- [Epic Branch Workflow](../how-to/epic-branch-workflow.md)
- [Artifact Conventions](../reference/artifact-conventions.md)
- [Artifact B2 Research Note](artifact-b2-research-note.md)
- [MCP Specification, 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25)
- [MCP Elicitation, 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25/client/elicitation)
- [MCP Resources, 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25/server/resources)
- [MCP Tools, 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25/server/tools)
