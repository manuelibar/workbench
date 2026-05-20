# Workbench Durable State, Prompts, and Refresh Contract Plan

Date: 2026-05-20 13:45:54
Status: Plan only — no implementation performed

## Goal

Build the next Workbench slice around durable local infrastructure and a clearer MCP contract:

1. Add a Docker Compose development stack centered on Postgres, with room for AFK/Ripple infrastructure such as NATS.
2. Move project identity and namespace mapping toward durable CRUD-backed storage.
3. Model namespaces as a flexible node/edge tree or tag-like hierarchy without introducing a graph database yet.
4. Make prompts first-class scoped capabilities alongside tools, resources, and skills.
5. Clarify the relationship between `refresh` and `workbench:///scope/overview` so agents do not need unnecessary extra round trips but can re-read the current briefing without mutating/synchronizing again.
6. Remove or replace redundant manifest-like resources that duplicate MCP-native list methods, especially `workbench:///scope/capabilities`.
7. Keep feedback-driven skill/prompt creation governed: persist intent as tasks/issues/knowledge first, rather than letting the MCP server mutate code or filesystem skills arbitrarily.

## Current Context / Assumptions

- Repository root: `/home/esquillax/Projects/workbench`
- Current implementation is a Go MCP stdio server.
- Current durable state is limited mostly to `ProjectStore`, backed by JSON through `WORKBENCH_STORE`.
- Current namespace, role, board, task, and knowledge state is process-local.
- Current `refresh()` both changes scope and returns a navigation briefing plus capability sync status.
- Current `workbench:///scope/overview` reads current selection and next resources but does not synchronize capabilities.
- Current `workbench:///scope/capabilities` duplicates MCP capability discovery and should likely be removed or narrowed.
- Current skill resources use canonical `skill://...` URIs, but at least one resource payload still references legacy `workbench:///skills/...` URIs.
- The repo currently does not appear to have a `.git` directory visible from this workspace, so GitHub issue/PR integration should be planned but not assumed available locally.
- The phrase “nets integrations” from voice input is interpreted as likely “NATS integrations” for the future AFK/Ripple service bus.

## Proposed Architecture Direction

Use a small DDD-style layering without overbuilding:

- Domain layer: entity/value definitions and state-machine rules.
- Application layer: manager interfaces and use cases.
- Infrastructure layer: Postgres repositories, Docker Compose, migrations, external integrations.
- MCP adapter layer: tools/resources/prompts that call application services and render protocol results.

Initial persistence should be Postgres via Docker Compose. Avoid graph databases for now by representing namespaces with a relational node/edge model.

## Refresh vs Scope Overview Contract

### `refresh()`

`refresh()` should remain the synchronization boundary.

Responsibilities:

1. Accept optional selection inputs: namespace, project, role, board, task/session later.
2. Validate and persist/update the active session selection.
3. Reconcile dynamic MCP tools, resources, and prompts for the selected scope.
4. Emit MCP list-changed notifications.
5. Wait briefly for client relist behavior where supported.
6. Return the complete current scope overview inline, plus capability sync metadata.
7. Include the URI of the equivalent read-only overview resource.

Important contract:

- In the happy path, the agent should not need to read `workbench:///scope/overview` immediately after `refresh()`.
- The inline `overview` inside `refresh()` and the `workbench:///scope/overview` resource should be the same logical representation, generated from the same composer.
- `refresh()` has side effects and synchronization semantics.
- `workbench:///scope/overview` is read-only and has no capability synchronization semantics.

### `workbench:///scope/overview`

This resource should exist as a snapshot/read model of the latest current scope briefing.

Responsibilities:

1. Let an agent re-read current scope context without calling `refresh()` again.
2. Avoid mutating selection or triggering dynamic capability reconciliation.
3. Point to relevant MCP-native capabilities and resources, but not duplicate full capability manifests.
4. Clearly state that if the caller needs to change scope or refresh dynamic capability lists, it must call `refresh()`.

### Shared Overview Shape

Create one internal composer, likely `ContextManager.ComposeScopeOverview`, used by both:

- `refresh()` result
- `workbench:///scope/overview` resource

Likely fields:

- `selection`
- `summary`
- `active_namespace`
- `active_project`
- `active_role`
- `active_board`
- `active_task` later
- `recommended_resources`
- `recommended_prompts`
- `recommended_tools`
- `knowledge_highlights`
- `task_highlights`
- `capability_state_note`
- `self_resource_uri: workbench:///scope/overview`

