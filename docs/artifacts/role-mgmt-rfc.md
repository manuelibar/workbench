---
id: "role-mgmt-rfc"
type: "rfc"
title: "Role Management RFC"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Role Management RFC

## Summary

Role management adds deterministic role definitions, active role selection,
instruction shaping, role-aware capability mapping, custom role storage, and
precedence boundaries to Workbench. The RFC is the living drill hub for the
epic and links the action artifacts that should drive the first implementation
passes.

Action artifacts:

- [Initial Implementation Plan](role-mgmt-initial-implementation-plan.md)
- [Deterministic Role Selection Requirement](role-mgmt-requirement.md)

## Problem

Roles are deferred from `main`, but future Workbench behavior needs them before
skills, memory, sessions, and harness configuration can compose cleanly. A role
can shape how an agent works, which context it sees, and which Workbench
capabilities are relevant. Without a deterministic contract, role behavior can
become hidden prompt text, untraceable tool exposure, or a precedence conflict
with higher-priority instructions.

The problem is not just naming roles. The epic must define the whole boundary:
where roles are stored, how one becomes active, how it affects instructions,
how it maps onto MCP-visible capabilities, and what happens when role guidance
conflicts with skills, memory, session context, or harness configuration.

## Proposal

Implement role management as a local Workbench policy layer:

- Add a `RoleDefinition` model with stable ID, title, summary, instruction
  profile, capability policy, source, status, and validation metadata.
- Add `RoleSelection` to Workbench context through a future tri-state
  `role_id` field that preserves, clears, and sets state like existing context
  fields.
- Render selected role state and diagnostics in the exact context document
  returned by `context` and served by `workbench:///context`.
- Add role inspection and selection affordances through MCP tools/resources
  only after the surface is narrowed by implementation drilling.
- Load custom roles through deterministic file-backed storage with schema
  validation, duplicate detection, reserved ID handling, and rejected-role
  diagnostics.
- Apply `CapabilityPolicy` as a filter over the base capability plan. Roles may
  hide or annotate available capabilities, but must not expose capabilities
  unavailable from current context, permissions, or harness configuration.
- Define a precedence boundary in which role guidance is lower than harness,
  current user/session instructions, and safety constraints, and cannot rewrite
  explicit skill workflows.

The first implementation pass should use
[Initial Implementation Plan](role-mgmt-initial-implementation-plan.md) and
must satisfy
[Deterministic Role Selection Requirement](role-mgmt-requirement.md).

## Tradeoffs

Keeping roles as local policy avoids protocol invention and keeps MCP
interoperability straightforward. The tradeoff is that Workbench must own the
instruction assembly and diagnostics needed to make role behavior explainable.

Letting roles narrow capabilities is conservative and testable. Allowing roles
to broaden capabilities would make a selected role feel more powerful, but it
would blur permission boundaries and complicate list-changed sync. The default
proposal is restriction-only unless the human-in-the-loop decision changes it.

File-backed custom roles fit Workbench's current no-database foundation, but
storage root selection has portability and precedence consequences. The RFC
keeps that as a decision nudge before runtime work.

## Rollout

Rollout should happen in small implementation passes:

1. Add role contracts, parser, validation, and fixture tests without changing
   visible MCP capabilities.
2. Add context state for `role_id`, context patch tests, and context document
   rendering.
3. Add read-only role inspection resources/tools and diagnostics.
4. Add role-aware capability mapping and list-changed sync tests.
5. Add custom role storage roots and invalid-definition reporting.
6. Add instruction shaping and precedence tests against skill, memory, session,
   and harness fixtures as those contracts become available.

Backward compatibility expectation: no existing context call should change
behavior unless `role_id` is supplied or a role-aware capability surface is
selected. Unknown-field rejection should remain strict for fields outside the
approved context contract.

## Open Questions

### Human-in-the-loop Index

| ID | Nudge | Type | Why it matters | Blocks | Default if unanswered |
|---|---|---|---|---|---|
| RM-Q1 | Choose the initial built-in role catalog. | question | Built-ins drive examples, fixtures, and the first user-visible role list. | Initial role fixtures and public docs. | Ship `default`, `software-engineer`, `reviewer`, and `documentarian` only. |
| RM-D1 | Decide custom role storage roots and override order. | decision | Storage order determines portability, conflict behavior, and whether project roles can shadow user roles. | `RoleStore` implementation and validation errors. | Use a project-local configured role directory first, reject built-in ID overrides, and defer user-global storage. |
| RM-T1 | Decide whether selected roles can broaden capabilities or only narrow them. | tradeoff | Broader capability exposure is convenient but weakens security and sync reasoning. | Role-aware capability planner. | Roles may only narrow or annotate the base capability plan. |
| RM-C1 | Confirm precedence boundaries across harness configuration, session, user directions, skills, memory, and roles. | challenge | Conflicts must be deterministic before instruction shaping ships. | Instruction assembler and precedence tests. | Harness/session/user constraints override role; role overrides passive memory defaults; explicit skill workflows remain intact. |

## Source References

- [Role Management Charter](role-mgmt-charter.md)
- [Problem Statement](role-mgmt-problem-statement.md)
- [Concept Map](role-mgmt-concept-map.md)
- [Assumption](role-mgmt-assumption.md)
- [Risk](role-mgmt-risk.md)
- [Research Note](role-mgmt-research-note.md)
- [Initial Implementation Plan](role-mgmt-initial-implementation-plan.md)
- [Deterministic Role Selection Requirement](role-mgmt-requirement.md)
- `README.md`
- `docs/how-to/epic-branch-workflow.md`
- `docs/reference/artifact-conventions.md`
- `docs/reference/context-contract.md`
- `docs/explanation/context-window-thesis.md`
- `docs/explanation/progressive-disclosure.md`
- [MCP specification overview](https://modelcontextprotocol.io/specification/2025-11-25)
- [MCP architecture](https://modelcontextprotocol.io/specification/2025-11-25/architecture)
- [MCP server feature overview](https://modelcontextprotocol.io/specification/2025-11-25/server/index)
- [MCP tools](https://modelcontextprotocol.io/specification/2025-11-25/server/tools)
- [MCP resources](https://modelcontextprotocol.io/specification/2025-11-25/server/resources)
- [MCP prompts](https://modelcontextprotocol.io/specification/2025-11-25/server/prompts)
- [MCP roots](https://modelcontextprotocol.io/specification/2025-11-25/client/roots)
- [modelcontextprotocol/modelcontextprotocol](https://github.com/modelcontextprotocol/modelcontextprotocol)
- [modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk)
