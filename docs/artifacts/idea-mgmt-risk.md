---
id: "idea-mgmt-risk"
type: "risk"
title: "Idea Management Can Blur Into Backlog Commitment"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Idea Management Can Blur Into Backlog Commitment

## Description

Idea management could become an unreviewed backlog if agents treat every
captured idea as planned work or automatically promote speculative notes into
artifacts, work items, opportunities, knowledge, or memory.

## Impact

If this risk materializes, Workbench will accumulate noisy obligations,
humans will lose trust in promoted records, and future work-management or
artifact surfaces will be polluted with low-confidence material. The practical
damage is not just clutter: agents may optimize around speculative records as
if they were accepted decisions.

## Likelihood

Medium. Agents are likely to preserve more context when a capture tool exists,
and without a strict promotion boundary that useful behavior can turn into
premature commitment.

## Mitigation

Keep lifecycle states explicit, default captured ideas to non-committed status,
require promotion targets to be separate records with backlinks, and require
human confirmation or a clearly documented policy before creating committed
work. Make resources read-oriented by default and reserve model-controlled
tools for explicit capture, refinement, relation, transition, and promotion
actions.

## Owner

The idea-management epic owner owns the risk until the promotion contract is
implemented. After promotion targets land, downstream owners for artifacts,
work items, opportunities, knowledge, and memory own validation of their target
records.

## Source References

- `docs/artifacts/idea-mgmt-rfc.md`
- `docs/artifacts/idea-mgmt-research-note.md`
- `docs/artifacts/idea-mgmt-requirement.md`

## Open Questions

The RFC human-in-the-loop index tracks the approval and tradeoff required to
lock the lifecycle and raw-capture policies as `IM-HIL-002` and `IM-HIL-003`.
