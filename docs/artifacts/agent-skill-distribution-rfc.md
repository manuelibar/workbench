---
id: "agent-skill-distribution-rfc"
type: "rfc"
title: "Distribute Agent Skills as MCP Resources"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Distribute Agent Skills as MCP Resources

## Summary

Use Workbench to distribute agent skills as MCP resources: a compact catalog for
discovery, versioned manifests for decision-making, and versioned file resources
for instructions, references, assets, and scripts. Reserve MCP tools for
explicit install planning, installation, activation, deactivation, and updates.

This RFC is the living drill hub for the epic. Action artifacts created from it:

- [agent-skill-distribution-initial-implementation-plan.md](agent-skill-distribution-initial-implementation-plan.md)

The surrounding packet:

- [agent-skill-distribution-charter.md](agent-skill-distribution-charter.md)
- [agent-skill-distribution-problem-statement.md](agent-skill-distribution-problem-statement.md)
- [agent-skill-distribution-concept-map.md](agent-skill-distribution-concept-map.md)
- [agent-skill-distribution-assumption.md](agent-skill-distribution-assumption.md)
- [agent-skill-distribution-risk.md](agent-skill-distribution-risk.md)
- [agent-skill-distribution-research-note.md](agent-skill-distribution-research-note.md)

## Problem

Skills are reusable workflow packages, but MCP does not define "skill" as a
first-class primitive. Workbench must choose how to expose them using MCP's
existing primitives without creating ambiguous capability boundaries.

If skills are exposed as prompts, they inherit a user-controlled template model
that does not represent a package with versions, references, assets, optional
scripts, compatibility, and install targets. If skills are exposed as tools,
passive package discovery becomes model-controlled execution. If Codex plugins
become the canonical contract, Workbench overfits to one agent's distribution
format and excludes other MCP-compatible hosts.

The missing contract must answer:

- how agents discover available skills without loading too much context;
- how agents inspect package metadata, instructions, references, assets, and
  scripts safely;
- how versioning, provenance, compatibility, and install targets are expressed;
- how installation and activation happen with explicit user approval;
- how Codex-compatible and generic MCP-compatible agents share the same
  distribution surface.

## Proposal

### Resource-first Distribution

Workbench should expose skills through a new skill resource namespace. The
catalog and manifest resources are the canonical discovery and inspection
surface. Agents can read progressively:

1. catalog metadata;
2. full manifest for a candidate skill;
3. primary instructions;
4. references, assets, scripts, or install-target guidance on demand.

Resource reads are read-only. They must not install files, enable tools, add MCP
servers, execute scripts, or change activation state.

### Skill Manifest

Each skill manifest should include:

- stable identity: package `id`, skill `name`, title, description, publisher,
  source URL, and repository URL when available;
- versioning: publisher version, immutable content digest, channel or source
  ref, creation/update timestamps, and selected version;
- package contents: instructions path, references, assets, scripts, file sizes,
  MIME types, and digests;
- compatibility: skill specification version, MCP protocol revisions, generic
  MCP support, Codex skill support, Codex plugin support, and required host
  features;
- safety and trust: license, provenance, requested tools, requested MCP
  servers, requested network access, filesystem writes, script execution, hook
  definitions, and trust status;
- activation policy: implicit invocation, explicit invocation, default enabled
  state, and agent-facing notes;
- install targets: known target profiles, target paths or generated guidance,
  approval requirements, and rollback hint.

Unknown metadata should be preserved so Workbench can round-trip
Codex-specific, open-skill, or future registry fields without making every field
normative.

### MCP Resource Shape

The first implementation should support plain resource listing before relying on
resource templates. Candidate resource families:

- `workbench-skill:///catalog`
- `workbench-skill:///skills/{skill_id}/manifest`
- `workbench-skill:///skills/{skill_id}/versions/{version}/manifest`
- `workbench-skill:///skills/{skill_id}/versions/{version}/instructions`
- `workbench-skill:///skills/{skill_id}/versions/{version}/files/{path}`
- `workbench-skill:///skills/{skill_id}/install-targets/{agent_profile}`

Resources should include MIME type, size where known, descriptions suitable for
model discovery, and annotations for audience, priority, and last modification
time where useful.

### Tools for Mutation

Tools should appear only when implementation reaches approved mutation flows:

- plan install without mutating state;
- apply an approved install plan;
- activate an installed skill;
- deactivate an installed skill;
- compare available versions and plan update;
- apply an approved update.

Each mutation tool should expose what will be read, written, enabled, disabled,
or executed. Install apply must verify that the planned version and digest still
match the package being installed.

### Compatibility Profiles

The generic profile is "MCP resource compatible": an agent can list and read
skill resources, understand the manifest summary, and stop before mutation.

The Codex-compatible profile can additionally map packages to Codex skill
folders, plugin manifests, optional `agents/openai.yaml` metadata, plugin MCP
server config, app mappings, and hook trust review.

