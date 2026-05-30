---
id: "knowledge-mgmt-initial-implementation-plan"
type: "implementation_plan"
title: "Knowledge Management Initial Implementation Plan"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Knowledge Management Initial Implementation Plan

## Objective

Implement the first local, source-governed knowledge management slice for
Workbench: register approved sources, index local docs and artifacts, retrieve
cited evidence, expose freshness and ranking metadata, and record retrieval
feedback without touching explicit memory or artifact authoring.

This plan is produced by [Knowledge Management RFC](knowledge-mgmt-rfc.md).

## Steps

1. Add a source registry model with source ID, connector type, root URI,
   authority, refresh policy, include/exclude patterns, access notes, status,
   last indexed timestamp, last source version, and content hash.
2. Implement local file and durable artifact connectors under the default
   `KM-NUDGE-001` source scope.
3. Add deterministic document extraction for Markdown and plain text, including
   title, canonical URI, MIME type, source ID, artifact ID when present, and
   section metadata when available.
4. Add stable chunking with content-derived chunk IDs and citation handles.
5. Build a lexical index with metadata filters, source priority, freshness
   state, and rank explanation fields.
6. Implement `knowledge.search` to return retrieval packets with query,
   candidates, snippets, citations, score details, freshness warnings, and
   no-evidence packets.
7. Implement `knowledge.source.list` and `knowledge.index.refresh`, including
   stale, missing, refreshed, and unchanged states.
8. Add `knowledge.feedback.record` with typed feedback events for usefulness,
   stale evidence, bad citation, bad ranking, missing source, and conflict.
9. Expose source and retrieval packet resources after the tool contracts are
   stable.
10. Add boundary tests proving search cannot write memory, refresh cannot edit
    artifacts, and feedback cannot become personalization.
11. Revisit embeddings or remote documentation connectors only after the local
    lexical baseline passes the test strategy.

## Verification

- Artifact validation passes for the packet files.
- Unit tests cover source registry parsing, include/exclude rules, stable chunk
  IDs, citation formatting, freshness transitions, and feedback event typing.
- Integration tests index local docs and artifacts, refresh unchanged and
  changed sources, search for known terms, and assert cited retrieval packets.
- Boundary tests assert no memory writes and no artifact edits during indexing,
  search, or feedback recording.
- Ranking tests show source priority, freshness, lexical score, and conflict
  handling in rank explanations.
- Manual MCP smoke tests can list sources, refresh the index, search, inspect a
  retrieval packet resource, and record feedback.

## Rollback

The first implementation should be isolated behind new knowledge tools,
resources, and storage paths. Rollback should disable the knowledge capability
registration and remove the new source/index storage without changing context
or artifact behavior.

If indexing corrupts or pollutes local state, delete the generated knowledge
index and rebuild from registered sources. Source files, durable artifacts, and
explicit memory stores must remain untouched by rollback.

## Source References

- [Knowledge Management RFC](knowledge-mgmt-rfc.md)
- [Knowledge Management Test Strategy](knowledge-mgmt-test-strategy.md)
- [Knowledge Management Concept Map](knowledge-mgmt-concept-map.md)

## Open Questions

Implementation defaults are governed by `KM-NUDGE-001` through `KM-NUDGE-006`
in the RFC human-in-the-loop index.
