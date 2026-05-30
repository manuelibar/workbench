---
id: "agent-skill-distribution-concept-map"
type: "spec"
title: "Agent Skill Distribution Concept Map"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Agent Skill Distribution Concept Map

## Context

This artifact is a taxonomy and concept map for the epic. It gives later
implementation work a shared vocabulary before deciding concrete Go types,
resource URIs, tools, or persistence.

The key design pressure is that "skill" is not an MCP primitive. MCP supplies
resources, prompts, tools, and server capabilities. Codex supplies a skill
authoring format and plugin distribution model. Workbench needs a resource-first
distribution contract that can bridge those worlds.

## Design

### Core Taxonomy

| Concept | Control Plane | Primary Purpose | Workbench Treatment |
|---|---|---|---|
| Skill | Agent-selected workflow package | Reusable task procedure with metadata, instructions, references, optional scripts, and assets | Distributed as a package made visible through MCP resources |
| Skill manifest | Application-controlled context | Small metadata record for discovery, compatibility, provenance, install targets, and activation policy | First resource an agent should read after catalog match |
| Skill instructions | Agent-loaded context | Full `SKILL.md`-style guidance loaded after a skill is selected | Versioned text resource |
| Skill reference | Agent-loaded context | Supporting docs loaded on demand | Versioned text or binary resource |
| Skill asset | Agent-loaded context or install payload | Templates, examples, images, schemas, or data files | Versioned resource with MIME type and size |
| Skill script | Executable payload | Optional deterministic helper code | Resource until installed; never executed during discovery |
| Prompt | User-controlled MCP primitive | Explicitly selected template or workflow starter | Adjacent but not a skill package |
| Tool | Model-controlled MCP primitive | Action or retrieval function invoked through MCP | Used for install planning and mutations, not passive skill content |
| Service capability | Server protocol feature | Declares what the MCP server can list, read, call, subscribe to, or notify | Describes Workbench itself, not an individual skill |
| Codex plugin | Codex installable unit | Bundles skills, MCP config, apps, hooks, and presentation metadata | Supported adapter target, not the generic Workbench package model |
| Registry entry | Distribution metadata | Points to package or remote service locations and versions | Optional upstream/downstream source of skill package metadata |

### Boundary Rules

- A skill package is distributed as resources because its default state is
  passive context.
- A prompt can reference a skill or help invoke one, but a prompt is not the
  skill's canonical distribution unit.
- A tool can install, activate, deactivate, validate, or update a skill only
  after approval; a tool does not represent the skill's content.
- Service capabilities describe Workbench's protocol behavior, such as whether
  the resource list can change, not the permissions a skill may request.
- Codex plugin manifests are adapter metadata. A generic MCP agent should still
  be able to discover and inspect the skill resources without understanding
  Codex plugins.
- Scripts, hooks, bundled MCP server configs, and app connector mappings are
  capability requests. They require trust review before activation.

### Lifecycle Map

| Phase | Agent Action | User Boundary | Workbench Surface |
|---|---|---|---|
| Discover | List catalog resources or search resource templates | No approval for metadata-only reads | `resources/list`, `resources/templates/list`, catalog resource |
| Inspect | Read manifest, instructions summary, compatibility, source, and digest | No approval for read-only local metadata | `resources/read` on skill manifest and selected files |
| Plan | Ask Workbench for an install or activation plan | Approval needed before writes or enablement | Future tool returning a plan without mutating state |
| Install | Copy or materialize skill package into agent-specific target | Explicit approval | Future tool with dry-run, target path, digest verification |
| Activate | Enable the installed skill for implicit or explicit use | Explicit approval for activation policy and dependencies | Future tool plus agent-specific adapter |
| Use | Agent loads skill instructions and references progressively | Per-agent context policy | Resource reads or agent-local skill loading |
| Update | Compare versions and content digests | Approval before replacing installed content | Future tool with rollback path |
| Deactivate | Disable the skill without deleting package state | Approval may be lightweight | Future tool or generated agent-specific instructions |

## Interfaces

### Resource Families

The exact URI grammar is implementation work, but the concept map assumes these
families:

- `workbench-skill:///catalog` lists all visible skills as compact metadata.
- `workbench-skill:///skills/{skill_id}/manifest` exposes the full manifest for
  the currently selected version or channel.
