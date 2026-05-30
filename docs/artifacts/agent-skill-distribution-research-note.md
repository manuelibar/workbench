---
id: "agent-skill-distribution-research-note"
type: "research_note"
title: "Agent Skill Distribution Research Note"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Agent Skill Distribution Research Note

## Question

What current Workbench, MCP, Codex, and open skill-package conventions should
shape a Workbench epic for distributing agent skills as MCP resources?

## Sources

- `README.md`: current Workbench foundation, stdio MCP shape, available tools,
  resources, and artifact kernel scope.
- `docs/how-to/epic-branch-workflow.md`: packet requirements, source hierarchy,
  RFC hub expectations, and archive-branch exclusion.
- `docs/reference/artifact-conventions.md`: frontmatter, artifact contract
  types, required sections, and Human-in-the-loop Index requirement.
- [MCP 2025-11-25 specification](https://modelcontextprotocol.io/specification/2025-11-25):
  current protocol overview, security principles, host/client/server roles, and
  feature categories.
- [MCP 2025-11-25 server feature overview](https://modelcontextprotocol.io/specification/2025-11-25/server/index):
  control hierarchy for prompts, resources, and tools.
- [MCP resources](https://modelcontextprotocol.io/specification/2025-11-25/server/resources):
  resource listing, reading, templates, list-changed notifications,
  subscriptions, URI schemes, annotations, and security considerations.
- [MCP prompts](https://modelcontextprotocol.io/specification/2025-11-25/server/prompts):
  prompt discovery, retrieval, and user-controlled interaction model.
- [MCP tools](https://modelcontextprotocol.io/specification/2025-11-25/server/tools):
  model-controlled tool invocation, structured outputs, error handling, and
  security guidance.
- [MCP registry overview](https://modelcontextprotocol.io/registry/about):
  public metadata registry role, namespace verification, server metadata, and
  relationship to package registries and aggregators.
- [MCP registry versioning](https://modelcontextprotocol.io/registry/versioning):
  version uniqueness, semantic version recommendations, package/API alignment,
  and aggregator comparison guidance.
- [MCP registry package types](https://modelcontextprotocol.io/registry/package-types):
  supported package metadata patterns and integrity expectations, including
  MCPB hash metadata.
- [Codex Agent Skills](https://developers.openai.com/codex/skills):
  skill structure, progressive disclosure, install locations, optional metadata,
  dependencies, and activation policy.
- [Codex plugin build documentation](https://developers.openai.com/codex/plugins/build):
  plugin manifest, marketplace, plugin structure, bundled skills, MCP servers,
  apps, hooks, and Codex distribution boundaries.
- [Agent Skills specification](https://agentskills.io/specification):
  open skill metadata, naming constraints, optional directories, references,
  scripts, assets, and progressive disclosure guidance.
- GitHub current-public scan: `https://api.github.com/users/manuelibar/repos?per_page=100&sort=updated`
  plus exact repository-name searches for `go-cli-assets`,
  `ai-assets-api-go`, and `ai-asset-search-api`. The scan did not resolve those
  names as current public source evidence for this packet, so no contract
  derives from them.

## Findings

Workbench `main` is intentionally small: a local stdio MCP server with context
and file-backed artifacts. Its current resource surface already treats
artifacts as readable resources, which is compatible with a resource-first
skill catalog.

MCP's current 2025-11-25 specification separates primitives by control model:
prompts are user-controlled templates, resources are application-controlled
context, and tools are model-controlled functions. Skill packages are mostly
context until installation or activation mutates local agent state, so resources
are the least surprising primitive for discovery and inspection.

MCP resources support list, read, resource templates, optional list-changed
notifications, optional subscriptions, MIME types, sizes, and annotations. Those
features map well to a skill catalog, versioned manifests, instructions,
references, assets, and scripts.

MCP tools have safety expectations around validation, access control,
confirmation for sensitive operations, user-visible inputs, timeouts, and audit.
That makes tools appropriate for install planning and state mutation, but not
for exposing passive package contents.

MCP prompts can include structured messages and embedded resources, but their
interaction model is user selection. They are useful for starter workflows that
reference skills, not for canonical skill distribution.

The MCP Registry is currently a preview metadata repository for publicly
accessible MCP servers, not a package registry and not a skill registry. Its
versioning guidance is still useful: immutable publication versions, semantic
version preference, package/API version alignment, and aggregator comparison
rules.

Codex documentation distinguishes the authoring format from the distribution
unit. Skills are `SKILL.md` directories with optional scripts, references,
assets, and metadata. Plugins are the Codex installable unit for sharing skills
across teams or bundling skills with MCP config, apps, hooks, and presentation
assets.

Codex and the Agent Skills specification both emphasize progressive disclosure:
start with `name` and `description`, load full instructions only when selected,
and load references/assets/scripts on demand. Workbench skill resources should
preserve that shape and avoid loading entire packages into initial context.

Codex optional metadata can declare UI presentation, implicit invocation policy,
and tool dependencies. Workbench should preserve that metadata where present,
but should expose a generic compatibility summary for non-Codex hosts.

The named user public repos were not available as current public evidence
through the GitHub searches performed. They should not be cited as design
authority until a resolvable current source URL is provided.

## Implications

The RFC should propose a resource-first distribution model with future mutation
tools for install and activation. Resource URIs should be stable and versioned,
with a compact catalog for discovery and full manifests for inspection.

The skill manifest should combine open skill metadata, MCP resource metadata,
Codex compatibility data when present, version/digest provenance, requested
capabilities, and install targets.

Codex plugin support should be an adapter path. It is important for
Codex-compatible agents, but making plugin manifests canonical would exclude
other MCP-compatible hosts and overfit Workbench to one distribution surface.

Security design must treat skill packages as untrusted content. Reading a skill
resource must never execute a script, enable a tool, add an MCP server, install
a connector, or write to an agent config.

Versioning should pair publisher-facing versions with immutable content digests.
Install and update plans should include both so an apply step can detect drift.

The first action artifact should sequence implementation from read-only catalog
resources to install planning, then approved install/activation tools.

## Source References

- [agent-skill-distribution-rfc.md](agent-skill-distribution-rfc.md)
- [agent-skill-distribution-concept-map.md](agent-skill-distribution-concept-map.md)

## Open Questions

Open human decisions are tracked in the RFC's Human-in-the-loop Index.