Do not include a full tool/resource/prompt manifest in the normal overview. MCP list methods are the manifest.

## Data Model Plan

### Tables

#### `namespaces`

Purpose: durable namespace nodes.

Likely columns:

- `id uuid primary key`
- `slug text not null`
- `display_name text not null`
- `description text`
- `kind text not null default 'namespace'`
- `metadata jsonb not null default '{}'`
- `created_at timestamptz not null`
- `updated_at timestamptz not null`

Constraints:

- unique slug, or unique slug within parent if a strict tree is enforced later.

#### `namespace_edges`

Purpose: represent tree/DAG/tag-like hierarchy without graph DB.

Likely columns:

- `parent_id uuid not null references namespaces(id) on delete cascade`
- `child_id uuid not null references namespaces(id) on delete cascade`
- `edge_type text not null default 'contains'`
- `sort_order int not null default 0`
- `created_at timestamptz not null`
- primary key over `(parent_id, child_id, edge_type)`

Rules:

- Start with tree behavior in application code: one `contains` parent per child.
- Allow future tag/DAG flexibility by keeping edge table general.
- Prevent self-edge.
- Add cycle detection in application code first; later consider recursive SQL checks if needed.

#### `projects`

Purpose: durable project identity and namespace mapping.

Likely columns:

- `id uuid primary key`
- `namespace_id uuid references namespaces(id)`
- `slug text not null`
- `name text not null`
- `description text`
- `system_prompt text`
- `root_path text`
- `metadata jsonb not null default '{}'`
- `created_at timestamptz not null`
- `updated_at timestamptz not null`

Constraints:

- unique `(namespace_id, slug)`

#### `roles`

Purpose: durable role definitions and system prompts.

Likely columns:

- `id uuid primary key`
- `namespace_id uuid references namespaces(id)`
- `slug text not null`
- `name text not null`
- `description text`
- `system_prompt text`
- `metadata jsonb not null default '{}'`
- `created_at timestamptz not null`
- `updated_at timestamptz not null`

#### `prompt_templates`

Purpose: durable reusable prompt templates.

Likely columns:

- `id uuid primary key`
- `namespace_id uuid references namespaces(id)`
- `project_id uuid references projects(id)` nullable
- `slug text not null`
- `name text not null`
- `description text`
- `template text not null`
- `scope text not null default 'global'`
- `status text not null default 'active'`
- `metadata jsonb not null default '{}'`
- `created_at timestamptz not null`
- `updated_at timestamptz not null`

Notes:

- Start with stored templates and context rendering.
- Later add versions, promotion workflow, and evaluation metadata.

#### `tasks`

Purpose: durable governed work queue for Workbench/Ripple slices and feedback-derived actions.

Likely columns:

- `id uuid primary key`
- `namespace_id uuid references namespaces(id)` nullable
- `project_id uuid references projects(id)` nullable
- `title text not null`
- `description text`
- `state text not null`
- `kind text not null default 'workbench_task'`
- `evidence jsonb not null default '[]'`
- `metadata jsonb not null default '{}'`
- `created_at timestamptz not null`
- `updated_at timestamptz not null`

#### `knowledge_items`

Purpose: durable feedback, notes, decisions, source excerpts, and future retrieval material.

Likely columns:

- `id uuid primary key`
- `namespace_id uuid references namespaces(id)` nullable
- `project_id uuid references projects(id)` nullable
- `task_id uuid references tasks(id)` nullable
- `kind text not null`
- `uri text`
- `summary text not null`
- `details text`
- `metadata jsonb not null default '{}'`
- `created_at timestamptz not null`

#### `sessions`

Purpose: durable selected scope and future resume support.

Likely columns:

- `id uuid primary key`
- `client_name text`
- `namespace_id uuid references namespaces(id)` nullable
- `project_id uuid references projects(id)` nullable
- `role_id uuid references roles(id)` nullable
- `board_id uuid` nullable
- `task_id uuid references tasks(id)` nullable
- `metadata jsonb not null default '{}'`
- `created_at timestamptz not null`
- `updated_at timestamptz not null`

Can be a later slice if too much for the first persistence pass.

## Docker Compose Plan

Add a local development stack:

