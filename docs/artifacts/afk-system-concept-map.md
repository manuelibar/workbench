---
id: "afk-system-concept-map"
type: "spec"
title: "AFK System Concept Map"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# AFK System Concept Map

## Context

This artifact is a taxonomy, not a runtime design. It names the core AFK
concepts so later implementation sessions can use stable language while staying
inside the current Workbench artifact-kernel architecture.

The concept map adapts Ripple's blueprint, signal, mode, run state, and bounded
execution vocabulary to Workbench. It does not require Ripple's Core service,
NATS streams, Postgres tables, or old worker implementation.

## Design

### Core nouns

| Concept | Meaning | Workbench v1 stance |
|---|---|---|
| AFK run | One bounded autonomous job delegated to an agent executor. | Reimplement as a Workbench-owned run record backed by artifacts or a small file ledger. |
| Run input | The artifact, backlog item, or explicit prompt that starts the run. | Adapt from typed artifacts first; add backlog adapter later. |
| Run settings | The resolved limits and permissions for one run. | Adapt Ripple blueprint fields into a smaller Workbench run profile. |
| Iteration | One agent dispatch and wrap-up cycle. | Reuse Ripple's prepare, execute, wrap-up shape conceptually. |
| Executor | Codex, Claude Code, Goose, or another runner invoked by Workbench. | Keep executor-neutral; Codex is the first practical target. |
| Signal | Agent outcome that routes the run. | Reuse four base signals: `continue`, `yield`, `done`, `failed`. |
| Stop condition | Rule that forces the run to end or pause. | Make stop conditions explicit in every run setting. |
| Progress artifact | Durable Markdown evidence of what happened. | Reimplement as typed artifacts or sections linked from the run. |
| Safety gate | Preflight, per-iteration, or wrap-up check that can block continuation. | Required before write-enabled AFK. |
| Human nudge | Indexed prompt for human input, approval, decision, challenge, or tradeoff. | Must be mirrored in the RFC human-in-the-loop index. |

### Run states

| State | Meaning | Terminal |
|---|---|---|
| `draft` | Run settings exist but have not passed preflight. | No |
| `ready` | Preflight passed and the run can start. | No |
| `running` | An iteration is active or dispatchable. | No |
| `yielded` | The run paused for human input or external input. | No |
| `cancelling` | Cancellation requested and wrap-up is collecting evidence. | No |
| `completed` | Success path reached and verification evidence is recorded. | Yes |
| `failed` | Failure path reached or retry/budget policy exhausted. | Yes |
| `cancelled` | User or system cancellation completed. | Yes |

### Signals

| Signal | Agent meaning | Default handling |
|---|---|---|
| `continue` | Progress was made and more work remains. | Dispatch another iteration if budgets and safety gates pass. |
| `yield` | External input is needed. | Pause as `yielded`, record a human nudge, and start yield timeout. |
| `done` | The assigned work is complete. | Run verification and transition to `completed` if gates pass. |
| `failed` | The agent cannot proceed. | Retry when policy allows, otherwise transition to `failed`. |

### Stop conditions

Every AFK run should stop or pause when any of these occurs:

- `done` accepted after verification.
- `failed` accepted after retry policy is exhausted.
- User cancellation requested.
- Wall-clock timeout reached.
- Iteration limit reached.
- Cost or token budget reached.
- Yield timeout reached.
- Worktree is dirty outside the allowed scope.
- Branch or worktree no longer matches the run settings.
- Required verification command fails.
- Safety gate detects secrets, destructive commands, unsupported permissions, or
  missing source references.
- Executor output cannot be parsed into an accepted signal.

### Settings taxonomy

| Setting group | Examples | Required for write-enabled v1 |
|---|---|---|
| Identity | run id, title, source artifact, operator | Yes |
| Scope | repository, branch, worktree, writable paths, read-only paths | Yes |
| Executor | executor name, model/profile, approval policy, sandbox mode | Yes |
| Budgets | max iterations, run timeout, per-iteration timeout, cost/token cap | Yes |
| Verification | test command, lint command, artifact validation, review command | Yes |
| Progress | progress cadence, summary format, artifact links, log retention | Yes |
| Recovery | retry count, backoff, yield timeout, resume input schema | Yes |
| Notifications | terminal output, artifact update, hook command, mobile notification | Optional in v1 |

## Interfaces

This taxonomy implies these later implementation interfaces:

- `run.begin`: create a run from an artifact, backlog item, or explicit input.
- `run.get`: read current run state, resolved settings, progress links, and last
  signal.
- `run.list`: list active and recent runs.
- `run.cancel`: request cancellation and force wrap-up evidence.
- `run.resume`: provide input after `yield`.
- `run.settings.validate`: check that settings are bounded before launch.
- `run.progress.append`: persist iteration summaries and verification evidence.

The initial implementation can avoid exposing every interface as MCP tools. It
can first express them as artifact shapes and internal functions, then promote
stable operations to tools after the workflow proves itself.

## Edge Cases

- The executor exits successfully but emits no parseable signal. Treat as
  `failed` unless a conservative parser can derive `yield` with diagnostics.
- Verification fails after `done`. Override the signal to `failed` or `yield`
  depending on whether retry policy can act without human input.
- A run reaches a budget limit while the work is partially complete. Stop as
  `failed` with progress artifacts and a suggested smaller follow-up item.
- The backlog source changes while the run is active. Continue only from the
  snapshotted input; later backlog changes require a new run or explicit resume
  payload.
- The worktree contains unrelated user edits. Preflight fails unless the allowed
  scope explicitly includes those files and the user approves.
- A notification hook fails. Continue the run if progress artifacts are durable;
  record the notification failure as non-blocking evidence.
- A persistent-session executor loses state. Fall back to a fresh prepare phase
  from durable progress artifacts or fail with recovery instructions.

## Test Plan

Validate the taxonomy by applying it to three dry examples before runtime work:

1. A docs-only artifact refinement run that must stay read-only outside
   `docs/artifacts/`.
2. A code implementation run that reaches `done`, fails verification, and loops
   back once.
3. A backlog-fed run that starts from a task-like artifact, yields for human
   input, resumes, then completes.

The [AFK System Test Strategy](afk-system-test-strategy.md) turns these examples
into contract, integration, and manual verification coverage for later code.

## Source References

- [AFK System RFC](afk-system-rfc.md)
- [Ripple signal vocabulary](https://github.com/manuelibar/ripple/blob/main/docs/reference/signal-vocabulary.md)
- [Ripple execution models](https://github.com/manuelibar/ripple/blob/main/docs/manifesto/10-execution-models.md)
- [Ripple blueprint schema](https://github.com/manuelibar/ripple/blob/main/docs/reference/blueprint-yaml-schema.md)
- [Codex config reference](https://developers.openai.com/codex/config-reference#configtoml)

## Open Questions

`AFK-HITL-003` decides how close the eventual run settings schema should stay to
Ripple blueprint YAML. `AFK-HITL-005` decides whether backlog-fed work is a v1
source or a later adapter.
