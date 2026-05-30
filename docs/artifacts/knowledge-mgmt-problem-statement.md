---
id: "knowledge-mgmt-problem-statement"
type: "problem_statement"
title: "Knowledge Management Problem Statement"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Knowledge Management Problem Statement

## Context

Workbench `main` is a local stdio MCP kernel with current context in memory and
typed Markdown artifacts under `docs/artifacts/`. It intentionally leaves
knowledge, memory, sessions, work management, and related systems to epic
branches that define contracts before runtime work.

The agent needs a way to retrieve current project knowledge without loading the
entire repository, external documentation, or prior artifacts into the prompt.
Current MCP primitives distinguish resources, which are application-driven
context, from tools, which a model can invoke. That distinction maps well to a
knowledge system that exposes source inventories as resources and retrieval as
bounded tools.

## Problem

Workbench does not yet have a source-governed knowledge layer. Without one, an
agent must rely on ad hoc file reads, external web search, stale model
knowledge, or explicit memory. Those mechanisms have different trust
properties, but the agent has no durable local contract for distinguishing:

- Knowledge: source material retrieved from registered sources with citations.
- Explicit memory: user-approved remembered facts, preferences, and working
  norms.
- Durable artifacts: typed Markdown contracts and decisions governed by the
  artifact system.

The missing contract makes it easy to cite the wrong thing, hide staleness,
over-trust low-authority sources, turn feedback into implicit memory, or let an
index become a shadow artifact store.

## Impact

The direct impact is poorer agent reliability:

- Answers may be plausible but uncited.
- Stale documentation may outrank current repository state.
- Project artifacts may be summarized as loose snippets instead of linked by
  artifact ID and section.
- Retrieval feedback may accidentally personalize future behavior without user
  approval.
- Index rebuilds may be nondeterministic or impossible to debug.

The broader impact is architectural: memory, artifacts, and knowledge all deal
with durable context, but they need separate ownership so later epics can evolve
without conflicting semantics.

## Constraints

- The bootstrap pass is docs-only and must not add runtime code.
- The first implementation should respect the current local stdio server shape
  and should not assume a hosted service or database.
- All knowledge must be traceable to source records with authority, freshness,
  and citation metadata.
- The system must be able to index durable artifacts as read-only sources while
  preserving artifact tools as the canonical edit path.
- The system must not infer or write explicit memory.
- External web or GitHub sources must be explicitly registered or approved; a
  broad crawler is outside the initial contract.
- Human nudges for open decisions are centralized in the RFC index.

## Source References

- [Workbench README](../../README.md)
- [Artifact Conventions](../reference/artifact-conventions.md)
- [MCP Resources](https://modelcontextprotocol.io/specification/2025-11-25/server/resources)
- [MCP Security Best Practices](https://modelcontextprotocol.io/docs/tutorials/security/security_best_practices)

## Open Questions

See the RFC human-in-the-loop index for source scope, ranking, freshness,
artifact indexing, feedback, and conflict-resolution nudges.
