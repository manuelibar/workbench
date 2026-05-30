---
id: "idea-mgmt-assumption"
type: "assumption"
title: "Idea Records Can Start as Local Durable Markdown"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Idea Records Can Start as Local Durable Markdown

## Statement

The first implementation can store idea records as local durable Markdown or
Markdown-adjacent files with deterministic metadata, without requiring a
database or remote service.

## Evidence

Current Workbench already uses file-backed Markdown artifacts as durable
contracts, and the epic workflow expects bootstrap packets to define contracts
before runtime implementation. MCP resources also fit read-oriented context
surfaces, which matches the need to expose idea context without making every
idea an executable action. Opposing evidence is that a high-volume idea store
may eventually need indexes, relation queries, or compaction that are more
efficient outside a flat-file layout.

## Validation Plan

Validate this assumption by implementing a narrow capture and promotion slice:
capture at least twenty idea records, list and retrieve them through resources,
relate several ideas to existing artifacts, promote at least one idea to a
typed artifact, and confirm that retrieval, validation, and backlinks remain
usable without a database. Revisit storage only if listing, relation lookup, or
promotion becomes slow or brittle in that slice.

## Source References

- `README.md`
- `docs/reference/artifact-conventions.md`
- `docs/artifacts/idea-mgmt-concept-map.md`
- `docs/artifacts/idea-mgmt-initial-implementation-plan.md`

## Open Questions

The storage path and edit policy should be revisited during the first
implementation slice if flat files create validation or relation-query limits.
