---
id: "harness-as-a-service-charter"
type: "charter"
title: "Harness as a Service Charter"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Harness as a Service Charter

## Mission

Define Workbench's public MCP harness architecture so Workbench features can be
distributed through MCP and used from any compatible agent while their stateful
or operational behavior lives in standalone backing services.

The harness starts as a personal-use distribution layer: a local MCP entrypoint
that exposes Workbench-provided feature defaults, routes tool/resource calls to
local services, and records enough provider metadata to make future replacement,
extension, and opt-out work additive instead of invasive. The later public shape
is a shareable MCP harness-as-a-service that lets other agents and users install
the same feature surface with their own service providers.

## Scope

In scope:

- MCP as the primary distribution surface for Workbench features.
- Feature access from any MCP-compatible agent, not only one Workbench-owned
  client.
- A thin harness MCP server that exposes tools, resources, prompts, and feature
  metadata while delegating durable behavior to standalone services.
- Default Workbench-provided services for the proof of concept, including local
  artifact/context behavior and a local backlog backing service.
- A provider catalog shape that supports enable/disable flags, endpoint
  configuration, provider identity, capability ownership, and future provider
  replacement.
- Opt-out architecture for users who do not want a bundled feature or backing
  service.
- Public extension story for third-party feature bundles and provider adapters.
- Personal-first defaults: local process boundaries, loopback endpoints,
  explicit configuration, and minimal operational assumptions.

Out of scope for this bootstrap:

- Deep provider override runtime, hot-swapping, marketplace governance, billing,
  multi-tenant hosting, remote authentication, or production fleet management.
- Porting every historical Workbench subsystem into the harness.
- Treating any standalone service as forced to share Workbench's internal
  package layout, storage engine, release cadence, or deployment topology.

## Stakeholders

- Branch owner for `epic/harness-as-a-service`: owns the architecture packet,
  RFC drill-down, and initial proof-of-concept sequencing.
- Workbench users running local agents: need a dependable personal harness with
  useful defaults and visible opt-outs.
- Compatible MCP agents: consume the harness through standard MCP capability
  discovery, tool calls, resource reads, and prompt listing.
- Default service owners: own backing service contracts, storage, operations,
  and evolution behind the MCP facade.
- Future extension authors: publish feature bundles, provider adapters, and
  replacement services without forking Workbench core.

## Success Criteria

The epic is successful when a local user can run a Workbench harness MCP server,
connect it to at least one compatible agent, discover a coherent Workbench
feature surface, and use default backing services through MCP tools/resources.

The architecture is successful when each feature declares its MCP surface and
provider binding explicitly, bundled services can be disabled or replaced by
configuration, standalone services remain free to follow their own internal
architecture, and the RFC contains the open human decisions needed before the
personal proof of concept becomes a shareable public MCP harness.

## Source References

- [Harness RFC](./harness-as-a-service-rfc.md)
- [Harness concept map](./harness-as-a-service-concept-map.md)
- `README.md`
- `docs/how-to/epic-branch-workflow.md`
- `docs/reference/artifact-conventions.md`
- https://modelcontextprotocol.io/specification/2025-11-25

## Open Questions

Open human nudges are tracked centrally in
[the RFC Human-in-the-loop Index](./harness-as-a-service-rfc.md#human-in-the-loop-index).
