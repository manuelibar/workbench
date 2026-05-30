---
id: "agent-skill-distribution-risk"
type: "risk"
title: "Agent Skill Distribution Risk"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Agent Skill Distribution Risk

## Description

Skill packages can blur the boundary between passive context and active
capability. A package may contain instructions, prompt-like language, scripts,
tool allowlists, MCP server dependencies, app connector mappings, or lifecycle
hooks. If Workbench exposes or installs that package without clear boundaries,
an agent or user may treat untrusted content as trusted capability.

The most important risk event is accidental privilege escalation: discovery or
activation causes filesystem writes, script execution, tool enablement, network
access, or connector activation that the user did not explicitly approve.

## Impact

If the risk materializes:

- untrusted skill text could steer the agent through prompt injection;
- bundled scripts or hooks could run with local filesystem access;
- a skill could request tools or MCP servers beyond the user's intent;
- an update could replace a trusted skill with different behavior under the
  same name;
- agents could install incompatible packages into the wrong host-specific
  directory;
- users may lose confidence in Workbench as a local-first coordination server.

The impact is high because skill distribution crosses trust, context, and
execution boundaries.

## Likelihood

Likelihood is medium. The current epic is docs-only, so there is no immediate
runtime exposure. The likelihood rises during implementation because Codex
skills, Codex plugins, MCP tools, and MCP resources all have adjacent but
different control semantics.

Confidence is medium-high because both MCP and Codex documentation explicitly
call out consent, progressive disclosure, tool safety, and installation
boundaries.

## Mitigation

- Treat all skill package content as untrusted until installed and activated by
  an explicit user-approved flow.
- Expose skill content through read-only resources first.
- Separate install planning from install application, and include target path,
  version, digest, requested capabilities, and rollback hints in the plan.
- Refuse to execute scripts, hooks, or bundled tools during discovery and
  metadata reads.
- Require compatibility and requested-capability metadata before activation.
- Preserve immutable content digests and publisher/version metadata in
  install records.
- Surface Codex plugin, MCP server, app connector, and hook requests as
  capability requests rather than ordinary skill metadata.
- Keep generic MCP support independent from Codex-specific install adapters.

## Owner

The epic owner for `epic/agent-skill-distribution` owns the risk until a later
implementation packet assigns runtime module owners. Future implementation
owners should include maintainers of resource listing, install tools,
compatibility validation, and agent-specific adapters.

## Source References

- [MCP 2025-11-25 specification](https://modelcontextprotocol.io/specification/2025-11-25)
- [MCP tools security considerations](https://modelcontextprotocol.io/specification/2025-11-25/server/tools)
- [MCP registry trust and security](https://modelcontextprotocol.io/registry/about)
- [Codex Agent Skills](https://developers.openai.com/codex/skills)
- [Codex plugin build documentation](https://developers.openai.com/codex/plugins/build)

## Open Questions

Open human decisions are tracked in the RFC's Human-in-the-loop Index.
