---
id: "knowledge-mgmt-risk"
type: "risk"
title: "Knowledge Can Collapse Into Ungoverned Memory"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Knowledge Can Collapse Into Ungoverned Memory

## Description

The knowledge layer may drift into an ungoverned memory system if it stores
inferred facts, user preferences, or feedback summaries as retrievable
knowledge without explicit source records and user-approved memory semantics.

The same risk appears if indexed artifacts become a shadow edit path, or if
retrieval results are treated as durable truth after the source has changed.

## Impact

- Agents may cite inferred or stale material as if it came from a source.
- User preferences may be personalized implicitly without the memory epic's
  approval or user consent.
- Artifact contracts may be bypassed by stale indexed snippets.
- Feedback intended for evaluation may affect future work in opaque ways.
- Security and trust reviews become harder because source authority,
  freshness, and access boundaries are blurred.

## Likelihood

Medium-high. Retrieval systems naturally accumulate chunks, summaries,
rankings, and feedback. Without explicit boundaries, those records can look like
memory or canonical documentation even when they are only indexed evidence.

Confidence is high because the current Workbench roadmap intentionally splits
knowledge, memory, and artifacts into separate epics.

## Mitigation

- Require every knowledge result to carry source ID, document ID, citation,
  indexed timestamp, and freshness state.
- Store feedback as typed evaluation events with no memory mutation side effect.
- Make boundary tests part of the implementation plan: search cannot write
  memory, index refresh cannot edit artifacts, and feedback cannot become a
  user preference.
- Treat artifacts as read-only source inputs and preserve artifact ID plus
  section in citations.
- Surface stale or conflicting results instead of silently suppressing them.
- Keep human nudges for source scope, ranking, artifact indexing, feedback, and
  conflict authority in the RFC index.

## Owner

The initial owner is the `epic/knowledge-mgmt` branch owner. Runtime ownership
should move to the Workbench maintainer role after the knowledge contract and
boundary tests are implemented.

## Source References

- [Knowledge Management RFC](knowledge-mgmt-rfc.md)
- [MCP Elicitation](https://modelcontextprotocol.io/specification/2025-11-25/client/elicitation)
- [MCP Security Best Practices](https://modelcontextprotocol.io/docs/tutorials/security/security_best_practices)

## Open Questions

The memory boundary challenge is tracked as `KM-NUDGE-004` in the RFC.
