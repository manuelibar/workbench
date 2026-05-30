---
id: "role-mgmt-assumption"
type: "assumption"
title: "Roles Are Local Workbench Policy"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Roles Are Local Workbench Policy

## Statement

Role management should treat roles as local Workbench policy objects that
compile into context, instructions, capability filters, and inspectable
resources. Roles should not be modeled as a new MCP protocol feature.

## Evidence

The current Workbench docs define `context` and artifacts as the foundation and
explicitly defer roles to an epic branch. MCP's current server feature model is
resources, prompts, and tools, with roots, sampling, and elicitation on the
client side. The latest MCP specification does not define a role primitive, and
the official Go SDK exposes server behavior through those MCP concepts.

## Validation Plan

Validate the assumption during implementation by proving that role behavior can
be represented through:

- A Workbench context field for selected role state.
- Resources for role inspection and diagnostics.
- Tools or context calls for role selection and validation.
- Prompts only where user-invoked templates are useful.
- Capability planning that remains explicit and testable without protocol
  changes.

If implementation requires protocol behavior outside MCP primitives, record the
gap in the RFC before adding code.

## Source References

- `README.md`
- `docs/reference/context-contract.md`
- [MCP specification overview](https://modelcontextprotocol.io/specification/2025-11-25)
- [MCP server feature overview](https://modelcontextprotocol.io/specification/2025-11-25/server/index)
- [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk)

## Open Questions

No assumption-specific human nudges. Packet-level nudges are indexed in
[Role Management RFC](role-mgmt-rfc.md#human-in-the-loop-index).
