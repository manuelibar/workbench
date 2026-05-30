---
id: "memory-mgmt-test-strategy"
type: "test_strategy"
title: "Memory Management Test Strategy"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Memory Management Test Strategy

## Scope

This strategy covers the first runtime implementation of Workbench durable
memory described by [memory-mgmt-rfc.md](memory-mgmt-rfc.md). It focuses on
contract behavior rather than retrieval quality: explicit writes, scoped
recall, correction, forgetting, metadata completeness, sensitivity handling,
and precedence against artifacts, knowledge, session context, and current user
requests.

Out of scope for the first strategy are embedding model quality, remote storage
availability, ranking optimization, and UI-specific flows outside MCP tool and
resource behavior.

## Test Levels

Unit tests:

- Validate memory record schema and required metadata.
- Reject missing source, confidence, scope, state, timestamps, and sensitivity.
- Reject or require confirmation for sensitive categories according to policy.
- Verify scope matching and non-leakage across user, workspace, project,
  artifact, and session labels.
- Verify correction supersession and recall preference for active corrections.
- Verify forgotten records are excluded from normal recall.

Integration tests:

- Call `memory.remember`, `memory.recall`, `memory.correct`, and
  `memory.forget` through the MCP server surface.
- Inspect memory through resources or resource templates.
- Confirm capability lists and sync behavior remain deterministic when memory
  capabilities are added.
- Confirm malformed requests produce structured errors that agents can recover
  from.

System behavior tests:

- Current user instruction contradicts memory; current user instruction wins.
- Selected artifact contradicts memory during artifact-centered work; selected
  artifact wins and conflict is visible.
- Sourced knowledge contradicts memory for a factual claim; cited knowledge is
  preferred and the memory conflict is visible.
- Session note exists but is not durable; recall after session boundary does
  not return it unless explicitly remembered.
- Correction supersedes stale memory; recall returns the correction and not the
  stale active value.

Manual review:

- Inspect sample exported memory for readable provenance and no secret leakage.
- Exercise broad forget requests and confirm the user sees matched targets
  before destructive action.

## Fixtures

- Memory records for user preference, project fact, workflow hint, correction,
  negative memory, sensitive rejected record, and forgotten record.
- Artifact fixture that intentionally conflicts with a project-scoped memory.
- Knowledge fixture with a cited current fact that conflicts with an older
  memory.
- Session fixture that should be visible during a run but absent after a
  simulated boundary.
- Secret-like inputs such as API key and payment credential patterns to verify
  rejection.
- Ambiguous correction and forget queries that match multiple records.

## Risks

- Tests may overfit to file-backed storage and miss bugs in future retrieval
  backends. Keep a store interface test suite that any backend must pass.
- Ranking quality is hard to prove with deterministic tests. The first suite
  should verify filters, metadata, and precedence, then leave ranking metrics
  for a later retrieval-specific strategy.
- Forgetting tests can accidentally preserve content in logs or snapshots.
  Test fixtures should use fake non-sensitive content and assert logs do not
  include rejected secrets.
- User-consent behavior may differ by client UI. MCP-level tests should focus
  on explicit tool inputs and structured responses, while UI flows can be
  covered later.

## Source References

- [memory-mgmt-rfc.md](memory-mgmt-rfc.md)
- [memory-mgmt-risk.md](memory-mgmt-risk.md)
- [memory-mgmt-concept-map.md](memory-mgmt-concept-map.md)
- [memory-mgmt-initial-implementation-plan.md](memory-mgmt-initial-implementation-plan.md)

## Open Questions

No test-strategy-specific nudges are open. Packet-level human nudges are
indexed in [memory-mgmt-rfc.md](memory-mgmt-rfc.md).