- Postgres for durable Workbench state.
- NATS for future AFK/Ripple eventing and worker coordination.
- Optional pgadmin/adminer only if useful; avoid by default unless Manuel wants it.

Likely files:

- `docker-compose.yml`
- `.env.example`
- `internal/mcpserver/storage/postgres/migrations/0001_initial.sql`
- optionally `scripts/dev-db-reset.sh` later, but skip shell scripts in the first minimal slice unless necessary.

Environment variables:

- `WORKBENCH_DATABASE_URL=postgres://workbench:workbench@localhost:5432/workbench?sslmode=disable`
- `WORKBENCH_STORE_BACKEND=postgres` or infer from `WORKBENCH_DATABASE_URL`
- `WORKBENCH_NATS_URL=nats://localhost:4222`

## Step-by-Step Implementation Plan

### Slice 1: Contract cleanup before persistence

Goal: make refresh/scope overview and capability discovery semantics explicit.

Files likely to change:

- `internal/mcpserver/wire.go`
- `internal/mcpserver/refresh.go`
- `internal/mcpserver/resources.go`
- `internal/mcpserver/mcpserver.go`
- `internal/mcpserver/mcpserver_test.go`
- `internal/mcpserver/refresh_sync_test.go`
- `internal/mcpserver/skills/seeds/workbench-orient/SKILL.md`
- `README.md`
- `docs/adr/0002-use-refresh-as-the-synchronization-boundary.md`
- `docs/adr/0003-distribute-skills-as-skill-uri-resources.md`

Tasks:

1. Introduce a single scope overview composition function used by both `refresh()` and `workbench:///scope/overview`.
2. Change `RefreshResult` to include an `overview` field and `overview_uri` or include URI inside the overview.
3. Ensure the scope overview resource returns the same logical payload without side effects.
4. Remove `workbench:///scope/capabilities` if it only duplicates MCP list methods.
5. If a lightweight resource index is still needed, name it narrowly, e.g. `workbench:///scope/navigation`, but prefer no extra resource until proven needed.
6. Fix `manifest_uri` values to canonical `skill://{name}/manifest` wherever still needed.
7. Update embedded orientation text to say: call `refresh` to synchronize; use scope overview to re-read current briefing; use MCP list methods for full capabilities.
8. Add tests proving `refresh().overview` and `workbench:///scope/overview` have the same selection/navigation content.
9. Add tests proving `scope/capabilities` is absent if removed.
10. Add tests proving no new payload emits `workbench:///skills/...` as canonical skill URI.

Validation:

- `go test ./...`
- `go test -race ./...`
- `make build`
- `hermes mcp test workbench`

### Slice 2: Docker Compose local infrastructure

Goal: provide local Postgres and NATS foundations without changing app behavior yet.

Files likely to change:

- `docker-compose.yml`
- `.env.example`
- `README.md`
- possibly `Makefile`

Tasks:

1. Add Postgres service with named volume.
2. Add NATS service with JetStream enabled if AFK/Ripple event persistence is expected soon; otherwise plain NATS is enough initially.
3. Add health checks.
4. Add Makefile targets such as `dev-up`, `dev-down`, `dev-logs`, `dev-ps`, and possibly `db-url`.
5. Document startup and teardown.
6. Keep application startup compatible with no database configured.

Validation:

- `docker compose up -d postgres nats`
- `docker compose ps`
- `docker compose exec postgres pg_isready -U workbench`
- `docker compose down`

### Slice 3: Postgres migrations and storage package

Goal: create durable schema and migration runner.

Files likely to change:

- `go.mod`
- `go.sum`
- `internal/mcpserver/storage/postgres/db.go`
- `internal/mcpserver/storage/postgres/migrations.go`
- `internal/mcpserver/storage/postgres/migrations/0001_initial.sql`
- `internal/mcpserver/storage/postgres/migrations/0002_seed_default_namespace.sql` or seed in Go startup
- `cmd/workbench-mcp/main.go`

Tasks:

1. Choose Postgres driver, likely `github.com/jackc/pgx/v5`.
2. Add a small migration runner using embedded SQL files to avoid introducing a large migration framework immediately.
3. Create schema for namespaces, namespace_edges, projects, roles, prompt_templates, tasks, knowledge_items, and optionally sessions.
4. Seed default namespace deterministically, e.g. slug `organization` or `default`.
5. Add startup behavior: if `WORKBENCH_DATABASE_URL` is set, connect and migrate; otherwise keep current in-memory/file behavior.
6. Keep migrations idempotent and testable.

