---
id: "artifact-b2-test-strategy"
type: "test_strategy"
title: "Artifact B2 Test Strategy"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Artifact B2 Test Strategy

## Scope

This strategy covers future Artifact B2 parser, validator, list/read, mutation,
relationship, elicitation, and sign-off behavior. It does not require runtime
changes in the bootstrap pass. The first implementation pass should prove that
B2 metadata is additive and that existing artifacts keep validating.

## Test Levels

Unit tests should cover frontmatter parsing, lifecycle enum validation,
relationship edge validation, sign-off record validation, and transition rules.

Integration tests should exercise artifact store operations against fixture
directories containing active, archived, superseded, tombstoned, and legacy
artifacts.

System-level MCP tests should verify capability behavior after B2 tools exist:
tool discovery, resource reads, list filtering, error reporting, and human
confirmation paths for governed operations.

Manual review should inspect generated artifact Markdown to confirm that
lineage, successor links, sign-off evidence, and human nudges are readable
without special tooling.

## Fixtures

The fixture set should include:

- A legacy artifact with only the current required frontmatter.
- An archived artifact with a reason and timestamp.
- A superseded artifact with a valid successor and a successor pointing back.
- A tombstoned artifact that retains enough metadata for audit.
- A relationship-rich artifact using every approved relationship type.
- A sign-off-gated transition with an approval record.
- Invalid examples for unknown lifecycle state, missing successor, relationship
  cycle, hard delete without approval, and unanswered required human nudge.

## Risks

The main test risk is validating syntax without validating workflow meaning.
Tests must assert transition rules and relationship consistency, not only parse
success. Another risk is overfitting to proposed field names before `B2-D2` is
resolved; fixtures should be easy to rename while preserving the behavioral
cases. MCP elicitation tests should stay separate until client capability and
sensitive-data handling are explicit.

## Source References

- [Artifact B2 RFC](artifact-b2-rfc.md)
- [Artifact B2 Initial Implementation Plan](artifact-b2-initial-implementation-plan.md)
- [Artifact B2 Risk](artifact-b2-risk.md)
- [MCP Elicitation, 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25/client/elicitation)
- [MCP Tools, 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25/server/tools)

## Open Questions

The fixture names and expected transition checks should be updated after
`B2-D1`, `B2-D2`, and `B2-Q1` are resolved in the
[Artifact B2 RFC](artifact-b2-rfc.md#human-in-the-loop-index).
