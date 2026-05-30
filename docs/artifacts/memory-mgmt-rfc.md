---
id: "memory-mgmt-rfc"
type: "rfc"
title: "Memory Management RFC"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Memory Management RFC

## Summary

Add a user-controlled durable memory system to Workbench for explicit remember,
recall, correction, and forgetting. Memory records carry source, confidence,
scope, sensitivity, timestamps, and state so agents can use durable context
without confusing it with artifacts, knowledge, session state, or live context.

This RFC is the living drill hub for the memory management epic. It links the
bootstrap packet, action artifacts, open questions, research, and later
decisions.

Packet references:

- [memory-mgmt-charter.md](memory-mgmt-charter.md)
- [memory-mgmt-problem-statement.md](memory-mgmt-problem-statement.md)
- [memory-mgmt-concept-map.md](memory-mgmt-concept-map.md)
- [memory-mgmt-assumption.md](memory-mgmt-assumption.md)
- [memory-mgmt-risk.md](memory-mgmt-risk.md)
- [memory-mgmt-research-note.md](memory-mgmt-research-note.md)

Action artifacts:

- [memory-mgmt-initial-implementation-plan.md](memory-mgmt-initial-implementation-plan.md)
- [memory-mgmt-test-strategy.md](memory-mgmt-test-strategy.md)

## Problem

Workbench needs memory that survives conversation boundaries, but durable memory
can silently affect future behavior if it is injected without provenance or
user control. The current kernel has deliberately narrow context and typed
artifacts. Memory must preserve that clarity by staying explicit, scoped, and
inspectable.

The core design problem is conflict management. Memory may disagree with the
current user request, selected artifact, repository-local docs, sourced
knowledge, session context, or another memory record. If Workbench does not
define precedence, correction, and forgetting semantics before implementation,
memory will become a source of hidden state rather than a trusted coordination
primitive.

## Proposal

### Capability Shape

Introduce memory as a distinct Workbench domain with four primary operations:

- `memory.remember`: creates a durable memory record from explicit user intent
  or confirmed agent suggestion.
- `memory.recall`: retrieves active memory candidates by query, scope, kind,
  confidence, and freshness.
- `memory.correct`: creates a correction record that supersedes one or more
  active records.
- `memory.forget`: removes memory content from active recall and records the
  minimum deletion state allowed by policy.

Inspection should be available through resources or resource templates so users
and agents can view memory without requiring every memory record to be injected
into the live context document.

### Record Contract

Every durable memory record should include:

- Stable id.
- Kind: preference, project fact, correction, workflow hint, negative memory,
  or explicitly added future kind.
- Scope: user, workspace, project, artifact, or session.
- Source and source reference.
- Confidence: explicit, high, medium, or low.
- Sensitivity: public, personal, sensitive, secret, or unknown.
- State: active, superseded, forgotten, or rejected.
- Created and updated timestamps.
- Supersession links for corrections.

### Boundary Rules

Memory is not knowledge. Knowledge is source-driven and should be cited or
refreshed. Memory is user-controlled durable context that may include
preferences, corrections, and scoped project facts.

Memory is not an artifact. Artifacts are typed Markdown contracts and remain the
source of truth for planned work. Memory may link to artifacts, but it should
not overwrite or silently summarize them.

Memory is not session state. Session observations may be candidates for memory,
but they are not durable unless promoted by explicit remember or confirmation.

Memory is not live context. `context` should remain focused on routing fields
such as `focus` and `artifact_id`; recall can return candidate context for the
agent to use consciously.

### Precedence

Initial precedence for Workbench memory behavior:

1. Platform, system, developer, and safety instructions.
2. Current explicit user request.
3. Selected artifact contract for artifact-centered work.
4. Current Workbench context fields for focus and routing.
5. Sourced knowledge and repository-local evidence for factual claims.
6. Active durable memory matching the current scope for user preferences,
   corrections, and project-local context.
7. Session state not promoted to durable memory.

Current explicit user instructions override memory. Selected artifacts override
memory for artifact work. For factual claims, current sourced knowledge should
be preferred over stale memory. For user-owned preferences and corrections,
memory can be the better source. When there is a material conflict, the agent
should surface it instead of silently choosing the surprising source.

### Forgetting

Forgetting must remove records from normal recall. The default proposal is:

