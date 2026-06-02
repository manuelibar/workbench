---
id: "workbench-context-artifact-kernel-charter"
type: "charter"
title: "Workbench Context and Artifact Kernel Charter"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Workbench Context and Artifact Kernel Charter

## Mission

Keep `main` as a small stdio MCP kernel that manages current context and typed
Markdown artifacts.

## Scope

In scope: context patching, capability planning, capability relist sync,
`workbench:///context`, artifact resources, artifact draft creation, guidance,
validation, listing, reading, and section updates.

Out of scope: backlog, ideas, knowledge, AFK loops, persisted sessions,
namespaces, roles, memory, and advanced artifact workflows.

## Stakeholders

The primary stakeholder is the local agent workflow that uses Workbench as a
context and artifact coordination server.

## Success Criteria

The foundation is successful when `main` runs without Postgres or HTTP, the
`contextualize` tool reliably synchronizes capability changes, artifacts are
durable Markdown files, and every deferred subsystem has an epic branch with
kickoff artifacts.

## Source References

- `docs/reference/context-contract.md`
- `docs/reference/artifact-conventions.md`

## Open Questions

None for the foundation cut.
