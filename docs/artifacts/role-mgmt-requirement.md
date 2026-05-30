---
id: "role-mgmt-requirement"
type: "requirement"
title: "Deterministic Role Selection Requirement"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Deterministic Role Selection Requirement

## Statement

Workbench role selection must be explicit, deterministic, reversible, and
non-privileging. Setting, preserving, clearing, and reporting the active role
must follow the same tri-state semantics as existing context fields, and a
selected role must not grant capabilities beyond the base capability plan.

This requirement was produced by [Role Management RFC](role-mgmt-rfc.md).

## Rationale

Roles affect instruction shape and capability visibility, so ambiguous role
state would undermine the current Workbench context contract. Users and agents
need to inspect why a role is active, what source defined it, and which
capabilities changed. Restricting roles from broadening capability access keeps
role selection from becoming an implicit permission escalation path.

## Acceptance Criteria

- Calling `context` with an omitted future `role_id` preserves active role
  state.
- Calling `context` with `role_id: null` clears active role state and restores
  the base role-neutral capability plan.
- Calling `context` with a valid role ID selects that role and returns a
  context document that includes selected role ID, title, source, and
  diagnostics.
- Calling `context` with an unknown or invalid role ID returns an error and
  leaves the previous context state unchanged.
- The `context` result and `workbench:///context` resource are byte-for-byte
  identical for selected role state.
- Role capability policy can hide or annotate capabilities but cannot expose a
  tool, resource, resource template, or prompt absent from the base plan.
- Any role-driven capability visibility change triggers the same sync/fallback
  behavior used by current artifact-driven visibility changes.
- Invalid custom role files are reported through diagnostics without becoming
  selectable.

## Source References

- [Role Management RFC](role-mgmt-rfc.md)
- [Role Management Concept Map](role-mgmt-concept-map.md)
- `docs/reference/context-contract.md`
- `docs/explanation/progressive-disclosure.md`

## Open Questions

The capability-broadening tradeoff is indexed as `RM-T1` in
[Role Management RFC](role-mgmt-rfc.md#human-in-the-loop-index).
