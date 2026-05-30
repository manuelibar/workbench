---
id: "afk-system-problem-statement"
type: "problem_statement"
title: "AFK System Problem Statement"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# AFK System Problem Statement

## Context

Workbench `main` is a small stdio MCP context and artifact kernel. It stores
typed Markdown artifacts under `docs/artifacts/`, exposes deterministic context
selection, and intentionally defers backlog, sessions, memory, and AFK behavior
to epic branches.

The user also has earlier Ripple documentation for a broader agent-native system:
blueprints, modes, signals, bounded execution, run state, workers, commit
trailers, observability, and an autonomous execution engine. Those primitives
were written for a larger Core plus NATS/Postgres architecture and for multiple
executors. Workbench needs the useful invariants without inheriting unnecessary
runtime weight.

Current external tooling reinforces the same problem shape. Backlog-style tools
make work durable in Markdown; kanban-oriented agent tools isolate each agent in
workspaces and branches; AFK desktop/mobile tools emphasize notification and
remote observation. None of those directly define the small, local-first
Workbench contract.

## Problem

Workbench lacks a contract for bounded autonomous runs. Without that contract, a
future AFK implementation would have to improvise answers to high-risk
gaps:

- The exact input that starts a run.
- The files or commands the agent may touch.
- The conditions that force the run to stop.
- The persistence model for progress, verification, cost, and decisions.
- The boundary between continuing autonomously and yielding to the user.
- The path for backlog-fed work before the backlog epic is done.
- The prior Ripple primitives that are safe to reuse versus too coupled to the
  old architecture.

The absence of a docs-first packet makes AFK risky because autonomy can hide
state changes, spend budget, continue from stale assumptions, or block silently.

## Impact

The impact is primarily workflow and safety:

- The user cannot delegate multi-step work confidently because no bounded run
  envelope exists.
- Later implementation agents may over-port Ripple's old infrastructure instead
  of adapting the few primitives Workbench actually needs.
- Backlog-fed autonomous work may become tightly coupled to a future backlog
  implementation instead of using a narrow task-source interface.
- Reviewers may receive commits without enough provenance to understand what ran,
  why it stopped, and which verification gates passed.
- Codex-specific configuration such as sandbox mode, approval policy, MCP
  servers, notifications, and history may be mixed into Workbench state without a
  clear ownership boundary.

## Constraints

- This bootstrap pass is docs-only and edits only `docs/artifacts/`.
- Final artifacts must cite current local docs and current external sources.
- Workbench `main` has no database, HTTP listener, or durable session store.
- The RFC must be the living drill hub and include the human-in-the-loop index
  required by artifact conventions.
- The first implementation should prefer Workbench-native file-backed artifacts
  over NATS/Postgres orchestration unless later evidence justifies the heavier
  path.
- Any write-enabled AFK run must have explicit bounded settings, safety checks,
  and reviewable progress artifacts before it can be considered acceptable.

## Source References

- [AFK System RFC](afk-system-rfc.md)
- [AFK System Research Note](afk-system-research-note.md)
- [README.md](../../README.md)
- [Epic branch workflow](../how-to/epic-branch-workflow.md)
- [Artifact conventions](../reference/artifact-conventions.md)

## Open Questions

The problem framing is blocked from runtime implementation by `AFK-HITL-001`,
`AFK-HITL-002`, and `AFK-HITL-005` in the RFC's human-in-the-loop index.