- Delete active memory content from the recallable store.
- Keep a minimal tombstone containing id, scope, deletion time, and deletion
  reason class if needed for audit and sync.
- Do not keep enough content in the tombstone to reconstruct the forgotten
  memory.
- Require explicit user confirmation for broad or ambiguous forget requests.

### Correction

Correction should append a superseding record rather than editing history in
place. The prior record becomes `superseded`, and recall should prefer the
newer active correction. This keeps auditability while preventing stale records
from affecting behavior.

## Tradeoffs

Explicit remember-only writes reduce surprising personalization but may miss
useful information the user expected the system to retain. The default should
favor trust and inspectability over aggressive inference.

Returning recall candidates instead of mutating live context makes agent use
more deliberate, but it requires callers to handle provenance and conflicts.
That cost is acceptable because memory has long-lived behavioral impact.

Tombstones support auditability and sync but can conflict with user
expectations of deletion. The proposed compromise keeps only non-reconstructive
metadata unless the human-in-the-loop decision chooses full hard deletion.

File-backed storage fits the current kernel and keeps the first implementation
simple. It may not be enough for advanced semantic retrieval, so the storage
interface should leave room for a later index or embedding-backed recall layer.

## Rollout

1. Finalize the record contract, precedence model, and human-in-the-loop
   decisions in this RFC.
2. Implement local file-backed memory storage behind a small interface.
3. Add `memory.remember`, `memory.recall`, `memory.correct`, and
   `memory.forget` tools with structured results.
4. Add memory inspection resources or resource templates.
5. Add tests from [memory-mgmt-test-strategy.md](memory-mgmt-test-strategy.md),
   including conflict, correction, forgetting, scope, and sensitivity fixtures.
6. Document runtime usage after the tool contract stabilizes.

## Open Questions

### Human-in-the-loop Index

| ID | Nudge | Type | Why it matters | Blocks | Default if unanswered |
|---|---|---|---|---|---|
| MM-HIL-001 | Should durable writes require an explicit user remember command, or may agents infer and save useful memories? | decision | This sets the trust boundary for personalization and surprise writes. | `memory.remember` write policy and tests. | Require explicit user command or confirmation before durable writes. |
| MM-HIL-002 | Which scopes are required at launch: user, workspace, project, artifact, and session? | tradeoff | Too few scopes cause leakage; too many scopes slow implementation and UX. | Record schema, recall filters, and conflict tests. | Implement user, workspace, project, artifact, and non-durable session scope labels. |
| MM-HIL-003 | Should recall provenance be shown in normal agent answers or only in structured tool results? | tradeoff | Visible provenance builds trust but can make answers noisy. | Recall result formatting and agent guidance. | Always include provenance in structured results; summarize it in prose only when relevant or conflicting. |
| MM-HIL-004 | Should corrections rewrite records in place or append superseding records? | decision | Auditability and stale-record suppression depend on the correction model. | `memory.correct` semantics and storage migration. | Append a correction record and mark prior records superseded. |
| MM-HIL-005 | Should forgetting keep minimal tombstones or perform full hard deletion with no retained id? | approval | Deletion expectations affect privacy, audit, sync, and user trust. | `memory.forget` storage semantics. | Remove content from recall and keep only non-reconstructive tombstone metadata. |
| MM-HIL-006 | Which sensitive categories must memory refuse even when a user asks to remember them? | challenge | Secrets and sensitive personal data can leak through recall, logs, or exports. | Sensitivity policy and rejection tests. | Refuse credentials, API keys, payment data, and illegal or unsafe operational secrets; require explicit confirmation for other sensitive personal data. |
| MM-HIL-007 | Should imported artifact or knowledge evidence be allowed to create memory records automatically? | question | This decides whether memory can be populated from research or docs without direct user intent. | Import flows and knowledge-memory boundary. | Do not auto-create memory from artifacts or knowledge; require user confirmation and source links. |

## Source References

- [README.md](../../README.md)
- [docs/how-to/epic-branch-workflow.md](../how-to/epic-branch-workflow.md)
- [docs/reference/artifact-conventions.md](../reference/artifact-conventions.md)
- [docs/reference/context-contract.md](../reference/context-contract.md)
- [memory-mgmt-research-note.md](memory-mgmt-research-note.md)
- [memory-mgmt-concept-map.md](memory-mgmt-concept-map.md)
