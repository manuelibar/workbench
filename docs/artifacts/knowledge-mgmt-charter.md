---
id: "knowledge-mgmt-charter"
type: "charter"
title: "Knowledge Management Charter"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Knowledge Management Charter

## Mission

Define Workbench knowledge management as a source-governed retrieval system
that lets an agent find, rank, cite, refresh, and learn from explicit feedback
about source material without confusing that material with explicit memory or
durable artifacts.

Knowledge answers must be grounded in identifiable sources. A knowledge result
is not a remembered preference, not a private inferred fact about the user, and
not the canonical editing surface for an artifact. It is retrieved evidence with
provenance, ranking metadata, freshness metadata, and citation handles.

## Scope

In scope:

- A source registry for files, repositories, documentation sites, API docs, and
  explicitly approved web sources.
- Index lifecycle: discovery, extraction, chunking, metadata capture, hashing,
  refresh, deletion, stale marking, and rebuilds.
- Retrieval lifecycle: query normalization, candidate generation, ranking,
  reranking, citation packaging, result boundaries, and no-result behavior.
- Freshness policies that expose `indexed_at`, source version, content hash,
  source authority, and staleness in results.
- Feedback events that record whether retrieved evidence was useful, missing,
  stale, incorrectly ranked, or incorrectly cited.
- Boundary contracts with memory and artifact epics.

Out of scope for this epic:

- Explicit memory capture, user preference storage, and personalization rules.
- Artifact creation, artifact editing, artifact lineage, and artifact review
  workflows beyond indexing artifacts as read-only sources.
- A broad autonomous web crawler, private remote credential broker, or hosted
  multi-user search service.
- Model training or automatic promotion of retrieved facts into memory.

Boundary rules:

- Knowledge may index durable artifacts as sources, but artifact tools remain
  the canonical authoring and validation path.
- Knowledge may use explicit memory to shape a query only after the memory epic
  defines that contract; it must not create or mutate memory.
- Feedback on retrieval quality is evaluation data, not memory, unless a later
  memory workflow imports it through an explicit user-approved path.

## Stakeholders

- The local human operator, who approves source boundaries, freshness defaults,
  and trust decisions.
- The agent using Workbench over stdio, which needs bounded retrieval instead
  of loading all documents into the context window.
- The context and artifact kernel, which owns current context and typed
  Markdown artifacts.
- The memory epic, which owns explicit remembered preferences, identity facts,
  and personalization behavior.
- Future implementation passes that add retrieval, indexing, and feedback tools
  without changing this source-governed contract.

## Success Criteria

- Every retrieved claim can be traced to a registered source, source version or
  content hash, and citation span or section.
- Stale, missing, restricted, or conflicting sources are surfaced explicitly in
  retrieval results.
- Ranking behavior is explainable enough to debug with score components,
  source priority, freshness, and feedback signals.
- Retrieval feedback improves evaluation and ranking without silently becoming
  personal memory.
- Durable artifacts remain durable contracts and can be indexed without losing
  artifact identity, validation, or edit boundaries.
- The first implementation can run locally for one agent process and can be
  tested without a network dependency.

## Source References

- [Workbench README](../../README.md)
- [Epic Branch Workflow](../how-to/epic-branch-workflow.md)
- [Artifact Conventions](../reference/artifact-conventions.md)
- [MCP Resources](https://modelcontextprotocol.io/specification/2025-11-25/server/resources)
- [MCP Tools](https://modelcontextprotocol.io/specification/2025-11-25/server/tools)

## Open Questions

The human nudges for this packet are tracked in
[knowledge-mgmt-rfc.md](knowledge-mgmt-rfc.md#human-in-the-loop-index).
