---
id: "harness-as-a-service-research-note"
type: "research_note"
title: "Harness as a Service Research Note"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Harness as a Service Research Note

## Question

What current repository and MCP evidence should shape the Harness as a Service
epic so Workbench can distribute features through MCP while delegating behavior
to standalone services?

## Sources

- `README.md`: current Workbench is a local stdio MCP context/artifact kernel
  with file-backed artifacts and deterministic capability visibility.
- `docs/how-to/epic-branch-workflow.md`: epic packets must be self-contained,
  use current docs and targeted web research, and use the RFC as the drill hub.
- `docs/reference/artifact-conventions.md`: artifacts require deterministic
  frontmatter, non-placeholder required sections, and an RFC human nudge index.
- https://modelcontextprotocol.io/specification/2025-11-25: current MCP
  specification; defines hosts, clients, servers, JSON-RPC, capability
  negotiation, server features, client roots, and security principles.
- https://modelcontextprotocol.io/specification/2025-11-25/server/tools:
  authoritative tool behavior, discovery, tool call shape, list-changed
  notification, tool-name guidance, structured output, and safety expectations.
- https://modelcontextprotocol.io/specification/2025-11-25/server/resources:
  authoritative resource behavior, URI identity, templates, subscriptions,
  annotations, and list-changed notification.
- https://modelcontextprotocol.io/specification/2025-11-25/server/prompts:
  authoritative prompt behavior, user-controlled prompt discovery, and
  prompt-list capability.
- https://modelcontextprotocol.io/specification/2025-11-25/client/roots:
  authoritative roots behavior for client-exposed filesystem/workspace
  boundaries.
- https://github.com/modelcontextprotocol/servers: official reference server
  repository; useful as an example set, but its README explicitly distinguishes
  reference implementations from production-ready solutions.
- https://registry.modelcontextprotocol.io: official registry for discovering
  MCP servers, relevant to the later public extension story.
- https://github.com/manuelibar/mcp-control-plane: related catalog/lifecycle
  service for MCP server registration, configuration, lifecycle state, HTTP
  administration, and optional MCP-tool adapter.
- https://github.com/manuelibar/mcp-data-plane: related service for routing
  tool-call traffic across upstream servers and namespaces.
- https://github.com/manuelibar/mcp-remote-proxy: related proxy that can expose
  a remote MCP server locally and filter mirrored tool visibility.
- https://github.com/manuelibar/go-service-config-provider: related
  configuration provider toolkit with reload, fallback, hooks, and no global
  side effects.
- https://github.com/manuelibar/go-config: related layered profile loader with
  filesystem abstraction and command-line override support.
- https://github.com/manuelibar/backlog-service: related local-first HTTP
  backlog service backing Workbench `backlog.*` tools with localhost/no-auth v1
  defaults, Postgres storage, optimistic concurrency, audit headers, and events.

## Findings

MCP is a strong fit for Workbench feature distribution because it standardizes
capability discovery and invocation without standardizing how a server stores
data or implements domain behavior. Tools are model-controlled actions,
resources are application-driven context/data, prompts are user-controlled
workflow templates, and roots define filesystem/workspace boundaries when the
client supports them.

The MCP specification also makes security and user control part of the design
surface: tools represent arbitrary execution paths, clients should show and
confirm sensitive tool use, servers must validate inputs and protect access, and
hosts must preserve user control over exposed data. A public harness cannot
treat provider replacement or extension as purely mechanical plumbing.

Current Workbench `main` is deliberately a kernel, not a full app. That supports
building the harness as a new public distribution layer instead of expanding the
kernel into a monolith.

The related service repositories show a consistent local-first decomposition:
catalog/lifecycle, data-plane routing, remote proxying, configuration loading,
dynamic provider composition, and backlog domain behavior are each expressed as
small services or libraries. The harness can reuse those shapes conceptually
without forcing their internal architecture into the MCP contract.

The official MCP reference server repository is useful for examples, but it
warns that reference servers demonstrate SDK and protocol usage rather than
production-ready safeguards. Workbench should therefore use official reference
servers for protocol orientation, not as proof that a public harness is
production-ready by default.

## Implications

The harness should be a thin MCP facade plus a feature/provider catalog, not a
new all-in-one Workbench service. It should expose feature contracts through MCP
and communicate with standalone backing services that can follow their own
architecture.

The personal proof of concept should default to local services, loopback
addresses, and explicit configuration. That is consistent with current
Workbench and `backlog-service`, and it avoids pretending that public remote
hosting, authentication, policy, and marketplace concerns are already solved.

Provider replacement should be designed into manifests and bindings now, but
deep runtime override behavior can be deferred. The first critical override
behavior is opt-out: if a feature or provider is disabled, the harness should
remove or clearly mark its MCP surface so compatible agents do not see broken or
unwanted tools.

Public extension work should model feature packages and provider adapters
before marketplace distribution. The official MCP registry is relevant later,
but the immediate need is a stable Workbench feature manifest and provider
binding contract.

## Source References

- [Harness RFC](./harness-as-a-service-rfc.md)
- [Harness concept map](./harness-as-a-service-concept-map.md)
- [Initial implementation plan](./harness-as-a-service-initial-implementation-plan.md)

## Open Questions

Open human nudges are tracked centrally in
[the RFC Human-in-the-loop Index](./harness-as-a-service-rfc.md#human-in-the-loop-index).
