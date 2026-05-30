---
id: "agent-skill-distribution-problem-statement"
type: "problem_statement"
title: "Agent Skill Distribution Problem Statement"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Agent Skill Distribution Problem Statement

## Context

Workbench currently acts as a small local stdio MCP kernel. It keeps current
agent context in memory, stores typed Markdown artifacts, and exposes a limited
MCP surface through deterministic context changes.

Agent skills are emerging as reusable workflow packages made from metadata,
instructions, references, scripts, and assets. Codex treats skills as the
authoring format and plugins as its installable distribution unit. MCP, by
contrast, gives servers standard primitives for resources, prompts, and tools,
but does not define a skill package or install lifecycle.

This epic exists because Workbench is positioned to expose skills as MCP-native
context resources while still interoperating with Codex-compatible skill and
plugin conventions.

## Problem

There is no Workbench contract that says how an agent should discover a skill,
inspect its instructions and supporting files, understand its compatibility,
install it into an agent-specific location, activate or deactivate it, or update
it without confusing passive context with executable capability.

Without that contract, later implementation work could accidentally:

- expose skill instructions as prompts, forcing user-controlled prompt
  semantics onto reusable workflow packages;
- expose skill packages as tools, treating passive content discovery as model
  controlled execution;
- hard-code Codex plugin distribution as the only model, excluding other
  MCP-compatible agents;
- install or run untrusted scripts without an explicit approval boundary;
- lose version, provenance, compatibility, or digest information needed for
  safe updates.

## Impact

The immediate impact is design ambiguity. Agents cannot reliably determine
whether a Workbench-provided object is context, a prompt template, an executable
action, a server capability, a Codex plugin, or an installable skill package.

The downstream impact is broader:

- users may approve a harmless-looking skill and unintentionally grant tool or
  filesystem capability;
- agents may load too much skill content into context during discovery;
- teams may be unable to distribute reusable workflows across agent hosts;
- version drift may make activation and rollback difficult;
- Workbench may become tied to one agent's filesystem layout instead of
  remaining an MCP-oriented coordination server.

## Constraints

- Bootstrap work is docs-only and must update only `docs/artifacts/`.
- Final artifacts must derive from current Workbench docs, current repository
  state, and targeted current research, not archive branches.
- Workbench's current foundation is a local stdio MCP server with file-backed
  artifacts and no database requirement.
- MCP resources are application-controlled context; MCP prompts are
  user-controlled templates; MCP tools are model-controlled actions.
- Codex skills use progressive disclosure and `SKILL.md` metadata; Codex
  plugins are currently the Codex installable distribution unit.
- Generic MCP hosts may support resources but have no shared understanding of
  Codex skill folders or plugin marketplaces.
- Installation, activation, filesystem writes, network access, script
  execution, and tool enablement require explicit user consent.

## Source References

- [README.md](../../README.md)
- [docs/how-to/epic-branch-workflow.md](../how-to/epic-branch-workflow.md)
- [MCP resources](https://modelcontextprotocol.io/specification/2025-11-25/server/resources)
- [MCP prompts](https://modelcontextprotocol.io/specification/2025-11-25/server/prompts)
- [MCP tools](https://modelcontextprotocol.io/specification/2025-11-25/server/tools)
- [Codex Agent Skills](https://developers.openai.com/codex/skills)
- [Codex plugin build documentation](https://developers.openai.com/codex/plugins/build)

## Open Questions

Open human decisions are tracked in the RFC's Human-in-the-loop Index.
