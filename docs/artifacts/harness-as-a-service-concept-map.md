---
id: "harness-as-a-service-concept-map"
type: "spec"
title: "Harness as a Service Concept Map"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Harness as a Service Concept Map

## Context

The harness is the public MCP distribution layer for Workbench features. It does
not own every feature's storage or service runtime. Instead, it owns the feature
catalog, MCP capability projection, provider binding, and user-visible
configuration story.

Current Workbench `main` already proves the smallest version of this pattern:
MCP exposes `context`, artifact tools, and artifact resources, while artifact
state is file-backed under `docs/artifacts/`. The harness generalizes that
pattern to more Workbench features and to standalone services such as a local
backlog HTTP service.

## Design

### Taxonomy

| Concept | Meaning | Owned by |
|---|---|---|
| Harness MCP server | The MCP entrypoint agents connect to. It distributes Workbench feature capabilities and routes requests. | Harness epic |
| Feature | A user-facing Workbench capability such as artifacts, backlog, notes, knowledge, or sessions. | Feature epic |
| Feature package | Versioned declaration of a feature's MCP tools, resources, prompts, provider requirements, defaults, and opt-outs. | Feature owner |
| MCP projection | The concrete tools, resources, resource templates, prompts, and capability notifications exposed for a feature. | Harness plus feature owner |
| Provider | A backing implementation satisfying a feature contract. It can be Workbench-provided or third-party. | Provider owner |
| Default service | A Workbench-provided provider intended to make the personal proof of concept useful out of the box. | Workbench |
| Standalone backing service | A service behind MCP tools/resources that may use its own architecture, storage, transport, and release cadence. | Service owner |
| Provider binding | Configuration that maps a feature to a provider, endpoint, transport, credentials placeholder, and enabled state. | User/harness |
| Opt-out | A user-controlled setting that disables a feature or provider and removes its MCP surface from discovery. | User/harness |
| Extension | A third-party feature package or provider adapter that can be installed without changing harness core. | Extension author |
| Personal harness | Local-first runtime with explicit configuration, loopback defaults, and minimal remote assumptions. | User |
| Public harness-as-a-service | Later shareable packaging and service model for distributing the same MCP feature architecture beyond one machine. | Future epic work |

### Capability Roles

| MCP primitive | Harness use |
|---|---|
| Tools | Invoke actions on backing services, such as creating backlog items or updating artifacts. |
| Resources | Expose durable state, status, catalogs, artifact content, feature manifests, and service health summaries. |
| Resource templates | Let agents read parameterized entities such as `workbench:///features/{id}` or `workbench:///providers/{id}`. |
| Prompts | Offer user-invoked workflows such as triage, planning, or review prompts once a feature owns them. |
| Roots | Bound service access to user-approved filesystem/workspace locations when clients support roots. |

### Provider Layers

1. `feature manifest`: static declaration of feature ID, version, MCP
   projection, default provider, provider slots, and disabled behavior.
2. `provider binding`: local configuration selecting a provider and endpoint.
3. `service adapter`: harness code that translates MCP calls/resources to the
   provider transport and response shape.
4. `standalone service`: process, library, or remote endpoint that owns the
   actual domain behavior.
5. `agent surface`: tools/resources/prompts visible to a compatible MCP agent.

The personal proof of concept should implement this stack narrowly. It can use
Workbench-owned defaults and static configuration, but the manifest and binding
shapes should already distinguish "default provider" from "feature contract".

## Interfaces

Initial conceptual interfaces:

```yaml
feature:
  id: backlog
  title: Backlog
  status: default_enabled
  mcp:
    tools:
      - backlog.issue.create
      - backlog.issue.list
      - backlog.issue.claim_next
    resources:
      - workbench:///features/backlog
    resource_templates:
      - workbench:///backlog/issues/{id}
  provider_slots:
    - id: backlog.store
      required: true
      default_provider: workbench.backlog_service
  opt_out:
    disables_tools_with_prefix: backlog.
    disables_resources_with_prefix: workbench:///backlog/
```

```yaml
provider_binding:
  feature_id: backlog
  slot_id: backlog.store
  provider_id: workbench.backlog_service
  enabled: true
  transport: http
  endpoint: http://127.0.0.1:7778
  config_ref: local-profile
```

Resource URI conventions should prefer Workbench-owned schemes for harness
metadata and feature state:

- `workbench:///features`
- `workbench:///features/{id}`
- `workbench:///providers`
- `workbench:///providers/{id}`
- `workbench:///services/{id}/health`
- feature-owned URI families such as `workbench:///backlog/issues/{id}`

Tool naming should stay stable, unique within the harness server, and compatible
with the MCP recommendation to use ASCII letters, digits, underscore, hyphen,
and dot.

## Edge Cases

- A provider is disabled: the harness removes its tools/resources from MCP
  discovery and exposes an explanatory feature status resource.
- A provider is configured but unreachable: list operations remain available for
  feature/provider status; action tools return tool execution errors that are
  useful to the agent and user.
- Two extensions define the same tool name: the harness rejects the second
  package unless the user explicitly remaps or namespaces it.
- A standalone service exposes extra capabilities: the harness only projects the
  feature contract unless an extension explicitly declares additional MCP
  surface.
- A client does not support roots: root-scoped features fall back to explicit
  configured paths or stay disabled.
- A remote provider is selected during the personal proof of concept: the
  binding may record the intent, but remote auth, tenancy, and policy are not
  treated as complete until later work.
- Provider replacement is requested before override runtime exists: the harness
  supports manifest/binding representation and opt-out, while actual runtime
  replacement is deferred to a follow-up implementation slice.

## Test Plan

Validate the concept map by using it to derive the first harness RFC drill
tasks:

- Check that each default feature can declare an MCP projection without naming a
  concrete backing service in its public tool contract.
- Check that each default provider can be disabled by configuration and that the
  disabled state removes the relevant MCP surface from discovery.
- Check that at least one tool-backed feature and one resource-backed feature
  can be described using the manifest and provider binding shapes.
- Check that standalone services can be represented by endpoint and transport
  metadata without importing their internal architecture into the harness.
- Check that human nudges in this packet are indexed by the RFC before runtime
  work begins.

## Source References

- [Harness RFC](./harness-as-a-service-rfc.md)
- [Distribution requirement](./harness-as-a-service-requirement.md)
- https://modelcontextprotocol.io/specification/2025-11-25/server/tools
- https://modelcontextprotocol.io/specification/2025-11-25/server/resources
- https://modelcontextprotocol.io/specification/2025-11-25/server/prompts
- https://modelcontextprotocol.io/specification/2025-11-25/client/roots
- https://github.com/manuelibar/mcp-control-plane
- https://github.com/manuelibar/mcp-data-plane
- https://github.com/manuelibar/mcp-remote-proxy

## Open Questions

Open human nudges are tracked centrally in
[the RFC Human-in-the-loop Index](./harness-as-a-service-rfc.md#human-in-the-loop-index).
