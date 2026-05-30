---
id: "memory-mgmt-risk"
type: "risk"
title: "Memory Can Become Hidden Authority"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Memory Can Become Hidden Authority

## Description

Durable memory can become hidden authority if agents treat recalled memory as
an instruction source instead of advisory context. The risk is highest when a
memory record is stale, inferred, scoped too broadly, or missing source and
confidence metadata.

## Impact

If this risk materializes, Workbench can produce behavior that surprises the
user: old preferences override current instructions, stale project facts beat
selected artifacts, or unsupported user inferences are applied across unrelated
workspaces. This would weaken trust in Workbench and make failures difficult to
debug because the cause may not be visible in the current conversation.

## Likelihood

Likelihood is medium with high confidence. Memory systems commonly optimize for
helpful personalization, and Workbench agents will be incentivized to use
recalled context. Without explicit precedence and metadata requirements, it is
easy for memory to become indistinguishable from instructions, knowledge, or
session state.

## Mitigation

- Require source, confidence, scope, and state metadata on every durable
  memory record.
- Keep remember, correct, and forget as explicit auditable operations rather
  than passive side effects.
- Treat recall as returning candidates with provenance, not as automatically
  mutating the live context document.
- Define precedence so current user instructions and selected artifacts win
  over durable memory.
- Reject secrets and high-risk sensitive data by default.
- Test direct conflicts across memory, artifacts, knowledge, session context,
  and current user requests.

## Owner

The memory management epic owner is accountable for the contract. Later runtime
implementation owners are accountable for preserving the contract in tools,
resources, storage, and tests.

## Source References

- [memory-mgmt-rfc.md](memory-mgmt-rfc.md)
- [memory-mgmt-test-strategy.md](memory-mgmt-test-strategy.md)
- [memory-mgmt-research-note.md](memory-mgmt-research-note.md)

## Open Questions

No risk-specific nudges are open. Packet-level human nudges are indexed in
[memory-mgmt-rfc.md](memory-mgmt-rfc.md).
