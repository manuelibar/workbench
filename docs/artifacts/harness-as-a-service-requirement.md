---
id: "harness-as-a-service-requirement"
type: "requirement"
title: "Harness MCP Feature Distribution Requirement"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Harness MCP Feature Distribution Requirement

## Statement

The Harness as a Service implementation must expose Workbench features through a
stable MCP distribution contract that separates feature identity and MCP
projection from the default provider implementation behind each feature.

Every harness-managed feature must declare at least:

- feature ID and version;
- public MCP tools, resources, resource templates, and prompts it may expose;
- provider slots required by the feature;
- Workbench-provided default provider, when one exists;
- opt-out behavior for disabled feature/provider states;
- source/service status resource visible to the agent and user.

## Rationale

Workbench needs feature access from any compatible MCP agent while preserving
the ability to keep feature behavior in standalone backing services. If the MCP
surface is coupled directly to bundled service internals, the personal proof of
concept will work only by freezing current defaults into public architecture.

This requirement keeps the immediate implementation narrow but structurally
ready for later provider replacement, public extension packages, and shareable
harness-as-a-service distribution.

## Acceptance Criteria

- A compatible MCP client can list harness-managed tools/resources without
  knowing which default service implements them.
- At least one default feature routes an MCP tool call to a standalone backing
  service through an explicit provider binding.
- At least one default feature exposes a Workbench resource URI that reports
  feature/provider status.
- Disabling a feature or provider changes MCP discovery so unwanted action tools
  are not advertised as available.
- Feature/provider metadata is represented in docs and code without requiring
  deep runtime provider override in the first slice.
- Action artifacts and implementation work link back to
  [the harness RFC](./harness-as-a-service-rfc.md).

## Source References

- [Harness RFC](./harness-as-a-service-rfc.md)
- [Harness concept map](./harness-as-a-service-concept-map.md)
- https://modelcontextprotocol.io/specification/2025-11-25/server/tools
- https://modelcontextprotocol.io/specification/2025-11-25/server/resources

## Open Questions

Open human nudges are tracked centrally in
[the RFC Human-in-the-loop Index](./harness-as-a-service-rfc.md#human-in-the-loop-index).
