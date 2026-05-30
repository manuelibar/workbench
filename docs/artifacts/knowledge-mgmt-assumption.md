---
id: "knowledge-mgmt-assumption"
type: "assumption"
title: "Knowledge MVP Can Start Local and Source-Governed"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Knowledge MVP Can Start Local and Source-Governed

## Statement

The first knowledge management implementation can start with a local,
source-governed registry, deterministic metadata, lexical retrieval, and
citation packaging before adding vector search, hosted storage, or remote
credentialed connectors.

This assumption keeps the epic compatible with the current Workbench shape:
local stdio server, one agent process, file-backed artifacts, and no required
database.

## Evidence

Workbench `main` already keeps context and artifacts local, and the current
artifact conventions favor deterministic contracts. MCP resources and tools are
sufficient to expose source inventories and bounded retrieval without requiring
a hosted search system. The GitHub Docs API demonstrates that authoritative
documentation sources can expose pagelists, article metadata, Markdown content,
search, and `llms.txt` discovery endpoints that a connector can consume without
inventing a broad crawler.

Current MCP documentation also separates application-driven resources from
model-controlled tools, which supports a small local design: source lists as
resources, retrieval and refresh as explicit tools, and feedback as a separate
recording path.

## Validation Plan

- Build the initial source registry against local project docs and durable
  artifacts before adding remote connectors.
- Prove stable chunk IDs and citation generation on `README.md`,
  `docs/reference/artifact-conventions.md`, and a small artifact fixture.
- Add one explicit docs API connector only after the local connector passes
  freshness, deletion, and citation tests.
- Keep embeddings optional until lexical retrieval plus metadata ranking has a
  measurable baseline.
- Revisit this assumption if local retrieval cannot produce acceptable cited
  results on the fixture corpus.

## Source References

- [Workbench README](../../README.md)
- [Artifact Conventions](../reference/artifact-conventions.md)
- [MCP Resources](https://modelcontextprotocol.io/specification/2025-11-25/server/resources)
- [GitHub Docs API](https://docs.github.com/get-started/using-github-docs/github-docs-api)

## Open Questions

The ranking tradeoff is tracked as `KM-NUDGE-002` in the RFC.
