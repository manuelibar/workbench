---
id: "idea-mgmt-initial-implementation-plan"
type: "implementation_plan"
title: "Idea Management Initial Implementation Plan"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Idea Management Initial Implementation Plan

## Objective

Implement the smallest useful idea-management slice: durable idea records,
read-oriented idea resources, explicit capture and refinement tools, typed
relations, lifecycle transitions, and promotion to typed artifacts with
backlinks.

## Steps

1. Define the idea record schema using the concept map fields, including raw
   body, source references, lifecycle status, refinements, relations, promotion
   targets, and timestamps.
2. Add deterministic validation for required idea fields and allowed lifecycle
   states, using the RFC defaults unless a human resolves the lifecycle nudge
   differently.
3. Add read resources for idea lists, individual ideas, inbox views, and
   relation-filtered lookups.
4. Add `idea.capture` to create a raw idea with source references and default
   `captured` status.
5. Add `idea.refine`, `idea.relate`, and `idea.transition` with explicit
   validation and reasons for changes.
6. Add `idea.promote` for typed artifact targets, creating reciprocal source
   references between the idea and artifact.
7. Add tests for schema validation, resource reads, capture, refinement,
   relation creation, lifecycle transitions, artifact promotion, duplicate
   handling, and backlink preservation.
8. Update user-facing docs after the runtime slice lands, including examples
   that distinguish idea capture from committed work.

## Verification

Run the repository test suite and add focused tests around the idea store and
MCP handlers. Manually validate one end-to-end flow: capture a raw idea,
refine it, relate it to an existing artifact, promote it to a new typed
artifact, and confirm both records retain reciprocal references. Validate that
unpromoted ideas do not appear as committed artifacts or work items.

## Rollback

If the runtime slice causes instability, remove the idea tools and resources
from the MCP surface while leaving durable idea files readable for manual
recovery. Because promotion creates normal typed artifacts with source
references, rollback should not require deleting promoted artifacts; instead,
remove or ignore only the idea-management runtime surface until the contract is
fixed.

## Source References

- `docs/artifacts/idea-mgmt-rfc.md`
- `docs/artifacts/idea-mgmt-concept-map.md`
- `docs/artifacts/idea-mgmt-research-note.md`
- `docs/artifacts/idea-mgmt-requirement.md`

## Open Questions

This implementation plan follows the RFC defaults for unresolved nudges
`IM-HIL-001`, `IM-HIL-002`, and `IM-HIL-003`.
