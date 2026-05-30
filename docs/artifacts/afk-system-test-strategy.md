---
id: "afk-system-test-strategy"
type: "test_strategy"
title: "AFK System Test Strategy"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# AFK System Test Strategy

## Scope

This strategy covers verification for the future AFK runtime contract described
by the [AFK System RFC](afk-system-rfc.md). It focuses on run settings,
preflight safety gates, iteration signals, stop conditions, progress artifacts,
resume/cancel behavior, and executor integration boundaries.

It does not cover implementation of backlog storage, full mode mutation, mobile
notifications, distributed queues, or web dashboards unless later epics add those
surfaces to the AFK contract.

## Test Levels

### Contract tests

- Validate run settings require source ID, acceptance criteria, budgets,
  executor profile, writable scope, and verification policy.
- Validate run states and transitions reject impossible moves, such as
  `completed` without accepted `done`.
- Validate only known base signals are accepted by the default parser.
- Validate stop conditions override `continue` when budgets, timeouts, or safety
  gates are exhausted.
- Validate progress artifacts contain source references, settings snapshot,
  iteration summary, signal, verification result, and final status.

### Unit tests

- Signal parser handles structured output, missing output, unknown signal, and
  conflicting free-text claims.
- Budget checker handles wall-clock timeout, per-iteration timeout, iteration
  cap, and cost/token cap.
- Scope checker detects writes outside allowed paths.
- Worktree checker detects dirty state, branch mismatch, and changed files.
- Verification gate converts failed final verification into retry, yield, or
  failed according to policy.
- Human nudge collector records only indexed nudge IDs.

### Integration tests

- Read-only dry-run starts from a source artifact, runs one simulated executor
  iteration, and persists progress without changing unrelated files.
- Write-enabled simulated run modifies an allowed fixture path, passes
  verification, and completes.
- Verification-fail run receives `done`, fails the verification command, records
  the override, and loops or fails according to retry policy.
- Yield/resume run records a nudge, pauses, accepts resume input, and continues
  from the persisted state.
- Cancellation run stops dispatching, records worktree status, and terminates as
  `cancelled`.
- Restart inspection reconstructs current state from files after process restart.

### Manual tests

- Run the [AFK System Runbook](afk-system-runbook.md) against a docs-only task.
- Review generated progress artifacts for enough information to understand what
  ran, what changed, why it stopped, and what follow-up remains.
- Smoke-test Codex launch configuration without mutating user-level Codex config.
- Confirm a human can answer a yielded nudge and see the response attached to
  the resumed run.

## Fixtures

Required fixtures for the first implementation pass:

- A temporary git repository with a clean branch and a dirty-branch variant.
- A `docs/artifacts/` source artifact with acceptance criteria.
- A task-shaped artifact that mimics future backlog input.
- Allowed and denied writable directories.
- Fake executor outputs for `continue`, `yield`, `done`, `failed`, missing
  signal, malformed signal, and contradictory text.
- Verification commands that pass, fail, time out, and are missing.
- Run settings examples for read-only, dry-run, write-enabled, budget-exhausted,
  and yield-timeout cases.
- Progress artifact golden files.

## Risks

- A fake executor can miss real Codex behavior around approvals, sandboxing,
  command output, and signal formatting. Mitigate with a small manual Codex smoke
  suite before enabling writes.
- File-backed state can pass unit tests but still be hard to inspect after
  interrupted writes. Mitigate with restart and partial-write tests.
- Worktree checks can become platform-sensitive. Mitigate with git command
  fixtures and path normalization tests.
- Notification hooks can fail silently. Mitigate by treating progress artifacts
  as authoritative and notification as secondary.
- Human nudge coverage can drift. Mitigate by validating that every
  `AFK-HITL-*` reference appears in the RFC human-in-the-loop index.

## Source References

- [AFK System RFC](afk-system-rfc.md)
- [AFK System Runbook](afk-system-runbook.md)
- [AFK System Concept Map](afk-system-concept-map.md)
- [Artifact conventions](../reference/artifact-conventions.md)
- [Ripple worker README](https://github.com/manuelibar/ripple/blob/main/ripple-worker/README.md)
- [Codex CLI](https://developers.openai.com/codex/cli)

## Open Questions

The test strategy should be revised after `AFK-HITL-001` and `AFK-HITL-002`
resolve whether write-enabled AFK is part of v1.
