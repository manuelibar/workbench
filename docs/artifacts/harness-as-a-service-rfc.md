---
id: "harness-as-a-service-rfc"
type: "rfc"
title: "Harness as a Service MCP Architecture"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Harness as a Service MCP Architecture

## Summary

Build Harness as a Service as Workbench's public MCP distribution architecture.
The harness exposes Workbench features as MCP tools, resources, resource
templates, and prompts so any compatible agent can use them. Durable behavior
and domain state live behind standalone services that communicate with the
harness through explicit provider bindings.

The first slice is personal-use first: local defaults, loopback services, static
configuration, and Workbench-provided providers. The architecture still records
provider identity, opt-outs, feature manifests, and extension boundaries so the
same surface can later become a shareable public MCP harness-as-a-service.

This RFC is the living drill hub for the epic. Action artifacts produced by this
RFC:

- [Harness MCP Feature Distribution Requirement](./harness-as-a-service-requirement.md)
- [Harness as a Service Initial Implementation Plan](./harness-as-a-service-initial-implementation-plan.md)

## Problem

Workbench `main` is now a small context/artifact MCP kernel. That is the right
foundation, but it does not answer how higher-level Workbench features return
without rebuilding a monolith.

The harness needs to solve a specific distribution problem: make Workbench
features available through MCP from any compatible agent while letting backing
services remain standalone, replaceable, and independently architected.

The problem is not only transport. Without a feature/provider contract, the
first personal defaults will become the architecture: tools will imply one
service, resource URIs will imply one store, and disabling or replacing a
bundled service will become a breaking change.

## Proposal

### Architecture

Introduce a harness MCP server as the public entrypoint for Workbench feature
distribution. It owns:

- feature catalog and feature manifests;
- MCP projection of enabled features;
- provider binding resolution;
- opt-out and disabled-state behavior;
- harness metadata resources;
- translation between MCP calls/resources and provider adapters.

It does not own every feature's domain model, storage, or process lifecycle.
Standalone backing services can use their own architecture as long as they
satisfy the provider contract selected for a feature slot.

### Feature Manifest

Each feature declares a manifest with:

- stable feature ID and version;
- display title and lifecycle status;
- MCP tools/resources/resource templates/prompts it may expose;
- provider slots it requires;
- default Workbench provider, if available;
- opt-out behavior;
- status and health resource shape.

The manifest is the public feature contract. It should avoid naming local
deployment details in tool names or resource URI semantics.

### Provider Binding

Each provider binding selects an implementation for one feature slot:

- `feature_id`
- `slot_id`
- `provider_id`
- `enabled`
- `transport`
- `endpoint` or local process/config reference;
- health/status metadata;
- later credential or policy references.

The first implementation can use static configuration and Workbench-provided
providers. It should still represent provider identity explicitly.

### MCP Projection

The harness projects enabled feature manifests into MCP capability discovery.
Tools invoke actions on providers. Resources expose feature state, provider
status, catalogs, and durable content. Prompts expose user-invoked workflows
when a feature owns prompt templates. Roots are used when clients can provide
workspace boundaries.

When a feature or provider is disabled, action tools should not be advertised as
available. The harness can still expose a status resource that explains the
disabled state and how to re-enable it.

### Defaults

Proof-of-concept defaults are local and personal:

- current context/artifact kernel remains available as a default Workbench
  feature surface;
- backlog can be backed by the local `backlog-service`;
- configuration can start with profile/env patterns consistent with
  `go-config` and `go-service-config-provider`;
- control-plane/data-plane/proxy concepts can inform later slices, but the
  first slice does not need production fleet management.

### Public Extension Story

Later public extension work should add installable feature packages and provider
adapters. Extensions should be able to declare MCP surface, provider slots, and
configuration needs without patching harness core.

The official MCP registry is relevant to public discovery later, but the
immediate deliverable is a Workbench feature/provider contract that can survive
being shared.

## Tradeoffs

Keeping the first slice personal-first reduces immediate authentication,
tenancy, policy, and remote-hosting complexity. The tradeoff is that public
harness-as-a-service claims must remain future-oriented until those concerns are
designed and implemented.

