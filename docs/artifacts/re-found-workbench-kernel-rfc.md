---
id: "re-found-workbench-kernel-rfc"
type: "rfc"
title: "Re-found Workbench as a Context and Artifact Kernel"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Re-found Workbench as a Context and Artifact Kernel

## Summary

Replace the integrated HTTP/Postgres feature surface on `main` with a stdio
MCP context and artifact kernel.

## Problem

The previous `main` mixed foundation behavior with backlog, notes, projects,
namespaces, roles, prompts, blueprints, modes, onboarding, Postgres, HTTP, and
snapshot planning surfaces. That made the branch too broad to serve as a
stable base for feature drilling.

## Proposal

Keep the `contextualize` tool and file-backed typed artifacts on `main`. Move
non-core systems to epic branches that start from the cleaned foundation and
define their intended contracts through kickoff artifact packets before porting
old code.

## Tradeoffs

The breaking change removes useful old functionality from `main`, and historical
branches may help orient later recovery work. Final epic packets should still
derive their contracts from current repository state, current documentation, and
targeted research so stale implementation details do not silently become the
new contract.

## Rollout

Land one breaking-change commit on `main`, then create the epic branches from
that new commit. Push `main` and the current epic branches to GitHub.

## Open Questions

Each epic owns its feature-specific open questions.

### Human-in-the-loop Index

| ID | Nudge | Type | Why it matters | Blocks | Default if unanswered |
|---|---|---|---|---|---|

No foundation nudges are open.

## Source References

- `docs/how-to/epic-branch-workflow.md`
