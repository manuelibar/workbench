---
id: "memory-mgmt-charter"
type: "charter"
title: "Memory Management Epic Charter"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Memory Management Epic Charter

## Mission

Define Workbench durable memory as a user-controlled MCP capability for
explicit remember, recall, correction, and forgetting, with source,
confidence, and scope metadata on every durable record.

The epic should make memory useful without allowing it to become hidden
instruction state. Memory may inform an agent, but it must not silently
override current user instructions, selected artifacts, sourced knowledge, or
the live context contract.

## Scope

In scope:

- Explicit memory write, recall, correction, and forgetting flows.
- Durable memory records with source, confidence, scope, creation time, update
  time, and supersession or deletion state.
- Precedence rules across current user request, selected artifacts, sourced
  knowledge, session context, and durable memory.
- User inspection, export-oriented representation, correction, and deletion
  controls.
- Privacy and safety rules for sensitive information, secrets, and memory
  writes that could surprise the user.
- MCP surface design for tools, resources, and resource templates that keeps
  memory retrieval deliberate and auditable.

Out of scope for this epic:

- General factual knowledge ingestion that belongs to a knowledge epic.
- Artifact lifecycle work such as archive, supersession, and sign-off beyond
  memory links to artifacts.
- Long-running session persistence, AFK loops, role systems, and backlog
  workflows except where memory must define precedence against them.
- Vector database selection as a product requirement. Storage shape should be
  designed behind an interface so later retrieval implementations can evolve.

## Stakeholders

- Local Workbench users who need durable preferences, project facts, and
  corrections to survive conversation boundaries.
- Agents using Workbench over stdio MCP who need predictable recall and clear
  instructions for when memory may affect behavior.
- Maintainers of the Workbench context and artifact kernel, because memory must
  integrate without bloating the always-visible context surface.
- Future knowledge, session, backlog, and artifact workflow epics that need a
  clear boundary between their records and durable user memory.

## Success Criteria

- Users can explicitly ask Workbench to remember, recall, correct, and forget
  memory records, and each operation has deterministic tool behavior.
- Recall results include enough provenance to explain why a memory was
  retrieved: source, confidence, scope, freshness, and record state.
- Current user instructions and selected artifacts have documented precedence
  over memory, while memory has documented behavior when it conflicts with
  knowledge, session context, or stale artifacts.
- Forgetting removes memory content from normal recall and leaves only the
  minimum audit state approved by the user-controlled deletion policy.
- The first implementation can be tested without a remote service or database,
  consistent with the current local stdio MCP kernel.

## Source References

- [README.md](../../README.md)
- [docs/how-to/epic-branch-workflow.md](../how-to/epic-branch-workflow.md)
- [docs/reference/artifact-conventions.md](../reference/artifact-conventions.md)
- [docs/reference/context-contract.md](../reference/context-contract.md)
- [memory-mgmt-rfc.md](memory-mgmt-rfc.md)

## Open Questions

No charter-specific nudges are open. Packet-level human nudges are indexed in
[memory-mgmt-rfc.md](memory-mgmt-rfc.md).
