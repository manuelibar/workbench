---
id: "agent-skill-distribution-charter"
type: "charter"
title: "Agent Skill Distribution Charter"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Agent Skill Distribution Charter

## Mission

Define how Workbench can expose, describe, version, and distribute agent skills
through MCP resources without collapsing skills into prompts, tools, or
service-specific plugin formats.

The epic should make skills discoverable to Codex-compatible agents and other
MCP-compatible agents while preserving progressive disclosure: agents first see
small metadata, then read instructions, references, assets, or install plans
only when the task calls for them.

## Scope

In scope:

- MCP resource and resource-template contracts for skill catalogs, skill
  manifests, skill instructions, references, assets, and versioned snapshots.
- Skill metadata needed for discovery, compatibility, provenance, license,
  dependencies, activation policy, and install planning.
- Boundaries between skill content, MCP prompts, MCP tools, MCP server
  capabilities, Codex plugins, and external package registries.
- Agent-facing discovery, install planning, installation, activation,
  deactivation, and update flows.
- Compatibility profiles for Codex-compatible agents and generic
  MCP-compatible hosts.
- Trust and safety rules for untrusted skill text, scripts, references, and
  bundled service capabilities.

Out of scope for the bootstrap contract:

- Runtime implementation, persistence schema, or migration code.
- Building a public package registry or replacing the MCP Registry.
- Executing skill scripts during discovery or installation.
- Codex plugin publishing, app connector publishing, or marketplace curation
  beyond adapter metadata needed to interoperate with those surfaces.
- Skill authoring UX beyond the metadata and package shape required for
  distribution.

## Stakeholders

- Local agents using Workbench as a stdio MCP context server.
- Codex-compatible agents that understand `SKILL.md` directories, progressive
  disclosure, and plugin-based installation.
- MCP-compatible hosts that can list and read resources but may not understand
  Codex-specific skill locations.
- Skill authors who need a stable package shape, compatibility vocabulary, and
  install expectations.
- Users and workspace owners who approve installation, activation, network
  access, filesystem writes, and execution of any bundled scripts or tools.
- Future Workbench implementation owners who will turn this packet into MCP
  resources, tools, validation, and tests.

## Success Criteria

- Workbench exposes a read-only skill catalog as MCP resources with stable URIs,
  MIME types, resource annotations, pagination expectations, and list-changed
  behavior.
- Each advertised skill has enough metadata for an agent to decide whether to
  inspect, install, activate, defer, or reject it.
- Installation and activation are explicit flows with human approval before
  filesystem writes, tool enablement, network access, or script execution.
- Versioning combines immutable content identity with publisher-facing version
  labels and clear compatibility declarations.
- Codex-specific plugin and skill conventions are supported through adapters
  without becoming the generic Workbench contract.
- The first implementation pass can be planned from
  [agent-skill-distribution-rfc.md](agent-skill-distribution-rfc.md) and
  [agent-skill-distribution-initial-implementation-plan.md](agent-skill-distribution-initial-implementation-plan.md)
  without referring to archive branches.

## Source References

- [README.md](../../README.md)
- [docs/how-to/epic-branch-workflow.md](../how-to/epic-branch-workflow.md)
- [docs/reference/artifact-conventions.md](../reference/artifact-conventions.md)
- [MCP 2025-11-25 specification](https://modelcontextprotocol.io/specification/2025-11-25)
- [MCP 2025-11-25 server feature overview](https://modelcontextprotocol.io/specification/2025-11-25/server/index)
- [Codex Agent Skills](https://developers.openai.com/codex/skills)
- [Codex plugin build documentation](https://developers.openai.com/codex/plugins/build)
- [Agent Skills specification](https://agentskills.io/specification)

## Open Questions

Open human decisions are tracked in the RFC's Human-in-the-loop Index.
