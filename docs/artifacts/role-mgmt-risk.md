---
id: "role-mgmt-risk"
type: "risk"
title: "Role Precedence Can Expand Authority"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Role Precedence Can Expand Authority

## Description

Role definitions, especially custom roles, could accidentally or maliciously
claim higher precedence than user instructions, skills, memory boundaries,
session policy, harness configuration, or capability permissions.

## Impact

If this risk materializes, a selected role could broaden visible tools, mask
important safety constraints, override a skill workflow, ignore session
boundaries, or make the context document misleading. That would damage the
central Workbench property that visible context and visible capabilities are
deterministic and explainable.

## Likelihood

Medium. The behavior is likely to be designed correctly if precedence is a
first-class contract, but custom roles and future memory/skill/session epics
create many conflict points.

## Mitigation

Mitigations:

- Define a precedence matrix before runtime work.
- Treat custom role content as untrusted until parsed and validated.
- Permit role capability policy to narrow or annotate the base capability plan,
  but not broaden it.
- Reject custom roles that attempt to override reserved built-in IDs or reserved
  precedence fields.
- Expose role diagnostics showing the selected role, source, instruction
  fragments, and capability changes.
- Include tests that exercise conflicts with skills, memory, sessions, and
  harness configuration.

## Owner

The role-management epic owner owns this risk until implementation assigns a
runtime module owner.

## Source References

- [Role Management RFC](role-mgmt-rfc.md)
- [Role Management Concept Map](role-mgmt-concept-map.md)
- [MCP specification security and trust principles](https://modelcontextprotocol.io/specification/2025-11-25#security-and-trust--safety)
- [MCP tools security considerations](https://modelcontextprotocol.io/specification/2025-11-25/server/tools#security-considerations)

## Open Questions

The precedence-related human nudge is indexed as `RM-C1` in
[Role Management RFC](role-mgmt-rfc.md#human-in-the-loop-index).
