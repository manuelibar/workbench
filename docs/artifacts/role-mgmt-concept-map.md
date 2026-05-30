---
id: "role-mgmt-concept-map"
type: "spec"
title: "Role Management Concept Map"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Role Management Concept Map

## Context

Role management is a Workbench policy layer above the current context and
artifact kernel. It should define how named roles become selected context,
instruction fragments, capability policy, and inspectable resources while
remaining below user, session, harness, and security boundaries.

The concept map treats "role" as a local Workbench domain object, not as an MCP
primitive. MCP remains the transport surface for observing role state and its
effects.

## Design

### Taxonomy

| Concept | Meaning | Owned by role management |
|---|---|---|
| `RoleDefinition` | Stable definition for a role. | ID, title, summary, instruction fragments, capability policy, source, status, validation metadata. |
| `RoleSource` | Where a role definition came from. | Built-in, project custom, future user-global, or harness-provided. |
| `RoleSelection` | The currently active role ID, if any. | Context field semantics, validation, clearing, and context document rendering. |
| `InstructionProfile` | Role-owned guidance that shapes model behavior. | Ordered fragments such as stance, priorities, output style, and domain cautions. |
| `CapabilityPolicy` | Role-owned rule set for capability visibility. | Filters, annotations, and role-specific defaults over the base Workbench capability catalog. |
| `RoleStore` | Deterministic loader for custom definitions. | Storage roots, schema validation, conflict behavior, diagnostics. |
| `PrecedenceBoundary` | Rule that decides which instruction source wins in a conflict. | Matrix covering role, skill, memory, session, harness, artifact, and user sources. |
| `RoleDiagnostic` | Explainable output for debugging role effects. | Selected role, loaded source, rejected definitions, and capability changes. |

### Instruction Shaping

Instruction shaping should assemble a role-aware context document from ordered
inputs:

1. Harness and host configuration constraints.
2. Current user/session directives.
3. Workbench context fields such as `focus`, `artifact_id`, and future
   `role_id`.
4. Selected artifact guidance.
5. Selected role instruction profile.
6. Skill-specific procedure once a skill is explicitly invoked.
7. Passive memory or knowledge snippets retrieved for the task.

The selected role may provide defaults and constraints for the Workbench layer,
but it must not rewrite harness rules, user requests, skill execution
contracts, or safety requirements.

### Capability Mapping

The role capability policy should start from the base Workbench capability plan
for the current context. Role policy may:

- Hide capabilities that are irrelevant or unsafe for the selected role.
- Annotate active capabilities with role-specific rationale in diagnostics.
- Prefer role-appropriate prompts or resources when several equivalent choices
  exist.

Role policy must not expose a capability that the base context, harness
configuration, or permission state did not already make available.

### Custom Role Storage

Custom role storage should be file-backed and deterministic:

- Load from explicit roots in a stable order.
- Validate each file against a role schema before accepting it.
- Reject duplicate custom IDs unless the storage contract defines a specific
  override source.
- Reject any custom role that attempts to override a built-in role ID unless a
  later approved decision permits it.
- Report invalid definitions without changing the active role.

## Interfaces

Potential implementation interfaces to drill from this map:

- Extend `context` with a tri-state `role_id` field matching current
  `focus`/`artifact_id` semantics.
- Render selected role state in `workbench:///context` and in the
  `context` tool result.
- Add role inspection capabilities such as `role.list`, `role.get`,
  `role.select`, and `role.validate`, or equivalent context/artifact-oriented
  affordances if the RFC narrows the surface.
- Add resources such as `workbench:///roles` and
  `workbench:///roles/{id}` for passive inspection.
- Add a role-aware capability planner that takes `ContextState`,
  `CapabilityCatalog`, and `RoleDefinition` as explicit inputs.
- Add diagnostics to explain why a capability is visible, hidden, or unchanged
  under the selected role.

## Edge Cases

- The selected role ID no longer exists after custom role storage changes.
- A custom role file is syntactically valid Markdown or YAML but semantically
  invalid against the role schema.
- A custom role uses the same ID as a built-in role.
- A role selection changes capabilities while a client has not yet observed
  list-changed notifications.
- Role instructions conflict with explicit user instructions or a skill's
  required workflow.
- Memory content retrieved for a task claims to override the current role.
- A custom role includes prompt-injection text or attempts to broaden tool
  access.
- A role hides all non-core capabilities, leaving the user without a visible
  recovery path.

## Test Plan

Verification should include:

- Unit tests for role schema parsing, normalization, duplicate detection, and
  invalid custom role rejection.
- Context patch tests for `role_id` preserve, clear, set, unknown-field, and
  missing-role behavior.
- Capability planner tests proving roles can narrow but not broaden base
  capabilities.
- Context resource tests proving `context` and `workbench:///context` render
  the same role state.
- Sync tests proving role-driven capability changes emit list-changed
  notifications and return fallback capability indexes on timeout.
- Fixture tests for precedence conflicts among role, skill, memory, session,
  and harness inputs.

## Source References

- [Role Management RFC](role-mgmt-rfc.md)
- [Initial Implementation Plan](role-mgmt-initial-implementation-plan.md)
- [Deterministic Role Selection Requirement](role-mgmt-requirement.md)
- `docs/reference/context-contract.md`
- `internal/mcpserver/capabilities.go`
- [MCP tools specification](https://modelcontextprotocol.io/specification/2025-11-25/server/tools)
- [MCP resources specification](https://modelcontextprotocol.io/specification/2025-11-25/server/resources)
- [MCP prompts specification](https://modelcontextprotocol.io/specification/2025-11-25/server/prompts)

## Open Questions

The packet-level human nudges are indexed in
[Role Management RFC](role-mgmt-rfc.md#human-in-the-loop-index).
