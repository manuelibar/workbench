---
id: "knowledge-mgmt-rfc"
type: "rfc"
title: "Knowledge Management RFC"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Knowledge Management RFC

## Summary

This RFC is the living drill hub for the Workbench knowledge management epic.
It proposes a source-governed retrieval layer that owns registered sources,
indexing, freshness, ranking, citations, feedback, and boundaries with explicit
memory and durable artifacts.

The packet action artifacts are:

- [Initial Implementation Plan](knowledge-mgmt-initial-implementation-plan.md)
- [Test Strategy](knowledge-mgmt-test-strategy.md)

## Problem

Workbench needs current, inspectable project knowledge without relying on stale
model knowledge, unbounded prompt stuffing, ad hoc web search, or implicit
memory. The missing contract causes three classes of ambiguity:

- Source ambiguity: the agent cannot consistently explain where retrieved
  claims came from, how fresh they are, or why one source outranked another.
- Ownership ambiguity: knowledge, explicit memory, and durable artifacts all
  preserve context, but they have different truth and edit semantics.
- Feedback ambiguity: retrieval-quality feedback can become hidden
  personalization unless it is explicitly separated from memory.

## Proposal

Define knowledge management around these primitives:

- Source registry: explicit source roots, connector type, authority, refresh
  policy, inclusion/exclusion rules, access notes, last observed version, and
  content hash.
- Source snapshot: an auditable observed state of source content.
- Extracted document: normalized document with metadata and canonical URI.
- Chunk: stable addressable segment used for indexing and citations.
- Index: searchable representation of chunks and metadata.
- Retrieval packet: structured search result set with citations, freshness,
  rank explanations, warnings, and no-evidence behavior.
- Feedback event: typed evaluation signal about usefulness, ranking, freshness,
  missing evidence, or citation quality.

Boundary rules:

- Knowledge does not write explicit memory.
- Knowledge does not edit durable artifacts.
- Durable artifacts may be indexed read-only and cited by artifact ID and
  section.
- Feedback does not become memory unless a later memory workflow explicitly
  imports it with user approval.
- Results without citations are failures, not answers.

Packet map:

- [Charter](knowledge-mgmt-charter.md)
- [Problem Statement](knowledge-mgmt-problem-statement.md)
- [Concept Map](knowledge-mgmt-concept-map.md)
- [Assumption](knowledge-mgmt-assumption.md)
- [Risk](knowledge-mgmt-risk.md)
- [Research Note](knowledge-mgmt-research-note.md)
- [Initial Implementation Plan](knowledge-mgmt-initial-implementation-plan.md)
- [Test Strategy](knowledge-mgmt-test-strategy.md)

Likely future MCP surface:

- `knowledge.source.register`
- `knowledge.source.list`
- `knowledge.index.refresh`
- `knowledge.search`
- `knowledge.feedback.record`
- `workbench:///knowledge/sources`
- `workbench:///knowledge/results/{query_id}`

## Tradeoffs

Local-first MVP versus database-backed search:

Local-first keeps the implementation aligned with the current stdio kernel and
easy to test. It may limit scale and ranking quality until the index store
evolves.

Lexical-first retrieval versus immediate embeddings:

Lexical retrieval is deterministic and cheap for the first contract. Embeddings
can improve recall, but they add model dependencies, storage decisions, and
evaluation complexity. The proposed contract keeps embeddings optional behind
the same citation and source model.

Knowledge tools versus resources:

Source inventories and retrieval packets fit MCP resources because clients can
inspect them as context. Refresh, search, and feedback are actions and fit MCP
tools. This split follows the MCP resource/tool interaction model but creates
two surfaces to test.

Feedback as evaluation versus memory:

Keeping feedback outside memory prevents hidden personalization. It also means
useful user signals do not automatically improve personal workflow unless a
later memory integration imports them deliberately.

## Rollout

1. Land this docs-only packet as the source contract.
2. Implement source registry and artifact/local-file connector contracts.
3. Add deterministic extraction, chunking, metadata, and citation generation.
4. Add lexical retrieval with source priority, freshness, and rank explanation.
5. Add retrieval packet resources and search/refresh tools.
6. Add feedback event recording and evaluation fixtures.
7. Add optional remote documentation connector only after local source and
   citation tests pass.
8. Revisit embedding or reranking support after the lexical baseline has
   measurable fixture results.

## Open Questions

### Human-in-the-loop Index

| ID | Nudge | Type | Why it matters | Blocks | Default if unanswered |
|---|---|---|---|---|---|
| KM-NUDGE-001 | Decide which source classes are allowed in the MVP. | decision | Source scope determines connector design, access controls, and test fixtures. | Source registry and connector implementation. | Local repository files, `docs/artifacts/`, and explicit public docs URLs only; no broad crawl and no private remote auth. |
| KM-NUDGE-002 | Choose lexical-first retrieval or hybrid retrieval with embeddings for the first pass. | tradeoff | The choice affects index storage, determinism, dependencies, ranking tests, and offline operation. | Index schema and ranking implementation. | Start lexical plus metadata ranking; keep an optional embedding slot out of the first required path. |
| KM-NUDGE-003 | Decide how stale sources should behave when a search has otherwise relevant evidence. | decision | Freshness policy affects trust, result shape, and whether agents can use old evidence with warnings. | Freshness result semantics and UI/tool warnings. | Return stale evidence with `stale=true`, age metadata, and warning; fail closed only when required citations are missing. |
| KM-NUDGE-004 | Challenge the boundary that retrieval feedback is not explicit memory. | challenge | If this boundary is wrong, feedback storage and memory integration need a different contract. | Feedback schema and memory boundary tests. | Feedback remains retrieval evaluation data and never personalizes future sessions unless the memory epic imports it explicitly. |
| KM-NUDGE-005 | Approve indexing durable artifacts as read-only knowledge sources while artifact tools stay canonical. | approval | Artifact indexing improves discovery but can blur authoring and validation boundaries. | Artifact connector and citation format. | Index artifacts read-only with artifact ID and section citations; all edits stay with artifact tools. |
| KM-NUDGE-006 | Decide how to rank or expose conflicting source authority. | question | Conflicts between current code, docs, and artifacts are common and should not be hidden by a single score. | Rank explanation and conflict result shape. | Apply explicit source priority first, freshness second, score third, and surface conflicts instead of suppressing them. |

## Source References

- [Knowledge Management Research Note](knowledge-mgmt-research-note.md)
- [Knowledge Management Concept Map](knowledge-mgmt-concept-map.md)
- [Epic Branch Workflow](../how-to/epic-branch-workflow.md)
- [MCP Resources](https://modelcontextprotocol.io/specification/2025-11-25/server/resources)
- [MCP Tools](https://modelcontextprotocol.io/specification/2025-11-25/server/tools)
- [MCP Elicitation](https://modelcontextprotocol.io/specification/2025-11-25/client/elicitation)
