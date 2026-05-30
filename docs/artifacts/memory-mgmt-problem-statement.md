---
id: "memory-mgmt-problem-statement"
type: "problem_statement"
title: "Memory Management Problem Statement"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Memory Management Problem Statement

## Context

Workbench `main` is intentionally small: it exposes a stdio MCP context and
artifact kernel, stores typed artifacts under `docs/artifacts/`, and keeps the
live context document limited to `focus` and `artifact_id`. The README
explicitly defers memory, knowledge, sessions, namespaces, roles, backlog, and
related systems to epic branches.

That split creates the right foundation, but it also means durable user memory
has no contract yet. Without a contract, later implementation work could mix
memory with artifacts, knowledge, session summaries, or hidden context and make
agent behavior hard to inspect or correct.

## Problem

Workbench needs durable memory, but memory is uniquely risky because it can
change future behavior without being visible in the current prompt. If memory
records lack source, confidence, scope, correction history, and deletion
semantics, agents may treat stale or inferred preferences as facts, preserve
incorrect information after the user corrects it, or personalize work after
the user expected the information to be forgotten.

The problem is not only storage. The core problem is governance: when should
Workbench remember, when should it recall, what should be visible to the user,
and what wins when memory conflicts with the current request, selected
artifacts, sourced knowledge, session context, or runtime context?

## Impact

- Users lose trust if memory appears to be hidden, impossible to inspect, or
  difficult to delete.
- Agents can make repeated mistakes if corrections append noise rather than
  superseding earlier records.
- Project work can drift if a stale memory has more influence than the selected
  artifact that represents current intent.
- Knowledge and memory can become blurred: sourced facts should be refreshed
  or cited, while user memory should capture user-controlled preferences,
  corrections, and project-local facts.
- Security and privacy risks increase if secrets or sensitive information are
  remembered through a model-controlled tool without explicit consent.

## Constraints

- The bootstrap pass may only create docs artifacts; runtime behavior belongs
  to later implementation passes.
- The first runtime design must fit a local stdio MCP server with no required
  HTTP listener or database.
- Artifact conventions require typed Markdown files with deterministic
  frontmatter and non-empty required sections.
- The RFC must remain the living drill hub and index every human nudge in the
  packet.
- External research must use current authoritative sources and final artifacts
  must not cite archive branches or old snapshots.
- Memory must stay subordinate to higher-authority instructions and current
  explicit user intent.

## Source References

- [README.md](../../README.md)
- [docs/reference/context-contract.md](../reference/context-contract.md)
- [docs/reference/artifact-conventions.md](../reference/artifact-conventions.md)
- [memory-mgmt-research-note.md](memory-mgmt-research-note.md)
- [memory-mgmt-rfc.md](memory-mgmt-rfc.md)

## Open Questions

No additional problem-statement nudges are open. Packet-level human nudges are
indexed in [memory-mgmt-rfc.md](memory-mgmt-rfc.md).
