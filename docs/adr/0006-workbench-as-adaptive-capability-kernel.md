# 6. Treat Workbench as an adaptive capability kernel

Date: 2026-05-19

## Status

Accepted

## Context

Workbench must not become a single-purpose MCP server with hardcoded examples. It is the project/namespace-aware layer that prepares an agent to work: select scope, expose current context, surface the right capabilities, remember useful feedback, route work to deterministic services or headless agents, and keep durable state across sessions.

The server should learn from mature infrastructure patterns rather than only AI-agent conventions:

- Linux: small stable primitives, composable process boundaries, files as durable interfaces where appropriate.
- Git: content-addressed history, explicit state transitions, auditability, branch-like experimentation.
- Docker: pluggable drivers and runtime contracts hidden behind stable APIs.
- Kubernetes: declarative resources, controllers/reconcilers, status fields, event streams, and capability discovery.
- MCP: protocol-native resources, tools, prompts, list-changed notifications, and client-driven capability listing.

## Decision

Workbench is an adaptive capability kernel composed of pluggable managers. Each manager owns a domain, exposes a narrow interface, and can be backed by multiple implementations.

Manager boundaries:

- `SessionManager`: durable per-agent session state, selected namespace/project/role/task, context compaction recovery, and resume behavior.
- `NamespaceManager`: groups projects, roles, boards, tasks, knowledge, skills, policy, and integrations under a stable boundary.
- `ProjectManager`: project metadata, repository paths, tech stack, README/docs summaries, system prompt, and external integrations such as GitHub.
- `SkillManager` / `SkillRegistry`: discovers and serves skill bundles. Local filesystem is first; future implementations may include Git, S3, HTTP, IPFS, or signed remote registries.
- `KnowledgeManager`: maintains durable notes, facts, decisions, feedback, and a queryable knowledge graph. It exposes a `query` capability for grounded project knowledge.
- `TaskManager`: owns tasks governed by a deterministic state machine, records implementation progress, and emits resources/events as task state changes.
- `BackgroundManager`: runs background work through controllers/reconcilers, including headless Codex calls used as semantic routers for deterministic Go logic and later API calls.
- `ContextManager`: builds adaptive navigation briefings from selected scope, README/docs, tech stack, skills, tasks, and knowledge state.

`refresh` remains the synchronization boundary. It updates selected scope, runs manager reconciliation, emits MCP capability list-changed notifications, waits briefly for observed client relist calls, and returns a concise navigation briefing. It must not become an exhaustive capability dump except as a bounded compatibility fallback when the client does not relist before timeout.

When scope changes, Workbench uses MCP-native capability discovery:

1. Agent calls `refresh(namespace_id?, project_id?, role_id?, task_id?)`.
2. Workbench updates durable or in-memory selection according to the configured `SessionManager`.
3. Managers reconcile current resources/tools/prompts.
4. Workbench emits relevant MCP list-changed notifications for resources, tools, and prompts.
5. Workbench observes `tools/list` and `resources/list` when the MCP client refreshes its capability view.
6. If those calls are observed before timeout, `refresh()` returns a tidy `client_relisted` result. If not, it returns `fallback_index` with an inline capability index.
7. The agent navigates from the refresh briefing into resources such as project snapshot, README summaries, docs indexes, task state, project knowledge, and surfaced skills.

## Current first-pass implementation

The greenfield implementation now includes concrete early slices:

- project snapshot resource for README/docs/tech-stack discovery
- filesystem skill registry overlay via `WORKBENCH_SKILLS_DIR`
- permanent `query`, `feedback`, and `refresh` core tools
- task state machine with proposed/ready/in_progress/blocked/review/done/cancelled states
- feedback ingestion into queryable knowledge
- navigation hints that point to resources instead of pretending the refresh response is the entire context system

## Consequences

This makes Workbench less opinionated and more extensible. A single deployment can start with local filesystem JSON stores and evolve toward remote registries, object stores, graph databases, or IPFS without changing the agent-facing protocol.

The cost is that Workbench must maintain clear interfaces, resource schemas, and state machines. Every manager should be testable with an in-memory implementation first, then a filesystem implementation, then remote implementations only when the interface is stable.

## Implementation principles

- Prefer API-driven configuration over hardcoded behavior.
- Keep manager interfaces small and explicit.
- Model durable objects as resources with `spec`, `status`, and timestamps where useful.
- Treat tools as mutations, resources as state, prompts/skills as guidance, and notifications as discovery triggers.
- Use headless agents as semantic routers only behind deterministic boundaries: the Go server owns state transitions and validation.
- Store feedback as knowledge input, not as opaque logs.
- Make project context adaptive: tech stack, README/docs summaries, active tasks, selected role, and available skills should influence what the agent sees first.
- Keep ADRs current as part of every architecture-affecting issue or implementation slice.
