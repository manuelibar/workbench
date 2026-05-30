---
id: "role-mgmt-initial-implementation-plan"
type: "implementation_plan"
title: "Initial Role Management Implementation Plan"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Initial Role Management Implementation Plan

## Objective

Implement the first role-management slice behind deterministic contracts:
validated role definitions, explicit selected role context, read-only
inspection, restrictive capability mapping, and diagnostics that explain role
effects.

This plan was produced by [Role Management RFC](role-mgmt-rfc.md).

## Steps

1. Add a role package or module near the existing context and capability
   modules, following current Go style and avoiding unrelated refactors.
2. Define `RoleDefinition`, `InstructionProfile`, `CapabilityPolicy`,
   `RoleSource`, and validation errors.
3. Add built-in role fixtures for the approved initial catalog from `RM-Q1`, or
   the RFC default catalog if unanswered.
4. Add parser and validation tests for stable IDs, required fields, duplicate
   IDs, reserved IDs, and invalid capability policy.
5. Extend context state with tri-state `role_id` only after tests pin current
   `focus` and `artifact_id` behavior.
6. Render selected role state and diagnostics into the context document and
   verify byte-for-byte parity with `workbench:///context`.
7. Add read-only role inspection resources/tools, keeping mutation limited to
   `context` unless the RFC drill approves a separate selection tool.
8. Apply role capability policy as a filter over the base capability plan and
   add sync tests for changed tools, resources, resource templates, and prompts.
9. Add custom role storage loading with deterministic root order, schema
   validation, and rejected-role diagnostics.
10. Add instruction shaping tests that exercise precedence with harness,
    session, user, skill, and memory fixtures.

## Verification

Required verification before the implementation is considered complete:

- `go test ./...`
- Unit tests for role parsing, built-in catalog validation, custom role
  rejection, and duplicate handling.
- Context tests for `role_id` set, preserve, clear, and missing-role errors.
- Integration tests proving `context` and `workbench:///context` expose the
  same selected role state.
- Capability tests proving role policy cannot broaden the base capability plan.
- Sync tests proving role-driven capability changes trigger the correct
  list-changed categories and fallback capability index behavior.

## Rollback

Rollback should be a normal revert of the implementation commits. The design
should keep persisted custom roles isolated from core artifacts so removing the
runtime feature does not corrupt `docs/artifacts/` or current context/artifact
behavior.

If a role-aware release ships and then needs rollback, clearing `role_id` must
restore the pre-role capability plan before the feature is disabled.

## Source References

- [Role Management RFC](role-mgmt-rfc.md)
- [Role Management Concept Map](role-mgmt-concept-map.md)
- [Deterministic Role Selection Requirement](role-mgmt-requirement.md)
- `docs/reference/context-contract.md`
- `internal/mcpserver/context.go`
- `internal/mcpserver/capabilities.go`

## Open Questions

This plan uses the RFC defaults for `RM-Q1`, `RM-D1`, `RM-T1`, and `RM-C1`
until those nudges are answered in
[Role Management RFC](role-mgmt-rfc.md#human-in-the-loop-index).
