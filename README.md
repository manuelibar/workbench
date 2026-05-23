# Workbench MCP

Workbench is a Go MCP server that gives AI agents a scoped, adaptive working context.

Instead of treating an agent session as a static prompt plus a pile of always-on tools, Workbench acts as an **adaptive capability kernel**: it selects the active namespace, project, role, and board; exposes the resources and tools that make sense for that scope; serves relevant skill bundles; and keeps project metadata in a shared store.

The current implementation runs over MCP stdio. One Workbench process maps to one agent session, which keeps session selection and dynamic capabilities isolated while allowing durable project metadata to be shared through the configured store.

---

## Project status

Workbench is early but working. The current greenfield slice focuses on the MCP core: scoped refresh, dynamic tools/resources, project metadata, project snapshots, task state, feedback-backed knowledge, and skill bundle resources.

It is not yet a fully durable multi-agent work operating system. Session selection, namespaces, roles, boards, tasks, and knowledge are currently process-local unless otherwise noted in the state model below.

---

## Why this exists

Agents need more than a static prompt. They need to know:

- what scope they are working in;
- which project, namespace, role, and board are active;
- which tools are safe and relevant right now;
- where project context lives;
- which tasks are active;
- what prior feedback or knowledge applies.

Workbench provides that through MCP-native tools and resources.

The central pattern is `refresh()`. It is a **synchronization boundary**, not just a setter:

1. Applies the selected scope for the current session.
2. Reconciles dynamic tools and skill resources for that scope.
3. Lets MCP list-changed notifications fire.
4. Waits briefly for the client to call the relevant MCP list methods (`tools/list` and `resources/list`).
5. Returns the current scope overview inline, along with rendered skill instructions, resource hints, and capability sync status.

If the list calls arrive before the timeout, `refresh()` returns a tidy result with `capability_sync.status = "client_relisted"`. If they do not arrive, Workbench falls back to `capability_sync.status = "fallback_index"` and includes an inline capability index so the agent still has the new surface without relying on client cache freshness.

The inline `overview` in the refresh response and the read-only `workbench:///scope/overview` resource are generated from the same composer. In the happy path, an agent should not read the overview resource immediately after `refresh()`; the resource exists so the current briefing can be re-read later without mutating selection or synchronizing capabilities again.

The refresh response is intentionally not the whole world. It points the agent at resources such as project snapshot, tasks, knowledge, selected project, selected role, board, GitHub config, and skill manifests. Full capability manifests come from MCP-native list methods, not a duplicate Workbench `scope/capabilities` resource.

---

## Mental model

Workbench treats MCP capabilities as a live surface that changes with the agent's selected scope.

- `refresh` is the entry point and synchronization boundary.
- Tools are mutations: create projects, create tasks, transition tasks, record feedback.
- Resources are readable state: selected scope, project snapshots, task state, knowledge, GitHub config, and skill files.
- Skills are instruction bundles served as MCP resources and rendered into the refresh briefing.
- Stdio process isolation keeps one agent session from leaking selected scope or dynamic capabilities into another.

---

## Quick start

Prerequisites:

- Go 1.25+
- `make`
- an MCP client that can run stdio servers

Build and test:

```sh
make build
make test
```

Run directly:

```sh
make run
```

Then configure your MCP client to launch `build/workbench-mcp` and call `refresh()` at the start of each agent session.

---

## Architecture

```text
┌──────────────────────────────────────────────────────────────┐
│ Agent session                                                 │
│                                                              │
│  Agent                                                       │
│   │                                                          │
│   │ MCP tools/resources over stdio                           │
│   ▼                                                          │
│  workbench-mcp                                                │
│   ├── selection: project/namespace/role/board  (in memory)   │
│   ├── dynamic tool registrations              (in memory)    │
│   ├── dynamic skill resource registrations     (in memory)   │
│   ├── ProjectStore                             (file-backed)  │
│   ├── SkillRegistry                            (overlay)      │
│   ├── task state                               (in memory)    │
│   ├── feedback knowledge                       (in memory)    │
│   └── integration config                       (env-backed)   │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

One stdio process equals one agent session. Parallel agent sessions have isolated selections and dynamic capability surfaces. Projects are shared through `WORKBENCH_STORE`, defaulting to `~/.workbench/projects.json`.

---

## Current capability model

### Permanent core tools

These are always registered:

| Tool | Purpose |
|---|---|
| `refresh` | Select or clear scope, reconcile capabilities, and return navigation context. |
| `feedback` | Store feedback as queryable Workbench knowledge. |
| `ask` | Query the configured KB service, then synthesize a grounded answer or ad hoc `skill://...` resources. Falls back to in-memory feedback search when no KB is configured. |

