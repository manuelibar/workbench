---
id: "idea-mgmt-rfc"
type: "rfc"
title: "Idea Management RFC"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Idea Management RFC

## Summary

Create an idea-management layer for Workbench that captures speculative ideas
as durable context, refines and relates them over time, and promotes them only
through explicit paths into committed targets. This RFC is the living drill hub
for the epic and links the action artifacts that turn the packet into
implementation work.

Action artifacts:

- [Initial Implementation Plan](idea-mgmt-initial-implementation-plan.md)
- [Raw Idea Separation Requirement](idea-mgmt-requirement.md)

## Problem

Workbench has a durable artifact kernel but no contract for ideas before they
become artifacts or work. Raw ideas currently risk disappearing in chat
history, being scattered across notes, or being promoted too early into
committed surfaces. The missing layer makes it difficult to recover provenance,
compare related ideas, park low-confidence thoughts, and explain why a later
artifact or work item exists.

## Proposal

Define ideas as durable but non-committed context records. Each idea keeps raw
capture text, source references, lifecycle status, refinement fields, typed
relations, and explicit promotion targets. Promotion creates or links a
separate target record; it does not turn the idea itself into committed work.

Initial lifecycle states are provisional pending human approval:

- `captured`: raw idea exists and has not been triaged.
- `triaged`: idea has been checked for relevance, duplicate risk, and target
  area.
- `refining`: idea is being shaped into a problem, requirement, risk,
  opportunity, or proposal.
- `promoted`: at least one committed target exists and links back to the idea.
- `parked`: idea is intentionally preserved without active work.
- `closed`: idea is intentionally retired without current promotion.

Initial relation types should be lightweight context edges:

- `relates_to`
- `duplicates`
- `refines`
- `contradicts`
- `inspires`
- `blocks`
- `blocked_by`
- `promotes_to`

The first runtime slice should expose read-oriented idea resources and a small
set of explicit tools:

- `idea.capture` records a raw idea and source references.
- `idea.refine` updates interpretation fields without erasing provenance.
- `idea.relate` creates typed edges to ideas, artifacts, work items,
  opportunities, knowledge entries, memory candidates, or external sources.
- `idea.transition` changes lifecycle status with a reason.
- `idea.promote` creates or links a committed target and writes backlinks.

Promotion should start with typed artifacts because artifacts already exist on
`main`. Future slices can connect promotion targets to work management,
opportunities, knowledge, and memory after those epics define their own
contracts.

## Tradeoffs

Keeping ideas separate from committed targets adds one more object type and
requires agents to follow promotion rules. The benefit is stronger provenance
and less accidental backlog inflation. Starting with local durable files keeps
the first slice small and consistent with current Workbench artifacts, but may
need indexing once idea volume and relation queries grow. Starting promotion
with artifacts delays deeper integration with future systems, but avoids
inventing contracts for epics that have not landed yet.

## Rollout

1. Land this docs-only bootstrap packet on `epic/idea-mgmt`.
2. Use the implementation plan to define the first runtime slice around local
   idea records, idea resources, capture, refinement, relation, lifecycle, and
   artifact promotion.
3. Validate the raw idea separation requirement before adding broader target
   integrations.
4. Add work-item, opportunity, knowledge, and memory promotion targets only
   after those epics expose stable contracts.
5. Keep this RFC updated with resolved nudges, new action artifacts, research
   sessions, and decisions produced by later drill work.

## Open Questions

The RFC owns the packet's human-in-the-loop nudges. When a nudge is resolved,
append the outcome here and update the affected action artifact.

### Human-in-the-loop Index

| ID | Nudge | Type | Why it matters | Blocks | Default if unanswered |
|---|---|---|---|---|---|
| IM-HIL-001 | Decide whether the first promotion target is typed artifacts only or also stubs for future work items, opportunities, knowledge, and memory. | decision | Promotion scope determines the first schema and which downstream owners are touched. | `idea.promote` target schema and implementation sequencing. | Promote only to typed artifacts at first; allow relation stubs to future target kinds without creating those targets. |
| IM-HIL-002 | Approve provisional lifecycle states: `captured`, `triaged`, `refining`, `promoted`, `parked`, and `closed`. | approval | Status names become durable filters, validation values, and agent instructions. | Lifecycle validation, list filters, and transition guidance. | Use the provisional states with an extension point for later epic-specific statuses. |
| IM-HIL-003 | Choose the raw capture edit policy: immutable raw body with appended corrections, or directly editable raw body. | tradeoff | Provenance quality and cleanup ergonomics move in opposite directions. | `idea.refine`, audit behavior, and promotion provenance. | Treat raw capture as immutable and put cleanup in refinement fields or appended corrections. |

## Source References

- `README.md`
- `docs/how-to/epic-branch-workflow.md`
- `docs/reference/artifact-conventions.md`
- `docs/artifacts/idea-mgmt-charter.md`
- `docs/artifacts/idea-mgmt-problem-statement.md`
- `docs/artifacts/idea-mgmt-concept-map.md`
- `docs/artifacts/idea-mgmt-assumption.md`
- `docs/artifacts/idea-mgmt-risk.md`
- `docs/artifacts/idea-mgmt-research-note.md`
