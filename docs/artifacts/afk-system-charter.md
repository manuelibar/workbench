---
id: "afk-system-charter"
type: "charter"
title: "AFK System Charter"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# AFK System Charter

## Mission

Define Workbench's AFK system as a bounded, inspectable, docs-first autonomy
layer for running coding agents away from the keyboard. The system should let a
user delegate backlog-fed or artifact-fed work, preserve progress as durable
artifacts, stop predictably, and surface human decisions before risk escapes the
run boundary.

## Scope

In scope:

- Bounded autonomous run lifecycle and run state vocabulary.
- Stop conditions for completion, failure, cancellation, yield timeout, budget
  exhaustion, safety gate failure, and stale worktree detection.
- Run settings for iteration count, wall-clock budget, cost or token budget,
  executor profile, sandbox/network policy, writable scope, verification
  command, progress cadence, and notification hooks.
- Runbooks for launching, monitoring, yielding, resuming, cancelling, and
  reviewing AFK runs.
- Progress artifacts that capture run input, iteration summaries, verification
  results, human nudges, final status, and source references.
- Safety checks that make write-enabled autonomous work explicit and auditable.
- Backlog-fed autonomous work once the backlog epic provides a task source; until
  then, the AFK system can consume typed artifacts and explicit run inputs.
- Research-driven reuse, adaptation, or reimplementation of the user's prior
  Ripple primitives for Workbench and Codex.

Out of scope for this bootstrap packet:

- Runtime code, background daemons, database migrations, queues, web dashboards,
  mobile clients, or push notification services.
- Porting Ripple's NATS/Postgres execution engine into Workbench during this
  docs-only pass.
- Replacing Codex, Claude Code, Goose, or any other coding agent executor.
- Pushing commits or opening pull requests from AFK runs without a later,
  explicit write-enabled policy decision.

## Stakeholders

- The local Workbench user, who needs delegation without losing control of
  scope, cost, safety, or review responsibility.
- Codex or another coding agent executor, which needs a precise run contract,
  allowed context, stop conditions, and output format.
- Future backlog, artifact, and session epics, which need stable AFK concepts
  that can compose with their own packet contracts.
- Maintainers reviewing AFK output, who need provenance, verification evidence,
  and a clear path to replay or cancel work.
- The AFK system epic owner, accountable for keeping this packet self-contained
  and current-source based.

## Success Criteria

The AFK system epic is ready to move from docs to implementation when:

- A reviewer can understand the run lifecycle, settings, stop conditions, and
  safety gates from the packet without reading chat history.
- The RFC links every action artifact and can be used as the living drill hub for
  later implementation sessions.
- The packet identifies which Ripple primitives are reused conceptually, which
  are adapted to Workbench's artifact kernel, and which must be reimplemented.
- Write-enabled AFK work has an explicit safety gate policy before any runtime
  code can launch it.
- A backlog-fed path exists that does not require the backlog epic to be
  complete on day one.
- Progress artifacts are treated as the durable source of run truth, not as
  incidental logs.

## Source References

- [AFK System RFC](afk-system-rfc.md)
- [AFK System Concept Map](afk-system-concept-map.md)
- [AFK System Runbook](afk-system-runbook.md)
- [README.md](../../README.md)
- [Artifact conventions](../reference/artifact-conventions.md)
- [Epic branch workflow](../how-to/epic-branch-workflow.md)

## Open Questions

The charter depends on the indexed nudges in the RFC, especially
`AFK-HITL-001` for the first write-enabled scope and `AFK-HITL-002` for the
minimum safety gate policy.
