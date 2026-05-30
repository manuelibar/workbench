---
id: "idea-mgmt-concept-map"
type: "spec"
title: "Idea Management Concept Map"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Idea Management Concept Map

## Context

Workbench already treats Markdown artifacts as typed, durable contracts and
uses MCP resources to expose selected context. Idea management should build on
that pattern while remaining distinct from committed artifacts and future work
records. An idea is a speculative context object until an explicit promotion
creates or links a target record.

## Design

Core taxonomy:

| Concept | Meaning | Notes |
|---|---|---|
| Raw idea | The first captured text, source, and timestamp | Preserved as provenance even after refinement or promotion. |
| Idea record | Durable container for raw text, lifecycle, refinements, tags, and relations | The unit that can be listed, read, related, and promoted. |
| Capture source | Session, artifact, file, issue, discussion, research note, human note, or memory observation that produced the idea | Stored as references, not copied without need. |
| Refinement | Structured interpretation of a raw idea | Can add problem framing, hypotheses, impact, constraints, and candidate targets. |
| Relation | Typed edge between an idea and another record | Supports `relates_to`, `duplicates`, `refines`, `contradicts`, `inspires`, `blocks`, `blocked_by`, and `promotes_to`. |
| Promotion target | A committed record created from or linked to an idea | Initial target should be a typed artifact; later targets may include work items, opportunities, knowledge, and memory. |
| Lifecycle state | Current handling state of the idea | Provisional states: `captured`, `triaged`, `refining`, `promoted`, `parked`, and `closed`. |
| Human nudge | A question, decision, challenge, approval, or tradeoff that needs human attention | Indexed from the RFC so unresolved nudges do not hide in scattered notes. |

Lifecycle flow:

1. Capture the idea with raw text and source references.
2. Triage the idea for duplicate, relevance, sensitivity, and target domain.
3. Refine the idea into a clearer problem, opportunity, requirement, risk, or
   proposal when it merits more work.
4. Relate the idea to existing ideas, artifacts, work items, opportunities,
   knowledge, or memory candidates.
5. Promote the idea only when a human or explicit policy chooses a target.
6. Keep the idea record as provenance after promotion, parking, or closure.

Promotion boundaries:

- A promoted artifact is a new committed artifact with its own frontmatter,
  required sections, and source reference back to the idea.
- A promoted work item is future work-management ownership and must not be
  implied merely because an idea exists.
- A promoted opportunity, knowledge entry, or memory candidate is owned by the
  corresponding future epic and should preserve the idea backlink.
- Closing an idea does not delete its promoted targets or source references.

## Interfaces

Candidate idea fields:

| Field | Purpose |
|---|---|
| `id` | Stable idea identifier. |
| `title` | Human-scannable idea title. |
| `status` | Lifecycle state. |
| `raw_body` | Original captured text. |
| `source_refs` | Session, artifact, file, URL, or human note references. |
| `refinement_summary` | Current refined interpretation. |
| `relations` | Typed edges to ideas and target records. |
| `promotion_targets` | Explicit targets created from the idea. |
| `tags` | Lightweight retrieval and triage facets. |
| `created` / `updated` | Deterministic timestamps. |

Candidate MCP surface for later implementation:

- Resources: `workbench:///ideas`, `workbench:///ideas/{id}`,
  `workbench:///idea-inbox`, and relation-filtered resource templates.
- Tools: `idea.capture`, `idea.refine`, `idea.relate`,
  `idea.promote`, and `idea.transition`.
- Artifact bridge: promotion to a typed artifact must create a normal artifact
  and add reciprocal source references.

## Edge Cases

- A captured idea duplicates another idea but has better source evidence; keep
  both records until a human or policy marks one as duplicate and chooses the
  surviving refinement.
- A raw idea contains sensitive or irrelevant material; preserve a redacted
  durable record and keep the sensitive source out of broad context resources.
- One idea spans multiple future systems; use relations and promotion targets
  rather than forcing a single owner too early.
- A promoted target is later rejected; keep the idea record and add a relation
  explaining the target outcome instead of rewriting the original idea.
- Cyclic relations can be valid for context but must not be interpreted as work
  dependencies unless promoted into a system that supports dependency rules.
- A lifecycle transition is requested without enough context; keep the idea in
  its current state and create a human nudge rather than guessing.

## Test Plan

For this docs packet, validate each artifact against its frontmatter and
required-section contract. For a later runtime pass, test capture, listing,
resource reads, refinement updates, relation creation, promotion to artifacts,
backlink preservation, duplicate handling, lifecycle transitions, and
permission or confirmation behavior around destructive or commitment-creating
operations.

## Source References

- `docs/reference/artifact-conventions.md`
- `docs/artifacts/idea-mgmt-rfc.md`
- `docs/artifacts/idea-mgmt-research-note.md`

## Open Questions

The provisional lifecycle states and raw-capture edit policy are tracked as
`IM-HIL-002` and `IM-HIL-003` in the RFC human-in-the-loop index.