Profiles for other MCP-compatible agents should be additive. Workbench should
not require Codex plugin support for generic discovery and inspection.

### Versioning

A package version should have both:

- a publisher-facing version string, preferably semantic when supplied;
- an immutable content digest over the files being installed or activated.

Resource URIs for versioned content should be stable. Mutable aliases such as
"latest" may exist in catalog metadata, but install plans should pin concrete
version and digest.

### Trust Defaults

Default trust posture:

- reading metadata and instructions is allowed as untrusted context;
- scripts and hooks are inert files until installed and separately approved;
- requested tools, MCP servers, app connectors, network access, and filesystem
  writes are capability requests;
- activation is explicit unless the host profile and manifest both allow
  implicit invocation;
- updates are reviewable plans before mutation.

## Tradeoffs

Resource-first distribution may require more client logic than exposing a single
install tool. The benefit is clearer separation between passive context and
active mutation, and better support for hosts that can inspect resources but do
not support Workbench-specific tools.

Supporting Codex plugins as an adapter instead of the canonical package model
means some Codex-specific distribution features need mapping. The benefit is a
generic contract that remains useful to other MCP-compatible agents.

Pairing publisher versions with content digests adds metadata and validation
work. The benefit is safer install, update, and rollback behavior when mutable
sources or registry metadata change.

Keeping scripts inert during discovery limits turnkey automation. The benefit
is a safer default boundary for untrusted package contents.

## Rollout

1. Land this docs-only packet as the bootstrap contract for the epic.
2. Drill the manifest schema and resource URI grammar in follow-up artifacts if
   the Human-in-the-loop Index changes defaults.
3. Implement read-only resource discovery and manifest validation first.
4. Add resource templates and list-changed notifications after plain resource
   listing is stable.
5. Add install planning as a dry-run tool with no mutation.
6. Add approved install apply after digest and target-path verification are
   covered by tests.
7. Add activation/deactivation and Codex adapter behavior.
8. Add update planning and update apply after installed records exist.

Compatibility impact should be additive. Existing Workbench context and
artifact resources should continue to work while skill resources are introduced.

## Open Questions

### Human-in-the-loop Index

| ID | Nudge | Type | Why it matters | Blocks | Default if unanswered |
|---|---|---|---|---|---|
| ASD-HITL-001 | Decide whether Workbench should be only a local skill package server or also a registry aggregator. | decision | Aggregation changes metadata, caching, trust, and update scope. | Registry and source ingestion design. | Start as a local package server; add registry aggregation later through a separate artifact. |
| ASD-HITL-002 | Approve the first supported package shape. | approval | The shape determines fixture design and validation. | First implementation fixtures and manifest schema. | Support a directory package with `SKILL.md` instructions plus optional `references/`, `assets/`, and `scripts/`; no script execution at install time. |
| ASD-HITL-003 | Choose generic MCP compatibility or Codex compatibility as the normative baseline. | tradeoff | A Codex-first baseline could simplify one adapter but exclude other hosts. | Manifest required fields and install-target semantics. | Make MCP resource compatibility normative and Codex compatibility an adapter profile. |
| ASD-HITL-004 | Challenge whether skills truly belong as resources instead of prompts, tools, or plugins. | challenge | This is the core primitive boundary for the epic. | Resource URI grammar and tool surface. | Keep skills resource-first; use tools only for approved mutation and plugins only as adapter/distribution metadata. |
| ASD-HITL-005 | Decide how strict version labels must be. | question | Strict semantic versions improve sorting but may reject useful upstream packages. | Version selection and update planning. | Accept publisher strings, prefer semantic versions, and require immutable content digests for install/apply. |

## Source References

- [README.md](../../README.md)
- [docs/how-to/epic-branch-workflow.md](../how-to/epic-branch-workflow.md)
- [docs/reference/artifact-conventions.md](../reference/artifact-conventions.md)
- [MCP 2025-11-25 specification](https://modelcontextprotocol.io/specification/2025-11-25)
- [MCP 2025-11-25 server feature overview](https://modelcontextprotocol.io/specification/2025-11-25/server/index)
- [MCP resources](https://modelcontextprotocol.io/specification/2025-11-25/server/resources)
- [MCP prompts](https://modelcontextprotocol.io/specification/2025-11-25/server/prompts)
- [MCP tools](https://modelcontextprotocol.io/specification/2025-11-25/server/tools)
- [MCP registry overview](https://modelcontextprotocol.io/registry/about)
- [MCP registry versioning](https://modelcontextprotocol.io/registry/versioning)
- [MCP registry package types](https://modelcontextprotocol.io/registry/package-types)
- [Codex Agent Skills](https://developers.openai.com/codex/skills)
- [Codex plugin build documentation](https://developers.openai.com/codex/plugins/build)
- [Agent Skills specification](https://agentskills.io/specification)
