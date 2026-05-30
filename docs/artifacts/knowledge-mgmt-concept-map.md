---
id: "knowledge-mgmt-concept-map"
type: "spec"
title: "Knowledge Management Concept Map"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Knowledge Management Concept Map

## Context

This taxonomy defines the epic vocabulary before implementation. It keeps
knowledge management centered on source-governed retrieval and separates it
from explicit memory and durable artifacts.

Workbench currently has artifact resources and a `context` tool. Knowledge
management should build beside that kernel: it can expose registered sources
and retrieval results through MCP resources and tools, but it must not replace
artifact authoring or memory workflows.

## Design

| Concept | Definition | Owner | Notes |
|---|---|---|---|
| Knowledge | Retrieved evidence from registered sources, returned with provenance and citations. | Knowledge epic | Knowledge is only as authoritative as its source record and freshness state. |
| Explicit memory | User-approved durable facts, preferences, and working norms. | Memory epic | Knowledge feedback is not memory by default. |
| Durable artifact | Typed Markdown contract under `docs/artifacts/` with deterministic validation. | Artifact system | Artifacts may be indexed read-only, but edits stay with artifact tools. |
| Source registry | The catalog of approved source roots, connectors, authority, refresh policy, and access boundaries. | Knowledge epic | The registry is the trust root for retrieval. |
| Source snapshot | One observed state of a source, including URI, version/ref, content hash, fetched time, and parser metadata. | Knowledge epic | Snapshots make freshness and rebuilds auditable. |
| Connector | Adapter that discovers and reads a source class such as local files, Git repositories, docs APIs, or explicit URLs. | Knowledge epic | Connectors do not decide memory or artifact semantics. |
| Extracted document | Normalized source unit before chunking. | Knowledge epic | Keeps title, canonical URI, MIME type, license/access metadata, and body. |
| Chunk | Addressable segment of an extracted document. | Knowledge epic | Chunk IDs must be stable under unchanged content. |
| Citation | Pointer from an answer or result to a source, document, chunk, section, line range, or URL. | Knowledge epic | Results without citation should be treated as retrieval failures. |
| Index | Searchable representation of chunks and metadata. | Knowledge epic | May start lexical and add embeddings later behind the same contract. |
| Freshness policy | Rule for when a source is current, stale, missing, or requires refresh. | Knowledge epic | Results expose freshness, not just score. |
| Ranker | Deterministic policy that orders candidates with score, source priority, freshness, and feedback signals. | Knowledge epic | Ranking must be explainable enough for tests and debugging. |
| Feedback event | User or agent signal about retrieval quality, citation quality, stale data, or missing data. | Knowledge epic | Feedback improves evaluation and ranking, not memory. |
| Retrieval packet | Structured result set returned to the agent, including query, candidates, citations, ranking metadata, and warnings. | Knowledge epic | The packet is evidence; synthesis remains the agent's job. |
| Boundary violation | Any attempt to store inferred facts as knowledge, mutate memory, or edit artifacts through indexing. | Shared | Boundary tests should make violations visible. |

Boundary model:

- Knowledge can cite memory only if a future memory contract exposes memory as a
  source class. Until then, memory is not a knowledge source.
- Knowledge can index artifacts, but the citation must preserve artifact ID and
  section identity.
- Feedback can lower or raise ranking confidence, but it cannot become a
  remembered preference without explicit memory import.
- Source authority and freshness are ranking inputs, not hidden global truth.

## Interfaces

Future implementation surfaces should be evaluated against these interface
shapes:

- `knowledge.source.register`: add a source with connector type, root URI,
  authority, refresh policy, inclusion rules, and exclusion rules.
- `knowledge.source.list`: inspect registered sources, snapshot state, and
  stale/missing status.
- `knowledge.index.refresh`: refresh one source or all due sources and report
  document, chunk, and deletion counts.
- `knowledge.search`: return a retrieval packet with cited candidates,
  freshness warnings, rank explanations, and no-answer diagnostics.
- `knowledge.feedback.record`: record explicit quality feedback for a retrieval
  packet, result, source, or citation.
- `workbench:///knowledge/sources`: resource listing registered source state.
- `workbench:///knowledge/results/{query_id}`: resource for a persisted
  retrieval packet when a session needs to inspect evidence after a search.

Minimal source record fields:

- `id`
- `connector_type`
- `root_uri`
- `authority`
- `status`
- `refresh_policy`
- `include_patterns`
- `exclude_patterns`
- `last_indexed_at`
- `last_observed_version`
- `last_content_hash`
- `access_notes`

Minimal retrieval result fields:

- `result_id`
- `source_id`
- `document_id`
- `chunk_id`
- `canonical_uri`
- `citation`
- `score`
- `rank_explanation`
- `indexed_at`
- `source_version`
- `freshness`
- `snippet`

## Edge Cases

- A source is deleted or inaccessible after it has been indexed. The source
  should be marked missing, old chunks should not appear as fresh, and results
  should show the missing state if retained for audit.
- Two sources conflict. The ranker should not hide the conflict; it should
  return both with authority and freshness metadata.
- A durable artifact is updated after indexing. Artifact citation should expose
  the artifact ID and indexed timestamp so stale artifact snippets are visible.
- A source returns generated summaries without source URLs. Those summaries can
  be indexed only if the summary source itself is registered and cited.
- A user marks a result as useful. The feedback event can influence evaluation
  or ranking, but it must not become a user preference.
- A connector sees secrets or restricted files. Exclusion rules and access notes
  must prevent accidental indexing, and tests should include deny-listed paths.
- A retrieval query has no cited candidates. The tool should return an explicit
  no-evidence packet instead of an uncited answer.

## Test Plan

The taxonomy should be validated by later implementation tests:

- Unit tests for source record parsing, stable chunk IDs, citation formatting,
  freshness state transitions, and feedback classification.
- Integration tests over a small fixture corpus with local docs, durable
  artifacts, stale snapshots, conflicting sources, and denied paths.
- Ranking tests that assert source priority, freshness, lexical score, and
  feedback signals are explainable and deterministic.
- Boundary tests that prove search cannot write memory and indexing cannot edit
  artifacts.
- Manual MCP smoke tests that list sources, refresh an index, search, inspect a
  retrieval packet, and record feedback.

## Source References

- [MCP Resources](https://modelcontextprotocol.io/specification/2025-11-25/server/resources)
- [MCP Tools](https://modelcontextprotocol.io/specification/2025-11-25/server/tools)
- [GitHub Docs API](https://docs.github.com/get-started/using-github-docs/github-docs-api)
- [Knowledge Management RFC](knowledge-mgmt-rfc.md)

## Open Questions

The open taxonomy decisions are indexed in the RFC as `KM-NUDGE-001` through
`KM-NUDGE-006`.