- `workbench-skill:///skills/{skill_id}/versions/{version}/manifest` exposes an
  immutable versioned manifest.
- `workbench-skill:///skills/{skill_id}/versions/{version}/instructions` exposes
  the primary `SKILL.md`-style instructions.
- `workbench-skill:///skills/{skill_id}/versions/{version}/files/{path}` exposes
  references, assets, scripts, and other package files.
- `workbench-skill:///skills/{skill_id}/install-targets/{agent_profile}` exposes
  read-only install guidance for an agent family when mutation tools are not
  available.

### Manifest Vocabulary

A skill manifest should include these fields or equivalents:

- identity: `id`, `name`, `title`, `description`, `publisher`, `homepage`,
  `repository`;
- versioning: `version`, `content_digest`, `created`, `updated`, `source_ref`,
  `channel`;
- package shape: `instructions`, `references`, `assets`, `scripts`,
  `manifest_format`;
- compatibility: `skill_spec_version`, `mcp_protocol_revisions`,
  `agent_profiles`, `codex_plugin_compatible`, `required_features`;
- safety: `license`, `trust_level`, `allowed_tools`, `requested_tools`,
  `requested_mcp_servers`, `network_access`, `filesystem_writes`,
  `script_execution`;
- activation: `implicit_invocation`, `explicit_invocation`, `default_enabled`,
  `activation_notes`;
- install: `install_targets`, `requires_approval`, `post_install_steps`,
  `rollback_hint`.

### Resource Metadata

Workbench should use MCP resource metadata deliberately:

- `name` is stable and machine-readable within the skill namespace.
- `title` is human-readable and may change without changing identity.
- `description` is a compact discovery hint, not the full instructions.
- `mimeType` distinguishes Markdown instructions, JSON manifests, scripts,
  images, archives, and binary assets.
- `size` helps hosts estimate context cost before reading.
- annotations can indicate intended audience, priority, and last modification
  time.

## Edge Cases

- Two packages publish the same skill `name`; Workbench must disambiguate by
  stable package `id` and publisher namespace.
- A Codex-compatible skill uses metadata that a generic MCP host does not
  understand; Workbench should preserve unknown metadata and expose generic
  compatibility summaries.
- A resource list changes while an agent is planning installation; Workbench
  should include version and digest in the plan so the apply step can detect
  drift.
- A skill includes scripts but no declared execution policy; default behavior is
  to expose scripts as files and deny execution until approved.
- A package depends on an MCP server, tool, or app connector that is unavailable
  to the current host; discovery remains possible, but activation is blocked or
  degraded.
- A publisher updates metadata without changing content; Workbench should keep
  immutable content identity separate from human-facing version labels.
- A host supports resources but not resource templates or subscriptions; the
  catalog must remain useful through plain `resources/list` and `resources/read`.
- A malicious skill tries to include hidden prompt-injection instructions in
  references or assets; Workbench should treat skill content as untrusted
  context and make provenance visible.

## Test Plan

Later implementation should verify:

- catalog resources list and read correctly with pagination-safe behavior;
- manifest resources include the required metadata vocabulary and valid MIME
  types;
- versioned resource URIs are stable for immutable content;
- list-changed notifications fire when the catalog changes, if the resources
  capability advertises them;
- install planning returns a dry-run plan that does not mutate filesystem or
  agent config;
- install apply rejects digest drift between plan and apply;
- Codex-compatible fixtures map to Codex skill/plugin locations while generic
  MCP fixtures can still inspect all resource content;
- tools that perform mutation require explicit approval metadata and never run
  scripts during discovery.

## Source References

- [MCP 2025-11-25 server feature overview](https://modelcontextprotocol.io/specification/2025-11-25/server/index)
- [MCP resources](https://modelcontextprotocol.io/specification/2025-11-25/server/resources)
- [MCP prompts](https://modelcontextprotocol.io/specification/2025-11-25/server/prompts)
- [MCP tools](https://modelcontextprotocol.io/specification/2025-11-25/server/tools)
- [Codex Agent Skills](https://developers.openai.com/codex/skills)
- [Codex plugin build documentation](https://developers.openai.com/codex/plugins/build)
- [Agent Skills specification](https://agentskills.io/specification)

## Open Questions

Open human decisions are tracked in the RFC's Human-in-the-loop Index.
