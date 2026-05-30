---
id: "knowledge-mgmt-test-strategy"
type: "test_strategy"
title: "Knowledge Management Test Strategy"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Knowledge Management Test Strategy

## Scope

This strategy covers the first knowledge management implementation proposed by
[Knowledge Management RFC](knowledge-mgmt-rfc.md): source registry, local and
artifact connectors, extraction, chunking, freshness, ranking, citations,
retrieval packets, feedback events, and boundaries with explicit memory and
durable artifacts.

It does not cover broad web crawling, hosted multi-user search, remote private
auth connectors, or embedding quality beyond optional future evaluation.

## Test Levels

Contract tests:

- Validate every knowledge artifact frontmatter and required section.
- Validate source records reject missing IDs, unknown connector types, empty
  roots, unsafe include/exclude patterns, and invalid refresh policies.
- Validate retrieval packets require citations for returned evidence.

Unit tests:

- Source registry normalization and source status transitions.
- Include/exclude matching for files and artifact IDs.
- Markdown extraction for title, sections, links, and line or section anchors.
- Stable chunk IDs for unchanged content and changed IDs for modified content.
- Freshness classification for fresh, stale, missing, denied, and failed
  sources.
- Rank explanation assembly from source priority, freshness, lexical score, and
  feedback signals.
- Feedback event typing and validation.

Integration tests:

- Index a fixture corpus with local docs, durable artifacts, conflicting
  sources, denied paths, and stale snapshots.
- Search for known terms and assert result order, citation handles, freshness
  warnings, and no-evidence behavior.
- Refresh unchanged, changed, deleted, and newly added documents.
- Index artifacts read-only and prove artifact content is unchanged.
- Record feedback for a retrieval packet and prove it affects only feedback
  state or evaluation inputs.

Boundary and security tests:

- Assert knowledge tools cannot write explicit memory.
- Assert index refresh cannot call artifact update paths.
- Assert denied files and restricted patterns do not produce chunks.
- Assert stale or missing source states are visible in result warnings.
- Assert uncited generated summaries are rejected as evidence unless the summary
  source itself is registered and cited.

Manual MCP smoke tests:

- List registered sources as a resource.
- Refresh a source and inspect counts.
- Search a known query and inspect a retrieval packet.
- Record feedback on a result.
- Confirm capability changes do not break existing artifact tools.

## Fixtures

- `README.md`, `docs/how-to/epic-branch-workflow.md`, and
  `docs/reference/artifact-conventions.md` as current local project docs.
- A small synthetic artifact set with one charter, one RFC, and one updated
  artifact section to test artifact citations and stale artifact index state.
- A conflicting-source fixture where local docs and an artifact disagree.
- A denied-path fixture such as `.env`, private token-looking files, and
  excluded directories.
- A mock docs API fixture shaped like GitHub Docs API article, metadata,
  pagelist, search, and `llms.txt` responses.
- Feedback fixtures for useful result, stale result, bad citation, missing
  source, and ranking conflict.

## Risks

- Ranking tests can become brittle if they assert exact scores instead of
  stable ordering and rank explanation components.
- External documentation changes can make integration tests flaky; remote
  connectors should use recorded fixtures for deterministic tests.
- Embedding support can hide source and citation bugs behind better recall; it
  should remain optional until lexical retrieval has a baseline.
- Feedback can accidentally become personalization if tests only check retrieval
  output and not memory writes.
- Artifact indexing can bypass artifact contracts if tests do not assert that
  artifact files are unchanged.

## Source References

- [Knowledge Management RFC](knowledge-mgmt-rfc.md)
- [Initial Implementation Plan](knowledge-mgmt-initial-implementation-plan.md)
- [MCP Tools](https://modelcontextprotocol.io/specification/2025-11-25/server/tools)
- [MCP Security Best Practices](https://modelcontextprotocol.io/docs/tutorials/security/security_best_practices)

## Open Questions

Test defaults follow the RFC human-in-the-loop defaults until the indexed
nudges are answered.
