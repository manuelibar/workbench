# Adaptive Workbench Kernel Implementation Plan

> For Hermes: Use subagent-driven-development skill to implement this plan task-by-task once the architecture is accepted.

Goal: Evolve Workbench from a project-only skill distribution server into an adaptive MCP capability kernel with pluggable managers for sessions, namespaces, projects, skills, knowledge, tasks, and background work.

Architecture: Keep MCP protocol boundaries stable while replacing hardcoded internals with small manager interfaces. Use in-memory implementations for tests and filesystem implementations for first production use. Treat tools as mutations, resources as state, prompts/skills as guidance, and notifications as capability discovery triggers.

Tech Stack: Go, modelcontextprotocol/go-sdk, filesystem JSON stores, headless Codex CLI later behind BackgroundManager.

---

## Slice 1: Foundation models and manager interfaces

Objective: Define durable domain models and interfaces without changing behavior yet.

Files:
- Create: internal/mcpserver/domain.go
- Create: internal/mcpserver/managers.go
- Test: internal/mcpserver/managers_test.go

Tasks:
1. Define Namespace, Project, Session, Task, KnowledgeItem, BackgroundJob structs.
2. Include ID, spec/status-style fields where useful, CreatedAt, UpdatedAt.
3. Define manager interfaces: SessionManager, NamespaceManager, ProjectManager, SkillManager, KnowledgeManager, TaskManager, BackgroundManager, ContextManager.
4. Add compile-time assertions for existing stores where applicable.
5. Run go test ./...

## Slice 2: Namespace-aware project model

Objective: Add namespaces and make projects belong to namespaces.

Files:
- Modify: internal/mcpserver/store.go
- Modify: internal/mcpserver/tools.go
- Modify: internal/mcpserver/resources.go
- Test: internal/mcpserver/mcpserver_test.go

Tasks:
1. Add Namespace store and namespace CRUD tools.
2. Add NamespaceID to Project.
3. Preserve backwards-compatible default namespace for existing projects.
4. Add workbench:///namespaces/{id} and namespace listing resources.
5. Verify refresh can select namespace/project together.

## Slice 3: Durable sessions

Objective: Make session selection resumable.

Files:
- Create: internal/mcpserver/session_store.go
- Modify: internal/mcpserver/refresh.go
- Modify: cmd/workbench-mcp/main.go
- Test: internal/mcpserver/session_store_test.go

Tasks:
1. Add WORKBENCH_SESSION_ID config/env support.
2. Store selected namespace/project/role/task by session ID.
3. Add refresh(session_id?) semantics or server config-driven session selection.
4. Verify process restart can recover selection.

## Slice 4: Adaptive project context resources

Objective: Surface project README/docs/tech stack through resources.

Files:
- Create: internal/mcpserver/project_indexer.go
- Modify: internal/mcpserver/resources.go
- Test: internal/mcpserver/project_indexer_test.go

Tasks:
1. Add repository path and tech stack fields to Project.
2. Detect README.md and docs/*.md for selected project.
3. Expose workbench:///projects/{id}/summary, /readme, /docs-index, /tech-stack.
4. Make refresh briefing cite these resources as recommended reads.

## Slice 5: Filesystem SkillRegistry overlay

Objective: Load skills from a configured filesystem registry before embedded seeds.

Files:
- Create: internal/mcpserver/skills/filesystem.go
- Create: internal/mcpserver/skills/overlay.go
- Test: internal/mcpserver/skills/filesystem_test.go

Tasks:
1. Add WORKBENCH_SKILLS_DIR config.
2. Parse skill bundle directories containing SKILL.md and optional files.
3. Overlay filesystem skills over embedded skills by name.
4. Keep skill:// URI serving unchanged.

## Slice 6: TaskManager state machine

Objective: Add durable tasks with validated state transitions.

Files:
- Create: internal/mcpserver/tasks.go
- Modify: internal/mcpserver/tools.go
- Modify: internal/mcpserver/resources.go
- Test: internal/mcpserver/tasks_test.go

Tasks:
1. Define states: proposed, ready, in_progress, blocked, review, done, cancelled.
2. Add task.create and task.transition tools.
3. Enforce allowed transitions in Go.
4. Expose task resources and active task summary in refresh.

## Slice 7: KnowledgeManager and query tool

Objective: Ingest feedback into knowledge and query it with citations.

Files:
- Create: internal/mcpserver/knowledge.go
- Modify: internal/mcpserver/tools.go
- Modify: internal/mcpserver/resources.go
- Test: internal/mcpserver/knowledge_test.go

Tasks:
1. Store feedback as KnowledgeItem with links to namespace/project/task.
2. Add a `query` tool using deterministic keyword search first.
3. Return cited knowledge item IDs/resources.
4. Later route semantic query expansion through BackgroundManager/headless Codex.

## Slice 8: BackgroundManager reconcilers

Objective: Add controlled background jobs for indexing and semantic routing.

Files:
- Create: internal/mcpserver/background.go
- Test: internal/mcpserver/background_test.go

Tasks:
1. Add job spec/status model.
2. Add reconciler loop with explicit job states.
3. Add repo indexing job type.
4. Keep headless Codex behind an interface and validate all outputs before state mutation.

## Verification

The current implementation includes the first concrete slices of this architecture:

- project context snapshot resources
- filesystem skill registry overlay
- deterministic task state transitions
- feedback-backed knowledge ingestion and `query`
- refresh navigation toward resources rather than capability dumping

See `docs/iteration-log-adaptive-kernel.md` for the five-iteration implementation log.

After each slice:
- gofmt modified Go files.
- go test ./...
- make build.
- make test.
- hermes mcp test workbench.
- Update README and ADRs when contracts change.
