# 0003 — Program Charter (MVP)

Status: Draft
Depends on: 0001, 0002

## Mission
Build Workbench as a configurable MCP harness/facade for artifact-driven system building.

## North Star
An agent can select namespace + project + role and immediately receive the correct tools/resources/prompts, with all planning and execution traceable through artifacts and backlog.

## MVP Product Pillars
1. Workbench MCP Core
   - project selection (already present)
   - namespace selection (new)
   - role selection (new)
   - refresh-driven context mutation contract

2. Documentation Service (Git-first)
   - artifact registry in repo docs
   - standard templates
   - static rendering (GitHub Pages)
   - trace links to backlog

3. Backlog Service
   - issue lifecycle
   - take_next orchestration primitive
   - artifact trace metadata fields

## First Vertical Slice (proposed)
"From Problem Statement to Executable Backlog Item"
- Create artifact set (problem statement + RFC stub + spec stub)
- Link artifacts to one backlog issue
- Select project/namespace/role in Workbench
- Agent fetches context and executes one bounded implementation task

## Exit Criteria for MVP
- Demonstrable end-to-end flow without manual context patching.
- At least one completed slice with traceability proof.
- Documentation site renders and is navigable.
- Backlog item lifecycle exercised through service API.

## Risks
- Scope explosion from ecosystem ambition.
- Confusion between Workbench and Ripple responsibilities.
- Over-modeling before proving one thin slice.

## Mitigations
- Enforce artifact gates before coding expansions.
- Keep Workbench core minimal and adapter-oriented.
- Weekly architecture review using ADRs.