Validation:

- Unit tests for migration runner against a test Postgres, if available.
- Manual compose-backed validation with `docker compose exec postgres psql ...`.
- Existing tests still pass without Docker.

### Slice 4: Durable Project and Namespace repositories

Goal: replace JSON-only project storage with Postgres-backed project and namespace managers, while preserving old behavior when no DB is configured.

Files likely to change:

- `internal/mcpserver/domain.go`
- `internal/mcpserver/managers.go`
- `internal/mcpserver/store.go`
- `internal/mcpserver/tools.go`
- `internal/mcpserver/refresh.go`
- `internal/mcpserver/resources.go`
- `internal/mcpserver/storage/postgres/project_repository.go`
- `internal/mcpserver/storage/postgres/namespace_repository.go`
- `internal/mcpserver/*_test.go`

Tasks:

1. Introduce `NamespaceManager` and `ProjectManager` interfaces.
2. Add domain entities with namespace-aware project fields.
3. Implement Postgres repositories.
4. Keep memory implementations for tests.
5. Update `project.create` to accept optional `namespace_id` or infer active/default namespace.
6. Add project namespace mapping to `project.list` and `workbench:///projects/{id}`.
7. Add namespace tree CRUD:
   - create namespace node
   - list namespaces
   - attach namespace parent/child edge
   - optionally detach edge
8. Add cycle prevention for `namespace_edges` in application code.
9. Decide whether namespace delete is hard delete, soft delete, or blocked while children/projects exist; prefer blocked initially.

Validation:

- Unit tests for namespace tree operations.
- Unit tests for project CRUD with namespace mapping.
- Integration test against Postgres if test DB is enabled.
- Existing MCP tests still pass.

### Slice 5: Prompt templates as first-class scoped MCP capabilities

Goal: expose prompts as a dynamic MCP capability family, not just static strings inside projects/roles.

Files likely to change:

- `internal/mcpserver/mcpserver.go`
- `internal/mcpserver/refresh.go`
- `internal/mcpserver/prompts.go` new
- `internal/mcpserver/wire.go`
- `internal/mcpserver/storage/postgres/prompt_repository.go`
- `internal/mcpserver/resources.go`
- `internal/mcpserver/*prompt*_test.go`
- `README.md`
- relevant ADR/RFC docs

Tasks:

1. Verify exact prompt support in `github.com/modelcontextprotocol/go-sdk`.
2. Add `PromptManager` interface.
3. Add prompt template domain model.
4. Store prompt templates in Postgres.
5. Add seed prompt templates for globally useful Workbench operations.
6. On `refresh`, dynamically register prompts relevant to selected namespace/project/role.
7. Include prompt list-changed synchronization if SDK supports prompt notifications.
8. Render prompt templates using selected context.
9. Include recommended prompts in scope overview, but rely on MCP prompt list/get methods for full prompt discovery.
10. Add minimal CRUD tools only if needed now, e.g. `prompt.create`, `prompt.list`, `prompt.update`, but keep them scoped and not permanently visible unless justified.

Validation:

- Tests that prompts appear/disappear based on selected project/namespace/role.
- Tests that rendered prompts include selected project/role context.
- Tests that `refresh` waits for prompt relist if the SDK/client supports it; otherwise document fallback behavior.

### Slice 6: Feedback governance into durable knowledge/tasks/issues

Goal: prevent uncontrolled server-side code/filesystem mutation while preserving the path from feedback to improved skills/prompts.

Files likely to change:

- `internal/mcpserver/knowledge.go`
- `internal/mcpserver/tasks.go`
- `internal/mcpserver/tools.go`
- `internal/mcpserver/storage/postgres/knowledge_repository.go`
- `internal/mcpserver/storage/postgres/task_repository.go`
- `internal/mcpserver/feedback_policy.go` new
- `README.md`
- docs/RFC or ADR for feedback governance

Tasks:

1. Persist feedback as `knowledge_items` with scope links.
2. Add a deterministic feedback classifier/policy in Go.
3. Policy outputs should be explicit actions such as:
   - store knowledge only
   - create task to update a skill
   - create task to update a prompt template
   - create task to investigate bug
   - later create GitHub issue if integration is configured
