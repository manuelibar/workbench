---
id: "afk-system-rfc"
type: "rfc"
title: "Bootstrap the AFK System"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Bootstrap the AFK System

## Summary

Define the AFK system as Workbench's bounded autonomous run layer. The first
contract should be local, file-backed, artifact-driven, and executor-neutral
while making Codex the first practical integration target.

This RFC is the living drill hub for the epic packet. It links the stable packet
artifacts and the action artifacts that later implementation sessions should
use:

- [Charter](afk-system-charter.md)
- [Problem Statement](afk-system-problem-statement.md)
- [Concept Map](afk-system-concept-map.md)
- [Assumption](afk-system-assumption.md)
- [Risk](afk-system-risk.md)
- [Research Note](afk-system-research-note.md)
- Action: [Runbook](afk-system-runbook.md)
- Action: [Test Strategy](afk-system-test-strategy.md)

## Problem

Workbench does not yet have a safe contract for autonomous work. The current
kernel can hold context and artifacts, but it cannot start, bound, monitor,
pause, resume, cancel, or verify an AFK run.

The user's prior Ripple documentation already explored richer AFK primitives:
blueprints, modes, run state, iteration signals, bounded execution, worker
budgets, progress traces, and commit provenance. Those ideas are useful, but the
old architecture assumes Core, NATS, Postgres, worker services, and broader
session/memory systems that Workbench `main` intentionally does not have.

Workbench needs a smaller contract that preserves the autonomy invariants:

- A run must know its input, scope, settings, and owner.
- A run must stop predictably.
- Progress must survive process restart.
- Human questions, approvals, decisions, tradeoffs, and challenges must be
  indexed.
- Write-enabled runs must prove safety before touching code.
- Backlog-fed runs must not block on the backlog epic's final implementation.

## Proposal

### Runtime shape

Start with a Workbench-native AFK run model:

1. A source artifact, backlog item, or explicit input creates a run draft.
2. Run settings resolve into a snapshotted profile.
3. Preflight checks verify scope, branch, worktree, budgets, executor profile,
   verification commands, and source links.
4. The executor performs one iteration.
5. Wrap-up parses a signal, runs verification when applicable, persists progress,
   and decides whether to continue, yield, fail, complete, or cancel.
6. Every state transition appends durable evidence.

The first implementation should use file-backed artifacts or a small file ledger
for run state. A database or queue can be revisited only after the file-backed
path fails a concrete scaling or reliability requirement.

### Reused primitives

Reuse these Ripple primitives conceptually:

- `continue`, `yield`, `done`, and `failed` as the base signal vocabulary.
- Prepare, execute, and wrap-up as the iteration boundary.
- Bounded execution as a design invariant.
- Resolved config snapshot at run start.
- Traceability evidence for run, iteration, commit, cost, and verification.

### Adapted primitives

Adapt these primitives:

- Blueprint YAML becomes a smaller Workbench run settings profile.
- Modes become optional profile metadata until mode mutation exists in Workbench.
- MCP tasks become progress artifacts and terminal status first, notification
  hooks later.
- Backlog-fed work starts through a narrow task-source boundary and can initially
  consume typed artifacts.
- Commit trailers remain an idea for write-enabled runs, not a docs-bootstrap
  requirement.

### Reimplemented primitives

Reimplement these pieces for Workbench:

- Run ledger and state inspection.
- Codex executor adapter and launch boundary.
- Safety gate validation.
- Progress artifact writer.
- Backlog adapter.
- Cancellation and resume operations.

### Stop conditions

An AFK run must stop or pause on:

- Accepted `done` signal plus successful verification.
- `failed` signal after retry policy is exhausted.
- User cancellation.
- Wall-clock timeout.
- Per-iteration timeout.
- Iteration cap.
- Cost or token budget cap.
- Yield timeout.
- Worktree or branch drift.
- Dirty files outside the allowed scope.
- Verification failure when retry cannot continue safely.
- Missing or unparseable signal.
- Safety gate failure.

### Progress artifacts

Every run should produce durable progress evidence:

- Run input and source references.
- Resolved run settings.
- Preflight result.
- Per-iteration summary, signal, changed files, verification command, and result.
- Human nudges and responses.
- Final status with commit hashes or artifact links when applicable.
- Follow-up work if the run fails, yields, or stops on budget.

