---
id: "memory-mgmt-assumption"
type: "assumption"
title: "Memory Is Durable User-Controlled Context, Not Knowledge"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Memory Is Durable User-Controlled Context, Not Knowledge

## Statement

Workbench memory should store user-controlled durable context: preferences,
corrections, collaboration style, project-local facts, and explicit negative
memory boundaries. It should not become the general knowledge store for
external facts, repository facts without a scope, or sourced research notes.

## Evidence

The current Workbench README separates memory, knowledge, sessions, and
artifacts into distinct future epic branches. The context contract also keeps
the live context narrow, which suggests memory should be recalled deliberately
rather than appended to every context document.

Current product and protocol research points in the same direction. OpenAI's
memory documentation distinguishes saved memories from chat history and says
saved memories are for information the user wants kept in mind. MCP's current
server primitive model distinguishes resources as application-controlled
context and tools as model-controlled actions, which supports exposing memory
inspection as resources while treating remember, correct, and forget as
auditable tools.

## Validation Plan

- During RFC drilling, reject any proposed memory record kind that is really a
  knowledge document, artifact state, or session summary unless it has an
  explicit user-owned memory purpose.
- In the initial implementation plan, keep storage and retrieval interfaces
  capable of linking to artifacts and knowledge without copying their contents
  into memory.
- In tests, create conflict fixtures where knowledge says one thing and memory
  says another, then verify the agent can explain which source is being used
  and why.
- Review this assumption after the knowledge and session epics define their
  own contracts.

## Source References

- [README.md](../../README.md)
- [docs/reference/context-contract.md](../reference/context-contract.md)
- [memory-mgmt-concept-map.md](memory-mgmt-concept-map.md)
- [memory-mgmt-research-note.md](memory-mgmt-research-note.md)

## Open Questions

No assumption-specific nudges are open. Packet-level human nudges are indexed
in [memory-mgmt-rfc.md](memory-mgmt-rfc.md).