Designing provider configurability up front adds schema and catalog work before
multiple providers exist. The tradeoff is worth paying narrowly because opt-out
and provider identity are hard to retrofit without breaking tool/resource
contracts.

Using standalone services keeps Workbench modular and lets services evolve with
their own architecture. The tradeoff is more adapter and health/status plumbing
in the harness.

Exposing only enabled capabilities keeps agents from calling unwanted or broken
tools. The tradeoff is that users need clear status resources and list-changed
behavior so capability changes are explainable.

## Rollout

1. Land this docs-only bootstrap packet.
2. Drill the open human nudges in this RFC before coding provider override or
   public sharing behavior.
3. Implement the initial feature manifest and provider binding schema.
4. Add harness metadata resources and enabled/disabled capability projection.
5. Wire one useful default backed feature, preferably backlog through
   `backlog-service`.
6. Add tests and a local smoke path from a compatible MCP client.
7. Append decisions, implementation notes, and follow-up action artifacts to
   this RFC as the epic progresses.

Compatibility starts local-only. Public remote sharing, third-party provider
override runtime, auth, policy, marketplace workflows, and service hosting are
separate follow-up slices.

## Open Questions

### Human-in-the-loop Index

| ID | Nudge | Type | Why it matters | Blocks | Default if unanswered |
|---|---|---|---|---|---|
| HASS-HIL-001 | Confirm the first Workbench-provided default service set for the personal proof of concept. | decision | The first defaults determine which feature manifests and provider bindings are implemented first. | Initial runtime scope and smoke path. | Use context/artifacts plus local `backlog-service` as the first default set. |
| HASS-HIL-002 | Approve the local-only, no-remote-auth boundary for the first personal harness slice. | approval | It prevents the proof of concept from implying production-ready public hosting or remote access. | Any remote/public sharing claims in implementation. | Bind to local stdio or loopback-only services and document public sharing as future work. |
| HASS-HIL-003 | Choose how deep provider override support must go in the first implementation. | tradeoff | Deep override support costs time; no override shape risks lock-in. | Provider replacement runtime and extension adapter design. | Implement manifests, provider bindings, and opt-outs; defer hot-swap/provider SDK work. |
| HASS-HIL-004 | Challenge whether any default service should be folded into harness core instead of staying standalone. | challenge | Folding services into core may simplify early code but weakens the public service boundary. | Final service boundary for first backed feature. | Keep the harness thin and use standalone services behind provider adapters. |

These nudges are the only human nudges in the bootstrap packet.

## Source References

- [Harness charter](./harness-as-a-service-charter.md)
- [Problem statement](./harness-as-a-service-problem-statement.md)
- [Concept map](./harness-as-a-service-concept-map.md)
- [Provider assumption](./harness-as-a-service-assumption.md)
- [Harness defaults risk](./harness-as-a-service-risk.md)
- [Research note](./harness-as-a-service-research-note.md)
- [Distribution requirement](./harness-as-a-service-requirement.md)
- [Initial implementation plan](./harness-as-a-service-initial-implementation-plan.md)
- `README.md`
- `docs/how-to/epic-branch-workflow.md`
- `docs/reference/artifact-conventions.md`
- https://modelcontextprotocol.io/specification/2025-11-25
- https://modelcontextprotocol.io/specification/2025-11-25/server/tools
- https://modelcontextprotocol.io/specification/2025-11-25/server/resources
- https://modelcontextprotocol.io/specification/2025-11-25/server/prompts
- https://modelcontextprotocol.io/specification/2025-11-25/client/roots
- https://github.com/modelcontextprotocol/servers
- https://registry.modelcontextprotocol.io
- https://github.com/manuelibar/mcp-control-plane
- https://github.com/manuelibar/mcp-data-plane
- https://github.com/manuelibar/mcp-remote-proxy
- https://github.com/manuelibar/go-service-config-provider
- https://github.com/manuelibar/go-config
- https://github.com/manuelibar/backlog-service
