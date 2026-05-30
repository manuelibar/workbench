---
id: "idea-mgmt-requirement"
type: "requirement"
title: "Raw Idea Separation Requirement"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Raw Idea Separation Requirement

## Statement

Workbench must preserve a raw idea as a non-committed idea record that remains
separate from any artifact, work item, opportunity, knowledge entry, or memory
candidate created or linked through promotion.

## Rationale

The central value of idea management is preserving speculative context without
turning every thought into accepted work. Separation lets agents retrieve and
refine ideas while humans retain control over commitment. Backlinks from
promotion targets provide provenance, and backlinks from ideas show what the
idea influenced without rewriting its original meaning.

## Acceptance Criteria

- Capturing an idea creates an idea record, not a typed artifact or work item.
- Promotion creates or links a separate target record and records the target in
  the idea's `promotion_targets`.
- A promoted target records a source reference back to the originating idea.
- Editing refinement fields does not erase the raw capture text under the RFC
  default raw-capture policy.
- Closing or parking an idea does not delete promoted targets.
- Listing committed artifacts or work items does not include unpromoted ideas
  unless the user explicitly requests idea context.
- Tests cover capture without promotion, promotion with reciprocal references,
  and lifecycle changes after promotion.

## Source References

- `docs/artifacts/idea-mgmt-rfc.md`
- `docs/artifacts/idea-mgmt-concept-map.md`
- `docs/artifacts/idea-mgmt-risk.md`

## Open Questions

The raw-capture edit policy is tracked as `IM-HIL-003` in the RFC
human-in-the-loop index.
