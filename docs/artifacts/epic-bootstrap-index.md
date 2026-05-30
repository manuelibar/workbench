---
id: "epic-bootstrap-index"
type: "research_note"
title: "Epic Bootstrap Index"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Epic Bootstrap Index

## Question

Which Workbench MCP harness epic branches have been bootstrapped, what RFC and
action artifacts did each branch add, which human nudges remain open, and how
do the epics depend on one another?

## Sources

- Current `main` documentation: `docs/how-to/epic-branch-workflow.md` and
  `docs/reference/artifact-conventions.md`.
- Final epic branch packets under `docs/artifacts/` on each `epic/*` branch.
- Research notes embedded in each branch packet.

## Findings

All final epic branches start from current `main`, update only
`docs/artifacts/`, and include a self-contained kickoff packet with a living
RFC drill hub. Research status is complete for bootstrap scope: every branch
has a research note with current repository context and targeted external
sources where relevant.

| Branch | RFC artifact ID | Research status | Added action artifacts | Open human nudges |
|---|---|---|---|---|
| `epic/work-management` | `work-management-rfc` | Bootstrap research complete. | `work-management-initial-implementation-plan`, `work-management-test-strategy` | `WM-HIL-001`, `WM-HIL-002`, `WM-HIL-003`, `WM-HIL-004`, `WM-HIL-005`, `WM-HIL-006` |
| `epic/idea-mgmt` | `idea-mgmt-rfc` | Bootstrap research complete. | `idea-mgmt-initial-implementation-plan`, `idea-mgmt-requirement` | `IM-HIL-001`, `IM-HIL-002`, `IM-HIL-003` |
| `epic/knowledge-mgmt` | `knowledge-mgmt-rfc` | Bootstrap research complete. | `knowledge-mgmt-initial-implementation-plan`, `knowledge-mgmt-test-strategy` | `KM-NUDGE-001`, `KM-NUDGE-002`, `KM-NUDGE-003`, `KM-NUDGE-004`, `KM-NUDGE-005`, `KM-NUDGE-006` |
| `epic/afk-system` | `afk-system-rfc` | Bootstrap research complete. | `afk-system-runbook`, `afk-system-test-strategy` | `AFK-HITL-001`, `AFK-HITL-002`, `AFK-HITL-003`, `AFK-HITL-004`, `AFK-HITL-005` |
| `epic/session-management` | `session-management-rfc` | Bootstrap research complete. | `session-management-initial-implementation-plan`, `session-management-test-strategy` | `SM-HIL-001`, `SM-HIL-002`, `SM-HIL-003`, `SM-HIL-004`, `SM-HIL-005` |
| `epic/namespace-mgmt` | `namespace-mgmt-rfc` | Bootstrap research complete. | `namespace-mgmt-initial-implementation-plan`, `namespace-mgmt-requirement` | `NM-D1`, `NM-D2`, `NM-Q1` |
| `epic/role-mgmt` | `role-mgmt-rfc` | Bootstrap research complete. | `role-mgmt-initial-implementation-plan`, `role-mgmt-requirement` | `RM-Q1`, `RM-D1`, `RM-T1`, `RM-C1` |
| `epic/memory-mgmt` | `memory-mgmt-rfc` | Bootstrap research complete. | `memory-mgmt-initial-implementation-plan`, `memory-mgmt-test-strategy` | `MM-HIL-001`, `MM-HIL-002`, `MM-HIL-003`, `MM-HIL-004`, `MM-HIL-005`, `MM-HIL-006`, `MM-HIL-007` |
| `epic/artifact-b2` | `artifact-b2-rfc` | Bootstrap research complete. | `artifact-b2-initial-implementation-plan`, `artifact-b2-test-strategy` | `B2-D1`, `B2-D2`, `B2-Q1`, `B2-T1`, `B2-C1` |
| `epic/harness-as-a-service` | `harness-as-a-service-rfc` | Bootstrap research complete. | `harness-as-a-service-initial-implementation-plan`, `harness-as-a-service-requirement` | `HASS-HIL-001`, `HASS-HIL-002`, `HASS-HIL-003`, `HASS-HIL-004` |
| `epic/agent-skill-distribution` | `agent-skill-distribution-rfc` | Bootstrap research complete. | `agent-skill-distribution-initial-implementation-plan` | `ASD-HITL-001`, `ASD-HITL-002`, `ASD-HITL-003`, `ASD-HITL-004`, `ASD-HITL-005` |

Cross-epic dependencies:

| Epic | Depends on | Dependency reason |
|---|---|---|
| `work-management` | `namespace-mgmt`, `session-management`, `harness-as-a-service` | Work items need namespace scope, arrive-at-work needs session awareness, and tools should be exposed through harness providers. |
| `idea-mgmt` | `artifact-b2`, `work-management`, `knowledge-mgmt`, `memory-mgmt` | Ideas promote to artifacts and may relate to work, evidence, and explicit memories. |
| `knowledge-mgmt` | `artifact-b2`, `memory-mgmt`, `harness-as-a-service` | Retrieval indexes artifacts, stays bounded from explicit memory, and may use pluggable service adapters. |
| `afk-system` | `work-management`, `session-management`, `artifact-b2`, `harness-as-a-service` | AFK runs consume backlog-shaped work, persist run history, emit progress artifacts, and execute through harness boundaries. |
| `session-management` | `namespace-mgmt`, `memory-mgmt`, `artifact-b2` | Sessions need scoped identity, durable recall boundaries, and links to artifacts/checkpoints. |
| `namespace-mgmt` | current context and artifact kernel | Namespace identity scopes context, artifacts, and later feature providers. |
| `role-mgmt` | `harness-as-a-service`, `agent-skill-distribution`, `memory-mgmt`, `session-management` | Roles shape instructions and capability exposure while respecting harness, skill, memory, and session precedence. |
| `memory-mgmt` | `namespace-mgmt`, `knowledge-mgmt`, `artifact-b2` | Memories need scope, source/provenance, and clear precedence against knowledge and artifacts. |
| `artifact-b2` | current artifact kernel | B2 extends the flat artifact kernel with lifecycle, lineage, sign-off, elicitation, and relationship governance. |
| `harness-as-a-service` | all feature epics | The harness is the MCP distribution and provider boundary for feature services. |
| `agent-skill-distribution` | `harness-as-a-service`, `role-mgmt`, `knowledge-mgmt`, `artifact-b2` | Skill packages are distributed through MCP resources and need capability, role, discovery, and lifecycle boundaries. |

No final packet treats stale historical branches as source evidence. The final
branch set intentionally keeps packet files flat under `docs/artifacts/`
because the current artifact kernel only lists flat Markdown artifacts; nested
packet directories are deferred to Artifact B2.

## Implications

The next integration step is to push `main` and the final epic branches, then
remove stale historical branches from local and remote refs. Implementation
work should proceed from the RFC and action artifacts on each epic branch,
resolving indexed human nudges before committing runtime behavior that depends
on those decisions.