4. Do not let feedback directly mutate embedded skills or source code.
5. For filesystem skills, require a task/review step before materializing files.
6. Add optional GitHub issue integration only after task persistence works.

Validation:

- Tests for feedback policy decisions.
- Tests that feedback creates scoped knowledge and optionally durable tasks.
- Tests that no skill file mutation occurs from feedback alone.

## Files Likely to Change Across the Whole Plan

- `docker-compose.yml`
- `.env.example`
- `Makefile`
- `go.mod`
- `go.sum`
- `cmd/workbench-mcp/main.go`
- `internal/mcpserver/mcpserver.go`
- `internal/mcpserver/refresh.go`
- `internal/mcpserver/resources.go`
- `internal/mcpserver/tools.go`
- `internal/mcpserver/tasks.go`
- `internal/mcpserver/knowledge.go`
- `internal/mcpserver/store.go`
- `internal/mcpserver/wire.go`
- `internal/mcpserver/domain.go` new
- `internal/mcpserver/managers.go` new
- `internal/mcpserver/prompts.go` new
- `internal/mcpserver/feedback_policy.go` new
- `internal/mcpserver/storage/postgres/*` new
- `internal/mcpserver/skills/seeds/workbench-orient/SKILL.md`
- `README.md`
- `docs/adr/*`
- `docs/rfc/*`
- `docs/plans/*`

## Tests / Validation Matrix

Always run:

- `go test ./...`
- `go test -race ./...`
- `make build`
- `hermes mcp test workbench`

When Docker Compose is added:

- `docker compose config`
- `docker compose up -d postgres nats`
- `docker compose ps`
- `docker compose exec postgres pg_isready -U workbench`
- `docker compose down`

When Postgres repositories are added:

- Existing unit tests must pass without Docker.
- Add integration tests guarded by env var, e.g. `WORKBENCH_TEST_DATABASE_URL`, so normal local tests do not require Docker.
- Run integration tests explicitly against compose DB.

When prompts are added:

- Test prompt registration with no project selected.
- Test prompt registration after project selection.
- Test prompt rendering with project/role context.
- Test prompt list-changed behavior if supported by SDK.

## Risks / Tradeoffs

1. Overbuilding persistence too early
   - Mitigation: start with projects/namespaces and schema foundation, keep manager interfaces small.

2. Namespace graph complexity
   - Mitigation: relational node/edge model, enforce tree semantics first in application code, leave DAG/tag behavior for later.

3. MCP prompt SDK uncertainty
   - Mitigation: do a small SDK spike before committing the prompt capability implementation.

4. Redundant context payloads
   - Mitigation: one composer for both `refresh().overview` and `workbench:///scope/overview`; no separate duplicated logic.

5. Client compatibility with list-changed notifications
   - Mitigation: retain bounded `fallback_index` in refresh for tools/resources/prompts when relist is not observed.

6. Cross-process Postgres migration contention
   - Mitigation: use advisory lock around migrations or document one-instance migration assumption for first pass.

7. GitHub issue integration before durable task model
   - Mitigation: persist Workbench tasks first; GitHub issue creation becomes a later adapter.

## Open Questions

1. Should NATS be included in the very first compose file, or should compose start with Postgres only and add NATS in the AFK slice?
2. Should namespace hierarchy be strict tree initially, or should the edge table permit multiple parents immediately?
3. Should `scope/capabilities` be fully removed now, or replaced with a narrower `scope/navigation` resource?
4. Should prompt CRUD tools be exposed immediately, or should prompt templates be seeded/migrated first and edited through future governed tasks?
5. Should durable sessions be implemented in the first Postgres slice or after project/namespace persistence lands?
6. What should the default namespace be called: `organization`, `default`, `personal`, or configured from env?
7. Should project root path be part of project identity now, or remain `WORKBENCH_PROJECT_ROOT` until multi-project local indexing is ready?

## Recommended First Implementation Order

1. Contract cleanup: shared refresh/scope overview composer and remove redundant capability resource.
2. Docker Compose with Postgres and NATS.
3. Postgres migrations and storage package.
4. Durable namespace/project repositories.
5. Prompt templates as scoped capabilities.
6. Feedback governance into durable knowledge/tasks, then GitHub issue adapter later.
