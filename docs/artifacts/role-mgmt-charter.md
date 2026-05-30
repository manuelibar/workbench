---
id: "role-mgmt-charter"
type: "charter"
title: "Role Management Charter"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Role Management Charter

## Mission

Define deterministic role management for Workbench so a local agent can select a
role, receive role-shaped instructions, and see a role-aware capability map
without weakening the current context, artifact, skill, memory, session, or
harness boundaries.

## Scope

In scope:

- Built-in role definitions with stable IDs, names, descriptions, instruction
  fragments, and capability policies.
- Role selection as explicit Workbench context, including clear behavior for
  setting, preserving, clearing, and reporting the selected role.
- Instruction shaping that composes role guidance with current focus,
  selected artifact, retrieved memory, selected skills, and harness/session
  instructions in a deterministic order.
- Capability mapping that can filter or annotate Workbench tools, resources,
  resource templates, and prompts based on the selected role.
- Custom role storage, validation, loading order, conflict handling, and error
  reporting.
- Precedence boundaries with future skills, memory, persisted sessions, and
  harness configuration.

Out of scope for this epic bootstrap:

- Runtime implementation work.
- Porting historical role code.
- Owning the skill, memory, session, or harness configuration epics beyond the
  role-facing contracts they need.
- Treating roles as a new MCP protocol primitive. Roles must compile down to
  Workbench behavior exposed through MCP context, resources, prompts, and tools.

## Stakeholders

The primary stakeholder is the local user-agent workflow that depends on
Workbench for task context and capability planning. Secondary stakeholders are
the future skill, memory, session, and harness configuration epic owners because
role precedence must not create hidden conflicts with their contracts.

The role-management branch owner is accountable for producing the initial
contract packet and for keeping role decisions traceable through the RFC drill
hub.

## Success Criteria

This epic succeeds when:

- A selected role has a stable, inspectable definition and a deterministic
  effect on generated Workbench context.
- Role selection is explicit, reversible, and reflected in the same context
  document available through the `context` tool and `workbench:///context`.
- Capability mapping is reproducible from current context, role definition, and
  the base capability catalog.
- Custom roles can be loaded from deterministic file-backed storage, validated
  before use, and rejected without changing the active role when invalid.
- Role instructions cannot override higher-precedence user, session, harness,
  or safety constraints, and cannot grant capabilities unavailable from the
  current harness or context state.

## Source References

- `README.md`
- `docs/how-to/epic-branch-workflow.md`
- `docs/reference/artifact-conventions.md`
- `docs/reference/context-contract.md`
- `docs/explanation/context-window-thesis.md`
- `docs/explanation/progressive-disclosure.md`
- [Role Management RFC](role-mgmt-rfc.md)

## Open Questions

The packet-level human nudges are indexed in
[Role Management RFC](role-mgmt-rfc.md#human-in-the-loop-index).
