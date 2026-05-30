---
id: "memory-mgmt-concept-map"
type: "spec"
title: "Memory Management Concept Map"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Memory Management Concept Map

## Context

Workbench currently has two durable coordination primitives: a small live
context document and typed Markdown artifacts. Memory must be added as a
separate primitive, not as hidden prompt text and not as a replacement for
knowledge, artifacts, or session state.

This concept map defines the taxonomy that later implementation work should
preserve. It is a spec artifact because the boundaries are part of the
technical contract: if a record cannot be classified, it should not be stored
as durable memory without an explicit user decision.

## Design

### Core Entities

| Concept | Definition | Durable by default | Controlled by |
|---|---|---|---|
| Context | Current Workbench focus and selected artifact state. | No, except current process state. | User or agent through `context`. |
| Artifact | Typed Markdown document under `docs/artifacts/` with a contract. | Yes. | Artifact workflow. |
| Knowledge | Sourced external or repository-local factual material used for evidence. | Depends on future knowledge epic. | Source authority and citation policy. |
| Session | Conversation-local state, transient plans, and recent interaction state. | No. | Active conversation. |
| Memory | User-controlled durable fact, preference, correction, or project-local note intended to affect future sessions. | Yes, until corrected or forgotten. | User consent and memory tools. |

### Memory Record Taxonomy

| Kind | Example | Storage rule |
|---|---|---|
| Preference | "Use concise final answers for this workspace." | Store when explicitly requested or confirmed. |
| Project fact | "This repo uses stdio MCP only." | Store if scoped to project/workspace and sourced. |
| Correction | "My timezone is America/Argentina/Cordoba, not UTC." | Store as a superseding record linked to the prior record. |
| Workflow hint | "When I ask for a review, lead with findings." | Store only if it does not conflict with higher-authority instructions. |
| Negative memory | "Do not remember restaurant preferences." | Store as a consent boundary with high precedence inside memory. |
| Forbidden memory | API keys, passwords, or payment credentials. | Do not store in durable memory. |

### Required Metadata

Every durable memory record should have:

- `id`: stable identifier.
- `kind`: preference, project_fact, correction, workflow_hint,
  negative_memory, or another explicitly defined kind.
- `scope`: one of `user`, `workspace`, `project`, `artifact`, or `session`.
  `session` records are recallable during the session but are not durable unless
  promoted by explicit remember.
- `source`: explicit user command, user message, artifact reference, tool
  result, imported file, or correction.
- `source_ref`: optional message, artifact id, file path, or resource URI.
- `confidence`: `explicit`, `high`, `medium`, or `low`. Inferred records cannot
  exceed `medium` without user confirmation.
- `state`: active, superseded, forgotten, or rejected.
- `created_at` and `updated_at`.
- `supersedes` and `superseded_by` links when a correction changes behavior.
- `sensitivity`: public, personal, sensitive, secret, or unknown.

### Operation Taxonomy

| Operation | Purpose | User-control requirement |
|---|---|---|
| Remember | Create a durable memory record. | Requires explicit user command or confirmation. |
| Recall | Retrieve relevant active memory with metadata. | May be model-initiated, but results must expose provenance. |
| Correct | Supersede an incorrect memory and preserve the correction source. | Requires explicit correction intent. |
| Forget | Remove memory content from active recall. | Requires explicit user request and clear target resolution. |
| Inspect | List or read memory records and state. | User-visible and deterministic. |
| Export | Produce a portable representation of memory. | User-visible and bounded by scope. |

### Precedence Model

Memory is advisory context, not instruction authority. The initial precedence
model should be:

1. Platform, system, developer, and safety instructions outside Workbench.
2. Current explicit user request.
3. The selected artifact when work is centered on an artifact contract.
4. Current Workbench context fields, used for routing and focus.
5. Sourced knowledge and repository files, when the task needs external facts.
6. Active durable memory records within matching scope, when the task needs
   user preferences, corrections, or project-local context.
7. Session notes that have not been promoted to durable memory.

When memory conflicts with a higher item, the higher item wins. When memory
conflicts with sourced knowledge, the agent should expose the conflict. Current
cited evidence should drive factual claims, while memory should drive
user-owned preferences and corrections. When memory conflicts with another
memory record, active corrections and newer same-scope explicit records win
over older or inferred records.

## Interfaces

The concept map implies a later MCP surface with at least:

- `memory.remember`: create a durable record with required metadata.
- `memory.recall`: search active memory by query, scope, kind, and confidence.
- `memory.correct`: supersede one or more records with a corrected value.
- `memory.forget`: remove active content for a specific record or resolved
  query.
- `memory.list` or a resource template for user inspection.
- `workbench:///memory` or `workbench:///memory/{id}` resources for
  application-controlled inspection without injecting all memory into context.

Tool results should return structured records rather than only prose so the
agent can distinguish memory from knowledge and artifacts.

## Edge Cases

- A user asks to remember a secret. The operation should reject the write and
  explain that secrets belong in a credential store, not durable memory.
- A user correction matches several records. The tool should return candidates
  and require target selection unless the target is unambiguous.
- A recall result conflicts with the selected artifact. The selected artifact
  should drive the work, and the conflict should be exposed as a possible stale
  memory.
- A memory was inferred from a chat rather than explicitly requested. It should
  have lower confidence and be eligible for confirmation before it affects high
  impact behavior.
- A forget request is broad, such as "forget my deployment preferences." The
  tool should show matched records before destructive removal unless the user
  already named record ids.

## Test Plan

- Validate that each operation preserves the entity boundary between memory,
  knowledge, artifacts, session, and context.
- Test precedence with direct conflicts: current user request versus memory,
  selected artifact versus memory, sourced knowledge versus memory, and
  correction versus prior memory.
- Test metadata completeness by rejecting durable records without source,
  confidence, scope, state, and timestamps.
- Test forgetting by confirming forgotten records do not appear in normal recall
  and that any remaining tombstone cannot reconstruct the forgotten content.
- Test sensitivity handling by rejecting secrets and requiring explicit consent
  for personal or sensitive records.

## Source References

- [memory-mgmt-rfc.md](memory-mgmt-rfc.md)
- [memory-mgmt-research-note.md](memory-mgmt-research-note.md)
- [docs/reference/context-contract.md](../reference/context-contract.md)
- [docs/reference/artifact-conventions.md](../reference/artifact-conventions.md)

## Open Questions

No additional concept-map nudges are open. Packet-level human nudges are
indexed in [memory-mgmt-rfc.md](memory-mgmt-rfc.md).
