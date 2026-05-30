---
id: "harness-as-a-service-initial-implementation-plan"
type: "implementation_plan"
title: "Harness as a Service Initial Implementation Plan"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Harness as a Service Initial Implementation Plan

## Objective

Implement the first personal-use Harness as a Service slice: a local MCP
distribution layer that projects Workbench feature manifests into MCP
tools/resources and routes at least one feature to a standalone default backing
service through an explicit provider binding.

This plan is produced by [the harness RFC](./harness-as-a-service-rfc.md). It
does not require deep provider override runtime before the first useful proof of
concept.

## Steps

1. Define a feature manifest schema for feature ID, version, MCP projection,
   provider slots, default provider, and opt-out behavior.
2. Define a provider binding schema for provider ID, feature slot, enabled
   state, transport, endpoint/config reference, and health/status metadata.
3. Add a harness catalog that can read static default manifests and static local
   provider bindings.
4. Project enabled feature manifests into MCP tool/resource visibility using the
   existing Workbench capability planning style as the local guide.
5. Add harness metadata resources such as `workbench:///features`,
   `workbench:///features/{id}`, `workbench:///providers`, and
   `workbench:///providers/{id}`.
6. Wire one default standalone backing service, preferably the local
   `backlog-service`, behind a small set of MCP tools such as create/list/claim.
7. Add explicit opt-out behavior so a disabled feature or provider removes
   action tools from discovery and leaves a status resource explaining why.
8. Add focused tests for manifest parsing, provider binding, enabled/disabled
   capability projection, service-unavailable tool errors, and status resources.
9. Document the personal-first run path, local service assumptions, and deferred
   provider override work in the RFC append-only notes.

## Verification

- `go test ./...` passes after runtime work begins.
- Manifest and provider binding tests cover at least one enabled default
  feature and one disabled feature.
- MCP integration tests show that a compatible client can discover harness
  feature resources and only the tools for enabled providers.
- A manual smoke path connects an MCP client to the harness, invokes the default
  backed feature, then disables it and verifies the action tools disappear.
- Documentation cross-links remain valid and the RFC Human-in-the-loop Index
  contains every human nudge in the packet.

## Rollback

The first runtime slice should be independently revertable. If the harness
projection destabilizes the current context/artifact kernel, revert the harness
runtime commit and keep this docs packet as the source for a smaller follow-up
slice.

For configuration or manifest experiments, prefer additive files and feature
flags so users can return to the current `main` kernel behavior by disabling
harness-managed features rather than changing existing context/artifact tools.

## Source References

- [Harness RFC](./harness-as-a-service-rfc.md)
- [Distribution requirement](./harness-as-a-service-requirement.md)
- [Harness concept map](./harness-as-a-service-concept-map.md)
- https://github.com/manuelibar/backlog-service
- https://github.com/manuelibar/go-config
- https://github.com/manuelibar/go-service-config-provider

## Open Questions

Open human nudges are tracked centrally in
[the RFC Human-in-the-loop Index](./harness-as-a-service-rfc.md#human-in-the-loop-index).