`ask` accepts `criteria`, optional `scope` (`namespace_id`, `project_id`, `role`), and optional `limit`. The legacy `query` field still works for the in-memory fallback.

### Dynamic tools before project selection

After `refresh()` with no selected project, Workbench exposes discovery and setup tools:

| Tool | Purpose |
|---|---|
| `project.create` | Create a project. |
| `project.list` | List projects. |
| `project.delete` | Delete a project by ID. |
| `namespace.create` | Create a namespace. |
| `namespace.list` | List namespaces. |
| `role.create` | Create a role with optional system prompt. |
| `role.list` | List roles. |

### Dynamic tools after project selection

After `refresh(project_id=...)`, Workbench exposes project-scoped work tools:

| Tool | Purpose |
|---|---|
| `project.list` | List projects. |
| `namespace.list` | List namespaces. |
| `role.list` | List roles. |
| `task.create` | Create a project task in `proposed` state. |
| `task.list` | List tracked project tasks. |
| `task.transition` | Move a task through the governed state machine. |
| `board.create` | Create a project board within a namespace. |
| `board.list` | List boards for the selected or supplied project. |

Task states are deterministic: `proposed`, `ready`, `in_progress`, `blocked`, `review`, `done`, and `cancelled`.

---

## Resource surface

### Static resources and templates

Always registered:

| URI | Type | Purpose |
|---|---|---|
| `workbench:///scope/overview` | concrete resource | Read-only current scope briefing. Same logical shape as `refresh().overview`; no capability synchronization side effects. |
| `workbench:///github/config` | concrete resource | GitHub org/token-presence config supplied by env. |
| `workbench:///context/project-snapshot` | concrete resource | README/docs/tech-stack snapshot for `WORKBENCH_PROJECT_ROOT`. |
| `workbench:///tasks` | concrete resource | In-memory task state for the active server. |
| `workbench:///knowledge` | concrete resource | Feedback and notes captured as queryable knowledge. |
| `workbench:///projects/{id}` | template | Full project details. |
| `workbench:///roles/{id}` | template | Role details and system prompt. |
| `workbench:///boards/{id}` | template | Board details. |

### Dynamic skill resources

On each `refresh()`, Workbench removes old skill resources and registers the resources for currently visible skill bundles:

| URI | Type | Purpose |
|---|---|---|
| `skill://{name}/manifest` | concrete resource | Bundle metadata and file list. |
| `skill://{name}/SKILL.md` | concrete resource | Rendered skill instructions. |

Embedded seed bundles currently include:

- `workbench-orient`: always visible; explains how to use Workbench.
- `workbench-system-prompt`: project-scoped; renders the selected project's system prompt.
- `go-coding-guidelines`: project-scoped Go engineering guidance.

Filesystem skill bundles can be layered in with `WORKBENCH_SKILLS_DIR`. Filesystem skills win over embedded skills with the same name.

---

## `refresh()` flow

```text
Agent:  tools/call refresh(project_id?, namespace_id?, role_id?, board_id?)
          │
Server:   1. Validate supplied IDs.
          2. Replace the session selection.
          3. Remove old dynamic tools/resources.
          4. Register dynamic tools and skill resources for the new scope.
          5. Wait for MCP notification debounce.
          6. Wait for tools/list and resources/list, bounded by a timeout.
          7. If relisted: return tidy selection/overview/skills/navigation + client_relisted status.
          8. If timed out: return selection/overview/skills/navigation + fallback capability index.
Agent:    If client_relisted, normal MCP capability lists are current; otherwise use the fallback index.
          To re-read the current briefing later without side effects, read workbench:///scope/overview.
```

Selection is currently process-local. Restarting the agent loses selection by design; the agent should call `refresh()` at the start of every session and after context compaction.

---

## State model

