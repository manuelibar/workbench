---
id: "role-mgmt-problem-statement"
type: "problem_statement"
title: "Role Management Problem Statement"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Role Management Problem Statement

## Context

Workbench `main` is intentionally small: a local stdio MCP server with
in-memory context, file-backed typed artifacts, and deterministic capability
planning. The current live context fields are `focus` and `artifact_id`, and
roles are explicitly deferred to this epic along with other higher-level
systems such as memory, skills, namespaces, and sessions.

MCP exposes server behavior through resources, prompts, and tools rather than a
role primitive. Workbench therefore needs a local role contract that composes
with MCP capabilities instead of pretending the protocol owns role semantics.

## Problem

Workbench needs roles that are predictable enough to shape agent behavior, but
the repo currently has no contract for role definitions, active role selection,
custom role storage, role-to-capability mapping, or precedence against other
instruction sources.

Without that contract, later implementation can easily drift into ambiguous
behavior: a role might silently override a skill, memory might contradict a
role, a custom role might shadow a built-in role, or a role might expose tools
that the current context or harness did not authorize.

## Impact

The impact is operational and safety related:

- Agents cannot reliably reproduce why a role changed instructions or visible
  capabilities.
- Users cannot inspect or clear active role state with the same confidence they
  have for current `focus` and `artifact_id`.
- Future skill, memory, session, and harness work can create hidden precedence
  conflicts if role behavior is added opportunistically.
- Custom roles become risky if invalid or malicious definitions are accepted as
  trusted instruction material.

## Constraints

- This bootstrap pass is docs-only and may edit only `docs/artifacts/`.
- The packet must follow the artifact contracts and use the RFC as the living
  drill hub.
- Final artifacts must derive from current `main` docs, current repository
  code, and targeted authoritative research.
- Workbench is a stdio MCP server with no database requirement on `main`.
- The current context contract rejects unknown fields, so any future `role_id`
  field must be introduced deliberately with compatibility tests.
- Capability changes must preserve Workbench's relist and fallback behavior.
- Role semantics must respect MCP's control hierarchy: prompts are user
  selected, resources are application controlled, and tools are model invoked
  with human consent for sensitive operations.

## Source References

- `README.md`
- `docs/reference/context-contract.md`
- `docs/explanation/context-window-thesis.md`
- `docs/explanation/progressive-disclosure.md`
- [Model Context Protocol server feature overview](https://modelcontextprotocol.io/specification/2025-11-25/server/index)

## Open Questions

The packet-level human nudges are indexed in
[Role Management RFC](role-mgmt-rfc.md#human-in-the-loop-index).
