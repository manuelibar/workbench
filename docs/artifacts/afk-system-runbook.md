---
id: "afk-system-runbook"
type: "runbook"
title: "AFK System Runbook"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# AFK System Runbook

## Scope

This runbook defines the intended operator procedure for a bounded AFK run in
Workbench. It applies to future read-only, dry-run, and write-enabled runs that
start from a source artifact, backlog-shaped task, or explicit run input.

During the docs-only phase, use it as a manual checklist. During implementation,
turn each step into validation, state transitions, progress artifact updates, or
operator-facing guidance.

This action artifact links back to the [AFK System RFC](afk-system-rfc.md).

## Prerequisites

- The source input has a stable ID, title, context references, acceptance
  criteria, and requested verification.
- The operator has selected a run profile with bounded settings:
  max iterations, run timeout, per-iteration timeout, cost or token budget,
  executor profile, approval policy, sandbox/network policy, writable scope, and
  progress cadence.
- The worktree and branch are known.
- The worktree is clean unless the run is read-only or the dirty files are
  explicitly recorded as pre-existing context.
- The executor is available and authenticated if the run will call it.
- The verification command is available or the run settings document why no
  command applies.
- For write-enabled runs, `AFK-HITL-002` has been resolved and the safety gate is
  enforced.

## Procedure

### 1. Create run draft

Record:

- Run ID and title.
- Source kind and source ID.
- Acceptance criteria.
- Context references.
- Requested writable scope.
- Requested verification command.
- Operator and timestamp.

### 2. Resolve settings

Snapshot the run settings before launch. Include:

- Executor profile and model/profile name.
- Approval policy.
- Sandbox and network policy.
- Writable roots and denied paths.
- Max iterations.
- Per-iteration timeout.
- Run timeout.
- Cost or token budget.
- Retry policy.
- Yield timeout.
- Notification/status surface.

### 3. Run preflight

Fail preflight if any required item is missing. For write-enabled runs, also
fail if:

- The branch or worktree is not isolated for this run.
- The worktree contains unrelated changes.
- Writable scope is absent or broader than the source task needs.
- Verification command is absent without a recorded exception.
- The executor profile would allow unbounded approvals, full filesystem access,
  or broad network access without explicit policy.
- Source references do not resolve.

Persist the preflight result before starting execution.

### 4. Execute one iteration

Prepare context from source input, prior progress, current worktree status, and
the run profile. Dispatch the executor for one iteration. Capture:

- Start and end time.
- Tool or command summary.
- Changed files summary.
- Executor result.
- Parsed signal.
- Token, cost, or duration metrics when available.

### 5. Wrap up the iteration

Parse the executor signal:

- `continue`: run verification gates that apply to ongoing work, persist
  progress, and dispatch the next iteration if budgets allow.
- `yield`: persist the human nudge, set state to `yielded`, start yield timeout,
  and wait for resume input.
- `done`: run required verification, persist final evidence, and complete only if
  gates pass.
- `failed`: retry if policy allows, otherwise persist failure evidence and stop.

If the signal is missing or unknown, treat the iteration as failed with a parser
diagnostic.

### 6. Resume, cancel, or finish

For resume:

- Record human input.
- Re-run any stale preflight checks.
- Dispatch the next iteration with the resume payload.

For cancellation:

- Stop dispatching new iterations.
- Collect current worktree status and progress evidence.
- Mark state `cancelled` after wrap-up completes.

For completion:

- Record final status.
- Record verification evidence.
- Record generated artifacts or commits.
- Record follow-up work.

## Verification

A run is successful only when:

- The final state is `completed`.
- The source input and resolved settings are persisted.
- Every iteration has a progress entry.
- The accepted final signal is `done`.
- Required verification passed after the final signal.
- Changed files, commits, or generated artifacts are listed.
- Any human nudges are linked to the RFC human-in-the-loop index.
- The worktree status is recorded at completion.

For read-only or dry-run runs, success means the run produced useful progress
evidence and did not modify files outside allowed progress artifacts.

## Escalation

Escalate to the human operator when:

- The run emits `yield`.
- Yield timeout expires.
- Preflight fails.
- Safety gates fail.
- Verification fails repeatedly.
- The executor cannot be launched or authenticated.
- The executor emits no parseable signal.
- Budgets are exhausted before acceptance criteria are met.
- The worktree changes outside the allowed scope.
- Notification or progress persistence fails.

Escalation output must include the run ID, source ID, current state, last signal,
blocking condition, recommended next action, and links to progress evidence.

## Source References

- [AFK System RFC](afk-system-rfc.md)
- [AFK System Concept Map](afk-system-concept-map.md)
- [AFK System Risk](afk-system-risk.md)
- [Ripple execution models](https://github.com/manuelibar/ripple/blob/main/docs/manifesto/10-execution-models.md)
- [Codex config reference](https://developers.openai.com/codex/config-reference#configtoml)

## Open Questions

The runbook defaults to read-only/dry-run behavior until `AFK-HITL-001` and
`AFK-HITL-002` are resolved.
