---
id: "work-management-risk"
type: "risk"
title: "Work Management Can Become a Conflicting Second Context System"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Work Management Can Become a Conflicting Second Context System

## Description

Work management could accidentally become a parallel context system if active
namespace, active project, current board, selected work item, daily plan, and
agent session state are each stored independently without a single active work
environment contract.

The risk is highest around the boundary with namespace-management. This epic
needs to select and use a namespace for work, but namespace-management owns the
namespace identity, hierarchy, and lifecycle.

## Impact

If this risk materializes:

- The agent may load work from one namespace while the user thinks another
  namespace is active.
- Boards, backlog, and daily plans may disagree about WIP and priority.
- Arrive-at-work summaries may omit important reminders or surface stale work
  from the wrong scope.
- Future sessions, memory, AFK, and notification epics may attach context to
  incompatible identifiers.
- Users may stop trusting Workbench as the operational source of truth.

## Likelihood

Medium. The current kernel keeps context intentionally small, which reduces the
immediate risk. The likelihood rises when the first runtime implementation adds
new state fields, work item references, and namespace selection without a
cross-epic agreement.

Confidence is medium because namespace-management is a known dependency but its
final contract is outside this packet.

## Mitigation

- Define active work environment as one explicit object: namespace reference
  plus optional project, board/view, and day-plan date.
- Treat namespace IDs as foreign references owned by namespace-management.
- Make every work view derive from work item state and active environment
  filters, not from separately persisted board cards or plan copies.
- Surface unresolved namespace references as repairable state instead of
  creating replacement namespace records.
- Require the first-session flow to show which namespace and view it is using.
- Keep cross-epic questions centralized in the RFC Human-in-the-loop Index.

## Owner

The branch owner for `epic/work-management` owns the risk until a runtime
implementation owner is assigned. Namespace identity decisions remain owned by
the namespace-management epic.

## Source References

- [RFC: Work Management](work-management-rfc.md)
- [Concept Map](work-management-concept-map.md)
- [Test Strategy](work-management-test-strategy.md)
- `README.md`

## Open Questions

Open human nudges are centralized in the RFC's Human-in-the-loop Index.
