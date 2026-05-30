---
id: "afk-system-research-note"
type: "research_note"
title: "AFK System Research Note"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# AFK System Research Note

## Question

Which prior AFK and agent-workflow primitives should Workbench reuse, adapt, or
reimplement for a Codex-oriented AFK system that fits the current local
artifact-kernel architecture?

## Sources

Local Workbench sources read in this worktree:

- [README.md](../../README.md): Workbench is a local stdio MCP context and
  artifact kernel with no database or HTTP listener requirement.
- [Epic branch workflow](../how-to/epic-branch-workflow.md): epic packets must
  be self-contained, current-docs-based, and docs-only during bootstrap.
- [Artifact conventions](../reference/artifact-conventions.md): artifact
  frontmatter, type contracts, RFC hub behavior, and human-in-the-loop index
  requirements.

User GitHub sources inspected on current `main`:

- [manuelibar/ripple README](https://github.com/manuelibar/ripple): frames
  Workbench MCP, Core, and Execution Engine as three layers, with AFK as the
  autonomous backbone.
- [Ripple architecture](https://github.com/manuelibar/ripple/blob/main/docs/manifesto/04-architecture.md):
  separates Workbench MCP, Core, Execution Engine, and agent runner services.
- [Ripple proactive MCP](https://github.com/manuelibar/ripple/blob/main/docs/manifesto/05-proactive-mcp.md):
  describes elicitations, mode mutation, MCP tasks, artifact interaction, and
  traceability tools.
- [Ripple blueprints](https://github.com/manuelibar/ripple/blob/main/docs/manifesto/06-blueprints.md):
  defines versioned blueprint configuration, bounded execution, failure policy,
  trigger policy, and validation rules.
- [Ripple modes catalog](https://github.com/manuelibar/ripple/blob/main/docs/manifesto/07-modes-catalog.md):
  maps modes to tools, resources, output artifacts, and signal semantics.
- [Ripple execution models](https://github.com/manuelibar/ripple/blob/main/docs/manifesto/10-execution-models.md):
  defines prepare, execute, wrap-up; three-phase and persistent-session models;
  and bounded termination.
- [Ripple signal vocabulary](https://github.com/manuelibar/ripple/blob/main/docs/reference/signal-vocabulary.md):
  defines `continue`, `yield`, `done`, and `failed`.
- [Ripple blueprint schema](https://github.com/manuelibar/ripple/blob/main/docs/reference/blueprint-yaml-schema.md):
  details executor, sandbox, memory, tracker, failure strategy, and trigger
  fields.
- [Ripple worker README](https://github.com/manuelibar/ripple/blob/main/ripple-worker/README.md):
  shows the older prepare, execute, wrap-up worker shape, max turns, budget, idle
  timeout, and iteration result publishing.
- [Ripple loop runtime state ADR](https://github.com/manuelibar/ripple/blob/main/ripple-core/docs/adr/0042-loop-runtime-state.md):
  records durable loop state, statuses, config snapshots, and correlation IDs.
- [Ripple conventional commit tool](https://github.com/manuelibar/ripple/blob/main/ripple-cc/README.md):
  shows commit trailers for role, model, token, cost, and issue provenance.

External product and platform inspiration:

- [OpenAI Codex CLI](https://developers.openai.com/codex/cli): Codex runs
  locally in a selected directory, can inspect, edit, and run code, and supports
  scripting, MCP, approvals, and sandbox controls.
- [Codex config reference](https://developers.openai.com/codex/config-reference#configtoml):
  documents user-level and project-scoped config, approvals, sandbox settings,
  MCP server settings, notification hooks, history, and permissions.
- [AFK](https://afkdev.app/): emphasizes remote observation of a full local dev
  environment and agent notifications when work finishes or needs attention.
- [Backlog.md](https://github.com/MrLesk/Backlog.md): uses repo-local Markdown
  tasks, AI-ready MCP/CLI integration, acceptance criteria, planning checkpoints,
  and local/offline ownership.
- [Vibe Kanban](https://github.com/BloopAI/vibe-kanban): uses kanban issues,
  agent workspaces with branches, terminals, dev servers, diff review, and
  multiple coding agents. Its current README says the product is sunsetting, so
  it is inspiration, not a dependency.

## Findings

### Reuse conceptually

| Primitive | Source | Reason |
|---|---|---|
| Four signals: `continue`, `yield`, `done`, `failed` | Ripple signal vocabulary | Small, executor-neutral, and enough to route bounded runs. |
| Prepare, execute, wrap-up | Ripple execution models and worker docs | Clean iteration boundary for context assembly, executor launch, verification, and persistence. |
| Bounded execution invariant | Ripple blueprints and execution models | Essential AFK safety property: every run must terminate, fail, yield, or be cancelled. |
| Config snapshot at run start | Ripple loop runtime state ADR | Prevents mid-run settings drift and makes replay/debugging possible. |
| Progress and traceability evidence | Ripple architecture, observability, and commit tool | AFK output needs provenance, metrics, and source links. |

### Adapt to Workbench

| Primitive | Adaptation |
|---|---|
| Blueprint YAML | Use a smaller Workbench run settings profile. Keep budgets, executor, sandbox, verification, failure policy, and trigger concepts; do not require full blueprint compatibility. |
| Modes | Treat modes as optional run profile metadata at first. Workbench v1 can start with one "execute task" loop and later add mode-specific tool palettes. |
| MCP tasks and notifications | Start with durable progress artifacts and terminal-visible status. Add notification hooks after the run ledger is trustworthy. |
| Backlog-fed work | Define a narrow task-source interface. Consume typed artifacts first, then adapt a future backlog provider. |
| Commit trailers | Preserve the provenance idea, but let Workbench's commit policy be decided by a later write-enabled implementation packet. |
| Persistent session | Keep as an opt-in future setting only. Three-phase runs are safer for the first implementation because files can reconstruct state after restart. |

### Reimplement for Workbench

| Runtime piece | Why reimplement |
|---|---|
| Run ledger | Workbench has no Postgres and should prove file-backed restart inspection first. |
| Executor adapter | Codex configuration, approvals, sandboxing, and MCP setup need a Workbench-specific integration boundary. |
| Safety gate engine | Workbench needs local checks for clean worktree, writable scope, verification commands, budgets, and source references. |
| Progress artifact writer | The artifact kernel needs deterministic Markdown output rather than service logs as the first source of truth. |
| Backlog adapter | Backlog epic is separate; AFK should not depend on a concrete backlog implementation before its contract exists. |

### Codex-specific implications

Codex already has configuration concepts Workbench can align with: approval
policy, sandbox mode, writable roots, MCP servers, notification command, history,
and permissions. Workbench should not silently mutate user-level Codex config.
Instead, a run should resolve its own settings, pass them to the executor through
documented launch configuration or a dedicated profile, and record the resolved
settings in the run evidence.

### External product implications

AFK-style mobile observation is valuable, but Workbench should first make the run
state durable and inspectable. Backlog.md's main lesson is that Markdown tasks
with acceptance criteria and review checkpoints are agent-friendly. Vibe Kanban's
main lesson is that branches, terminals, workspaces, previews, and diff review
are operationally useful, but the sunsetting notice argues against making it a
dependency.

## Implications

The AFK system should proceed as a Workbench-native design:

- Use Ripple vocabulary and invariants as prior art.
- Adapt run settings from blueprint ideas without copying the full schema.
- Reimplement runtime state and safety gates around Workbench artifacts.
- Default to three-phase, file-backed runs until persistent-session recovery is
  proven.
- Treat backlog input as an adapter boundary, not a hard dependency.
- Require human approval before write-enabled AFK is implemented.
- Keep Codex integration explicit and inspectable rather than hidden in global
  config mutation.

## Source References

- [AFK System RFC](afk-system-rfc.md)
- [AFK System Concept Map](afk-system-concept-map.md)
- [AFK System Risk](afk-system-risk.md)

## Open Questions

The research leaves the main implementation choices to `AFK-HITL-001`,
`AFK-HITL-002`, `AFK-HITL-003`, `AFK-HITL-004`, and `AFK-HITL-005` in the RFC.