| State | Location | Lifetime |
|---|---|---|
| Project selection | In-memory `Server.sel` | Session/process |
| Namespace, role, board selection | In-memory `Server.sel` | Session/process |
| Dynamic tools/resources | MCP server registration state | Session/process |
| Projects | `WORKBENCH_STORE`, default `~/.workbench/projects.json` | Durable/shared |
| Namespaces | In-memory map seeded with `organization` | Session/process |
| Roles | In-memory map | Session/process |
| Boards | In-memory map | Session/process |
| Tasks | In-memory map | Session/process |
| Knowledge | KB service via `WORKBENCH_KB_URL`; in-memory feedback fallback when unset | Durable in KB; fallback is session/process |
| Postgres schema foundation | `WORKBENCH_DATABASE_URL` via embedded migrations | Durable/shared when configured |
| GitHub config | Environment variables | Process |

This is the first greenfield slice. ADR 0006 describes the intended direction: pluggable managers for sessions, namespaces, projects, skills, knowledge, tasks, background work, and context composition.

---

## Build, test, and run

```sh
make build          # builds build/workbench-mcp
make test           # go test -race ./...
make vet
make run            # runs the stdio server directly
make dev-up         # starts local Postgres + NATS
make dev-ps         # shows local service status
make dev-logs       # tails local service logs
make dev-down       # stops local services
make db-url         # prints the default local Postgres URL
```

Hermes MCP verification:

```sh
hermes mcp test workbench
```

Environment variables:

| Variable | Default | Purpose |
|---|---|---|
| `WORKBENCH_STORE` | `~/.workbench/projects.json` | Project file store path. |
| `WORKBENCH_PROJECT_ROOT` | process working directory | Repository/document root used for README/docs/tech-stack snapshots. |
| `WORKBENCH_SKILLS_DIR` | unset | Optional filesystem skill registry layered before embedded skills. |
| `WORKBENCH_DATABASE_URL` | unset | Postgres URL; when set, Workbench connects and applies embedded migrations on startup. |
| `WORKBENCH_STORE_BACKEND` | unset | Set to `postgres` to require `WORKBENCH_DATABASE_URL` and run the Postgres startup path. Project CRUD remains file-backed until the repository slice lands. |
| `WORKBENCH_NATS_URL` | `nats://localhost:4222` in `.env.example` | Reserved for future Ripple/AFK eventing; compose starts NATS with JetStream enabled. |
| `WORKBENCH_GITHUB_ORG` | unset | GitHub organization exposed through `workbench:///github/config`. |
| `WORKBENCH_GITHUB_TOKEN` / `GITHUB_TOKEN` | unset | Token presence for GitHub config; the token value is never exposed. |

---

## Claude Code / MCP client integration

Example MCP config using the built binary:

```json
{
  "mcpServers": {
    "workbench": {
      "command": "/path/to/workbench-mcp"
    }
  }
}
```

Example config running from source:

```json
{
  "mcpServers": {
    "workbench": {
      "command": "go",
      "args": ["run", "./cmd/workbench-mcp"],
      "cwd": "/path/to/workbench"
    }
  }
}
```

Optional KB-backed ask configuration:

```json
{
  "mcpServers": {
    "workbench": {
      "command": "/path/to/workbench-mcp",
      "env": {
        "WORKBENCH_KB_URL": "http://localhost:8080",
        "WORKBENCH_CODEX_COMMAND": "codex"
      }
    }
  }
}
```

Typical agent flow:

1. MCP client spawns `workbench-mcp`.
2. Agent calls `refresh()` to get orientation and discovery/setup tools.
3. Agent creates or lists projects, namespaces, and roles as needed.
4. Agent calls `refresh(project_id=..., namespace_id=..., role_id=..., board_id=...)` to select scope.
5. Agent follows navigation hints and reads resources such as project snapshot, tasks, knowledge, and skill manifests.
6. After context compaction or restart, agent calls `refresh()` again.

---

## What is implemented

