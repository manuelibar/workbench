---
id: "afk-system-assumption"
type: "assumption"
title: "AFK Can Bootstrap on File-Backed Artifacts"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# AFK Can Bootstrap on File-Backed Artifacts

## Statement

The first Workbench AFK implementation can use file-backed artifacts and a small
local run ledger instead of requiring a database, queue, HTTP API, or external
orchestrator.

## Evidence

Workbench `main` is explicitly a stdio MCP context and artifact kernel with flat
Markdown artifacts and no database requirement. The epic workflow requires
self-contained artifact packets before runtime work.

Ripple's prior design proves the need for run state, signals, bounded budgets,
and progress evidence, but its Core plus NATS plus Postgres architecture solves
a larger distributed problem than Workbench v1. The useful primitive is the
bounded run contract, not the deployment topology.

External backlog tools also show that Markdown-backed task state can be useful
for AI collaboration when tasks have acceptance criteria, review checkpoints,
and local ownership. That supports a Workbench-first design where progress
artifacts are durable before heavier runtime services exist.

## Validation Plan

Validate the assumption in the first runtime pass by building the smallest
read-only AFK dry-run path:

1. Create a run settings artifact from an existing source artifact.
2. Validate budgets, source links, allowed paths, and verification command.
3. Simulate or execute one read-only Codex iteration.
4. Persist a progress artifact with input, signal, summary, and verification
   evidence.
5. Confirm the run can be inspected after process restart using only files.

The assumption fails if restart inspection, progress provenance, or bounded stop
conditions require state that cannot be reconstructed from files without
introducing hidden in-memory coupling.

## Source References

- [README.md](../../README.md)
- [Artifact conventions](../reference/artifact-conventions.md)
- [Ripple execution models](https://github.com/manuelibar/ripple/blob/main/docs/manifesto/10-execution-models.md)
- [Ripple loop runtime state ADR](https://github.com/manuelibar/ripple/blob/main/ripple-core/docs/adr/0042-loop-runtime-state.md)
- [Backlog.md](https://github.com/MrLesk/Backlog.md)

## Open Questions

`AFK-HITL-003` may revise the run settings shape, but the default remains a
small Workbench-native profile rather than blueprint compatibility.
