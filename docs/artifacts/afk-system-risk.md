---
id: "afk-system-risk"
type: "risk"
title: "AFK Autonomy Escapes the Run Boundary"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# AFK Autonomy Escapes the Run Boundary

## Description

An AFK run could modify unintended files, spend more budget than expected, keep
running after useful progress has stopped, miss human review points, or produce
changes whose provenance cannot be reconstructed.

This risk is not limited to bugs in Workbench. It can also come from ambiguous
source tasks, stale branches, permissive executor configuration, unbounded
network access, missing verification commands, notification failure, or executor
output that looks complete but does not satisfy the requested acceptance
criteria.

## Impact

If the risk materializes, the user may return to a worktree containing
unreviewed or out-of-scope changes, hidden failures, exhausted budget, or
progress that cannot be safely resumed. The trust damage is larger than the code
damage: once AFK runs feel unpredictable, users will avoid delegation or will
require manual supervision that defeats the system's purpose.

## Likelihood

Medium before explicit safety gates exist. The likelihood is high for a naive
write-enabled implementation that simply launches an executor with broad access
and a long prompt. The likelihood drops to low only after Workbench enforces
bounded settings, scoped writable paths, clean-worktree preflight, iteration
budgets, signal parsing, verification gates, cancellation, and durable progress
artifacts.

Confidence is medium because Codex and similar executors already expose sandbox,
approval, MCP, notification, and configuration controls, but Workbench still has
to compose those controls into a local run contract.

## Mitigation

- Default v1 to dry-run or read-only AFK until `AFK-HITL-002` approves the
  minimum write-enabled safety gate.
- Require clean worktree and branch/worktree isolation before write-enabled
  launch.
- Snapshot run input and resolved settings at start.
- Require explicit writable scope and deny writes outside it.
- Cap wall-clock time, per-iteration time, iteration count, and cost or token
  budget.
- Parse only known signals: `continue`, `yield`, `done`, and `failed`.
- Override `done` when verification fails.
- Persist progress after every iteration, including command output summaries,
  verification results, last signal, and next action.
- Treat missing progress persistence as a run failure.
- Support cancellation as a first-class stop condition.
- Escalate to human review on yield timeout, safety gate failure, parse failure,
  or repeated verification failure.

## Owner

The AFK system epic owner owns the risk until implementation assigns more
specific owners for run settings, executor integration, safety gates, progress
artifacts, and notification adapters.

## Source References

- [AFK System RFC](afk-system-rfc.md)
- [AFK System Runbook](afk-system-runbook.md)
- [AFK System Test Strategy](afk-system-test-strategy.md)
- [Ripple bounded execution](https://github.com/manuelibar/ripple/blob/main/docs/manifesto/10-execution-models.md)
- [Codex config reference](https://developers.openai.com/codex/config-reference#configtoml)

## Open Questions

The mitigation set requires explicit human approval in `AFK-HITL-002` before
write-enabled AFK runtime work should proceed.
