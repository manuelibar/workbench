---
id: "idea-mgmt-charter"
type: "charter"
title: "Idea Management Charter"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Idea Management Charter

## Mission

Give Workbench a deliberate place to capture, refine, relate, and promote
ideas without confusing raw thought with committed work. The idea-management
epic owns the path from a quick capture to a refined proposal, artifact, work
item, opportunity, knowledge entry, or memory candidate while preserving the
original idea as independent context.

## Scope

In scope:

- Capture raw ideas from agent sessions, human notes, research, artifacts,
  work threads, opportunities, knowledge material, and memory observations.
- Maintain idea lifecycle states from capture through triage, refinement,
  promotion, parking, and closure.
- Model relations between ideas and artifacts, work items, opportunities,
  knowledge records, memory candidates, and other ideas.
- Define promotion rules that create explicit target records and backlinks
  instead of mutating an idea into committed work.
- Keep human review points visible when an idea becomes actionable or affects
  durable project direction.

Out of scope for this bootstrap:

- Runtime code, migrations, external service synchronization, and UI surfaces.
- Full backlog scheduling, sprint planning, or ownership workflows that belong
  to the work-management epic.
- Artifact deletion, supersession, sign-off, and review mechanics deferred to
  Artifact B2.
- Treating non-current planning material as authoritative source material.

## Stakeholders

The primary stakeholders are the human operator who decides what becomes
committed work, the local agent that needs durable idea context between
sessions, and future epic owners for artifacts, work management,
opportunities, knowledge, and memory. The idea-management epic is accountable
for the idea contract and promotion boundaries; downstream epic owners are
accountable for the target records created after promotion.

## Success Criteria

Idea management is successful when raw ideas can be captured quickly, recovered
later, refined without losing provenance, and promoted only through explicit
paths that preserve backlinks to the original idea. The system should make it
clear whether a record is speculative, refined, or committed, and should give
agents enough relation context to avoid duplicates, revive useful parked
ideas, and explain why an idea was promoted or closed.

## Source References

- `README.md`
- `docs/how-to/epic-branch-workflow.md`
- `docs/reference/artifact-conventions.md`
- `docs/artifacts/idea-mgmt-rfc.md`
- `docs/artifacts/idea-mgmt-research-note.md`

## Open Questions

Promotion scope, lifecycle status approval, and raw-capture edit policy are
tracked as `IM-HIL-001`, `IM-HIL-002`, and `IM-HIL-003` in the RFC
human-in-the-loop index.
