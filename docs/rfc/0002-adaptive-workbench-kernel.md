# RFC 0002: Adaptive Workbench Kernel

## Summary

Workbench is a project/namespace-aware MCP capability kernel. It should coordinate durable sessions, projects, skills, tasks, knowledge, feedback, background work, and adaptive context delivery through MCP-native resources, tools, prompts, and list-changed notifications.

## Core objects

- Namespace: a durable boundary grouping projects, roles, boards, knowledge, skills, and integrations.
- Project: repository path, tech stack, README/docs indexes, system prompt, external integrations, and active task state.
- Session: selected namespace/project/role/task plus resume metadata for a specific agent client.
- Skill bundle: guidance and optional contributed resources/tools/prompts, served through a pluggable registry.
- Knowledge item: note, fact, decision, question, feedback, or source excerpt linked into a queryable graph.
- Task: state-machine governed work item with status, owner, evidence, implementation notes, and transitions.
- Background job: reconciler-controlled work such as repository indexing, README summarization, or headless Codex semantic routing.

## Manager interfaces

Workbench should evolve around these pluggable managers:

- SessionManager: create, resume, update selection, persist session state.
- NamespaceManager: CRUD namespaces and namespace-scoped config.
- ProjectManager: CRUD projects, detect tech stack, index README/docs, expose project summaries.
- SkillManager: discover, rank, and serve skills from filesystem first, later Git/S3/HTTP/IPFS.
- KnowledgeManager: ingest feedback and notes, maintain a knowledge graph, answer `ask` queries with citations.
- TaskManager: create/update tasks through a deterministic state machine.
- BackgroundManager: run reconcilers and headless Codex/API jobs behind validated deterministic boundaries.
- ContextManager: compose the refresh briefing from selected scope, docs, tasks, knowledge, and skills.

## MCP contract

`refresh` is a synchronization and navigation tool, not a capability dump.

Expected flow:

1. Agent calls `refresh(namespace_id?, project_id?, role_id?, task_id?)`.
2. Workbench updates session selection.
3. Managers reconcile resources/tools/prompts for the active scope.
4. Workbench emits MCP list-changed notifications for every changed capability family.
5. MCP client lists capabilities and reads resources as needed.
6. `refresh` returns a concise briefing: selected scope, project summary, tech stack, active task/knowledge highlights, and recommended next reads.

## First implementation slices

1. Add durable models for Namespace, Project, Session, Task, and KnowledgeItem.
2. Replace hardcoded project-only selection with namespace/project/session selection.
3. Add filesystem-backed managers with in-memory test doubles.
4. Add project README/docs indexing resources.
5. Add TaskManager with explicit state transitions.
6. Add KnowledgeManager with feedback ingestion and an `ask` query tool.
7. Add filesystem SkillRegistry overlay and later remote registry adapters.
8. Add BackgroundManager reconciler loop for indexing and headless Codex routing.

## Non-goals for the first pass

- No remote S3/IPFS implementation until the registry interface is stable.
- No opaque LLM-owned state transitions; Go validates all mutations.
- No refresh response that substitutes for MCP capability discovery.