- [x] stdio MCP server; one process per agent session.
- [x] `refresh()` synchronization boundary with notification debounce, observed list-call barrier, fallback capability index, and a shared inline/read-only scope overview composer.
- [x] Permanent core tools: `refresh`, `feedback`, `ask`.
- [x] Dynamic scope tools before and after project selection.
- [x] Project CRUD: `project.create`, `project.list`, `project.delete`.
- [x] Namespace and role creation/listing.
- [x] Project-scoped boards.
- [x] Project-scoped task state machine.
- [x] Feedback ingestion into queryable fallback knowledge.
- [x] KB-backed `ask` orchestration over `/content/search`, `/knowledge/query`, and one headless Codex synthesis call.
- [x] Static resources for read-only scope overview, GitHub config, project snapshot, tasks, and knowledge.
- [x] Dynamic skill resources registered through `skill://...` URIs.
- [x] `ProjectStore` with memory and file-backed implementations.
- [x] `SkillRegistry` with embedded, filesystem, and overlay implementations.
- [x] Project snapshot indexing for README/docs/tech-stack discovery.
- [x] Seed skill bundles for Workbench orientation, system prompt rendering, and Go coding guidelines.
- [x] GitHub organization config exposed from MCP stdio environment.
- [x] Tests for refresh synchronization, project snapshots, filesystem skills, task transitions, feedback knowledge, and KB-backed ask routing.
- [x] ADR impact issue template at `.github/ISSUE_TEMPLATE/workbench_spec.yml`.

---

## Not yet implemented

- Durable session resumption.
- Durable namespace, role, board, task, and knowledge stores.
- `project.update`.
- Skill-contributed MCP tools and prompts beyond resources.
- BM25 or other ranked retrieval over skill trigger fields.
- Skill versioned URIs or resource subscription/update notifications.
- Background manager/controllers for autonomous work.
- Full GitHub issue/PR/CI integration beyond config exposure.
- Remote skill registries such as Git, HTTP, S3, or IPFS.

---

## File layout

```text
workbench/
├── cmd/workbench-mcp/
│   └── main.go                         entry point; wires store, registry, server config
├── internal/mcpserver/
│   ├── mcpserver.go                    server struct, setup, stdio run, capability refresh
│   ├── refresh.go                      refresh handler and navigation composition
│   ├── resources.go                    static and dynamic resource handlers
│   ├── tools.go                        core and dynamic tool handlers
│   ├── tasks.go                        task create/list/transition handlers
│   ├── knowledge.go                    feedback-backed knowledge and ask
│   ├── runtime_state.go                namespace, role, board, task, knowledge models
│   ├── project_indexer.go              README/docs/tech-stack snapshot indexing
│   ├── store.go                        project store interfaces and implementations
│   ├── wire.go                         JSON wire types
│   └── skills/
│       ├── registry.go                 skill registry interfaces and types
│       ├── embedded.go                 embedded seed bundles
│       ├── filesystem.go               filesystem skill loader
│       ├── overlay.go                  registry overlay behavior
│       └── seeds/                      embedded skills
├── docs/
│   ├── adr/                            architecture decision records
│   ├── rfc/                            request-for-comment design docs
│   ├── plans/                          implementation plans
│   └── iteration-log-adaptive-kernel.md
├── .github/ISSUE_TEMPLATE/
│   └── workbench_spec.yml              Workbench spec/ADR impact issue template
├── Makefile
├── go.mod
└── README.md
```

Runtime dependencies are `github.com/modelcontextprotocol/go-sdk` and `github.com/google/uuid`.

---

## Design rationale

### Why stdio instead of HTTP?

Workbench mutates tool and resource surfaces per agent session. Stdio gives one process per session, which keeps selections and dynamic registrations isolated. HTTP would require additional session multiplexing and careful notification routing.

### Why is `refresh()` a synchronization boundary?

MCP clients cache or list tool/resource capabilities. If Workbench changed dynamic capabilities and immediately returned, the agent could proceed with stale lists. Workbench now treats `tools/list` and `resources/list` calls as the practical one-to-one stdio signal that the client noticed the list-changed notifications and refreshed its views.

This is intentionally bounded. If those calls do not arrive before the refresh timeout, Workbench does not block forever; it returns a fallback capability index inline in the `refresh()` result.

### Why dynamic tools and skill resources?

The agent should see the capabilities that make sense for the current scope. Discovery/setup tools are useful before a project is selected; task and board tools are useful after project selection. Skill resources are likewise registered only when visible in the active scope.

### Why an adaptive capability kernel?

The long-term goal is not just a project prompt server. Workbench should become the stable, MCP-native control plane that selects context, exposes resources, serves skills, records knowledge, manages tasks, and delegates background work behind deterministic manager boundaries.

See ADR 0006 for the accepted architecture direction.
