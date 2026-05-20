# RFC-0001 — Workbench/Ripple Boundary Reset

Status: Proposed
Date: 2026-05-16
Owner: Hermes (PM) + Manuel

## Summary
This RFC defines a hard boundary after ecosystem reset:
- Workbench = conversational MCP harness/facade for contextual selection and capability surfacing.
- Ripple = autonomous/AFK orchestration runtime with loop execution and event-driven coordination.

This separation avoids architecture collapse and enables independent evolution.

## Motivation
Current ecosystem risk:
- overlapping abstractions
- duplicate ownership of context and execution
- weak traceability across docs/backlog/execution

We need one clear control plane for conversational context and one clear execution plane for AFK workflows.

## Decision

### Workbench owns
1. Interactive context selection
   - namespace
   - project
   - role
2. Context mutation protocol (refresh contract)
3. Capability surfacing to agents
   - tools/resources/prompts exposed via MCP
4. Artifact-facing workflow UX for conversational planning
5. Adapters to external services (docs/backlog/etc.) via stable interfaces

### Ripple owns
1. Long-running AFK loops/automata
2. Signal/event orchestration
3. Scheduled/external-trigger execution
4. Autonomous work progression across queued tasks
5. Runtime observability of loop state and outcomes

### Shared contracts (explicit, versioned)
1. Context contract
   - selected namespace/project/role + metadata
2. Artifact contract
   - IDs/types/trace links/lifecycle states
3. Backlog contract
   - issue IDs, workflow states, dependencies, correlation IDs
4. Execution request/result contract
   - handoff from conversational plan to AFK run and back

## Non-Goals
- Merging Workbench and Ripple into one runtime
- Implementing full multi-tenant cloud product in MVP
- Finalizing all domain entities before vertical slicing

## Consequences
Positive:
- clearer ownership
- lower coupling
- easier MVP delivery
- easier future hosted offering

Trade-offs:
- requires adapter and contract discipline
- introduces cross-repo coordination overhead

## MVP Implications
Phase 1 (now):
- Workbench: namespace+role selectors and refresh contract extension
- Docs service v0 and backlog service v0 integrated as adapters
- One vertical slice proving artifact -> backlog -> execution handoff

Phase 2:
- Ripple handoff integration for AFK loop execution
- richer eventing and observability

## Open Questions
1. Should context contract be JSON schema in a shared repo now?
2. Where to host shared contracts repo if we reorganize GitHub?
3. Do we preserve backward compatibility with current Workbench project-only refresh payload?

## Acceptance Criteria
- Team can explain boundary in 60 seconds with no ambiguity.
- At least one implemented feature in next sprint follows this boundary.
- No new feature PR merges without mapping to Workbench or Ripple ownership.
