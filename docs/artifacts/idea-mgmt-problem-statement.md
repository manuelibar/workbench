---
id: "idea-mgmt-problem-statement"
type: "problem_statement"
title: "Idea Management Problem Statement"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Idea Management Problem Statement

## Context

Workbench `main` is currently a small stdio MCP kernel for current context and
typed Markdown artifacts. The foundation deliberately defers backlog, notes,
memory, knowledge, sessions, and related systems to epic branches that define
contracts before runtime work. Ideas cut across those future systems: an idea
can start as a fleeting note, become research input, turn into an artifact,
feed a work item, expose an opportunity, update knowledge, or become a memory
candidate.

## Problem

There is no current contract for preserving ideas while they are still
speculative. Without one, agents either leave useful ideas inside transient
chat context, prematurely convert them into artifacts or work items, or scatter
them across notes with no lifecycle, relation model, or promotion trace. That
blurs the boundary between "this may be worth exploring" and "this is now
committed work."

## Impact

The missing idea layer causes context loss between sessions, duplicate
exploration, weak provenance for promoted decisions, and noisy committed work
surfaces. It also makes it harder for a human to see which ideas need
refinement, which ideas are intentionally parked, and which ideas already
became artifacts, work items, opportunities, knowledge, or memory.

## Constraints

This epic packet must be docs-only and self-contained under `docs/artifacts/`.
It must derive from the current repository docs and targeted current research.
Any future implementation must respect the existing artifact conventions, keep
deterministic validation possible, and fit Workbench's local stdio MCP model
where resources provide context and tools perform explicit actions with
appropriate human control.

## Source References

- `README.md`
- `docs/how-to/epic-branch-workflow.md`
- `docs/reference/artifact-conventions.md`
- `docs/artifacts/idea-mgmt-concept-map.md`
- `docs/artifacts/idea-mgmt-rfc.md`

## Open Questions

The problem framing assumes raw ideas remain distinct from committed work; the
RFC tracks the raw-capture edit policy as `IM-HIL-003`.