### Safety gate minimum

The default write-enabled safety gate should require:

- Clean worktree before launch.
- Isolated branch or worktree.
- Explicit writable scope.
- Explicit verification command or a documented reason no command applies.
- Bounded time, iterations, and cost or token budget.
- Known executor profile and approval policy.
- No push, deploy, destructive cleanup, or secret-modifying operation unless a
  later policy explicitly allows it.
- Human review before any generated commit is pushed.

### Backlog-fed work

AFK should accept a task source shaped like:

```yaml
source:
  kind: artifact | backlog
  id: string
  title: string
  acceptance_criteria: [string]
  context_refs: [string]
  writable_scope: [string]
  verification: [string]
```

Until the backlog epic exists, `kind: artifact` is enough. The backlog adapter
can later map backlog tasks into this shape without changing the run lifecycle.

## Tradeoffs

Using a file-backed run ledger is less powerful than Ripple's Core plus
NATS/Postgres design. It will not immediately support distributed workers,
multi-user dashboards, durable queues, or complex event replay. The benefit is
that it fits Workbench `main`, keeps the first implementation inspectable, and
avoids premature infrastructure.

Adapting blueprints into smaller run settings risks losing compatibility with
the richer Ripple schema. The benefit is that Workbench can avoid committing to
mode systems, memory providers, and trigger policies before those epics define
their own contracts.

Defaulting to three-phase iterations costs more context than persistent sessions.
The benefit is crash safety: progress artifacts can reconstruct the next
iteration after restart.

Starting with artifact-fed work means backlog integration is not complete in v1.
The benefit is that AFK can make progress independently while preserving a clean
adapter boundary for the backlog epic.

## Rollout

1. Land this docs-only bootstrap packet.
2. Drill the RFC into a run settings artifact and a progress artifact schema.
3. Implement read-only dry-run support that validates settings and writes
   progress evidence.
4. Add a Codex executor adapter behind an explicit profile without mutating
   user-level config.
5. Add write-enabled support only after `AFK-HITL-002` is resolved.
6. Add cancellation, resume, and yield timeout handling.
7. Add backlog adapter when the backlog epic defines its source contract.
8. Add notification hooks after progress artifacts are reliable.

## Open Questions

### Human-in-the-loop Index

| ID | Nudge | Type | Why it matters | Blocks | Default if unanswered |
|---|---|---|---|---|---|
| AFK-HITL-001 | Decide whether v1 may write code autonomously or must start read-only/dry-run. | decision | Write access changes the safety, verification, and review burden. | Write-enabled implementation scope. | Start with read-only/dry-run and require a later explicit approval for writes. |
| AFK-HITL-002 | Approve the minimum write-enabled safety gate. | approval | The system needs a non-negotiable boundary before delegating code changes. | Any write-enabled run, generated commit, or push policy. | Require clean worktree, isolated branch/worktree, explicit writable scope, verification command, bounded budgets, known executor profile, and human review before push. |
| AFK-HITL-003 | Choose whether Workbench run settings should be blueprint-compatible or only blueprint-inspired. | tradeoff | Compatibility could ease reuse but may import unnecessary schema complexity. | Run settings schema implementation. | Use a smaller Workbench-native run profile that borrows Ripple vocabulary without YAML compatibility guarantees. |
| AFK-HITL-004 | Choose the first notification/status surface for unattended runs. | question | AFK value depends on knowing when work finishes or needs attention. | Notification adapter design. | Use durable progress artifacts and terminal-visible status first; add hook/mobile integrations later. |
| AFK-HITL-005 | Challenge whether backlog-fed AFK should wait for the backlog epic or start from artifact-shaped tasks. | challenge | Blocking on backlog delays AFK, but premature integration can create the wrong task contract. | Backlog-fed launch design. | Support artifact-shaped tasks first and reserve a backlog adapter boundary. |

## Source References

- [AFK System Research Note](afk-system-research-note.md)
- [AFK System Concept Map](afk-system-concept-map.md)
- [AFK System Risk](afk-system-risk.md)
- [README.md](../../README.md)
- [Artifact conventions](../reference/artifact-conventions.md)
- [Epic branch workflow](../how-to/epic-branch-workflow.md)
