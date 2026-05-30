---
id: "agent-skill-distribution-assumption"
type: "assumption"
title: "Agent Skill Distribution Assumption"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Agent Skill Distribution Assumption

## Statement

Workbench should model distributed skills as resource-first packages, with tools
reserved for explicit mutations such as install planning, installation,
activation, deactivation, and updates.

The first implementation should not require a public registry, a Codex plugin,
or a host-specific skill directory to exist before an MCP-compatible agent can
discover and inspect a skill.

## Evidence

MCP defines resources as structured data or content that provides context to
the model, and resource interaction is application-controlled by the host.
Skills are mostly contextual artifacts until a user chooses to install or
activate them.

MCP defines prompts as user-controlled templates and tools as model-controlled
functions. Those semantics do not fit a passive package of instructions,
references, assets, and optional scripts.

Codex documentation says skills use progressive disclosure and are loaded first
by small metadata, then by full instructions only when selected. The open Agent
Skills specification also describes metadata-first loading and on-demand
references. That matches MCP resource discovery better than prompt or tool
exposure.

Codex documentation also says plugins are the installable distribution unit for
Codex, while skills remain the authoring format. That supports treating Codex
plugins as an adapter target rather than the generic Workbench contract.

## Validation Plan

- Draft the RFC around resource-first discovery and verify the concept map keeps
  prompts, tools, resources, service capabilities, and plugins separate.
- In the first implementation pass, create fixtures for at least one
  instruction-only skill and one skill with references/assets/scripts, then
  expose both through resources before adding mutation tools.
- Test that generic MCP resource reads can inspect the skill manifest and
  instructions without Codex-specific plugin support.
- Test that install, activation, and script execution are impossible through
  resource reads alone and require an explicit mutation flow.
- Revisit the assumption if a target host cannot present resources usefully but
  can present prompts or tools.

## Source References

- [MCP resources](https://modelcontextprotocol.io/specification/2025-11-25/server/resources)
- [MCP prompts](https://modelcontextprotocol.io/specification/2025-11-25/server/prompts)
- [MCP tools](https://modelcontextprotocol.io/specification/2025-11-25/server/tools)
- [Codex Agent Skills](https://developers.openai.com/codex/skills)
- [Agent Skills specification](https://agentskills.io/specification)

## Open Questions

Open human decisions are tracked in the RFC's Human-in-the-loop Index.
