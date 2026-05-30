---
id: "harness-as-a-service-assumption"
type: "assumption"
title: "Harness Provider Contracts Can Start Narrow"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Harness Provider Contracts Can Start Narrow

## Statement

The first Harness as a Service proof of concept can ship with Workbench-provided
default providers, static provider bindings, and explicit opt-outs, as long as
the artifact packet preserves a clean distinction between feature contracts,
provider bindings, and standalone service internals.

Deep provider override runtime is not required in the first implementation
slice. The architecture only needs to make replacement possible later without
hard-coding default services into the public MCP feature contract.

## Evidence

The current Workbench kernel already exposes MCP tools/resources backed by a
simple local file store. The related `backlog-service` repository describes a
local-first HTTP service backing Workbench `backlog.*` tools, which shows that a
feature can be useful through MCP while state and domain behavior live in a
separate service.

The `mcp-control-plane`, `mcp-data-plane`, and `mcp-remote-proxy` repositories
also support a narrow-first interpretation: catalog/lifecycle, routing, and
proxying are separable concerns, and none require the first harness slice to
solve all production gateway concerns.

The MCP specification supports this split because servers expose capabilities
through standard tools, resources, prompts, and notifications while the backing
implementation remains outside the protocol.

## Validation Plan

Validate the assumption during the initial implementation plan by producing one
manifest-backed feature projection and one provider binding before adding any
provider marketplace or hot-swap mechanism.

The assumption holds if a compatible agent can discover and invoke the default
feature surface, the user can disable a default feature or provider, and the
manifest does not require the backing service to share Workbench's internal
architecture. The assumption fails if the first feature cannot be implemented
without embedding a service-specific contract directly into harness core.

## Source References

- [Harness implementation plan](./harness-as-a-service-initial-implementation-plan.md)
- https://github.com/manuelibar/backlog-service
- https://github.com/manuelibar/mcp-control-plane
- https://github.com/manuelibar/mcp-data-plane
- https://github.com/manuelibar/mcp-remote-proxy
- https://modelcontextprotocol.io/specification/2025-11-25

## Open Questions

Open human nudges are tracked centrally in
[the RFC Human-in-the-loop Index](./harness-as-a-service-rfc.md#human-in-the-loop-index).
