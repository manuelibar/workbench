# workbench

Personal MCP server for the agent-native development workflow described in
[manuelibar/ripple](https://github.com/manuelibar/ripple). Single-user,
local-first, streamable HTTP, stateful.

This is the **Workbench MCP** — the front door of a three-layer system
(Workbench / Core / Execution Engine). v0 implements the primitives directly
so the user can dogfood; subsystems extract into separate services later.

## Status

`v0.x.y`. Pre-1.0. Surface is unstable.

## Quickstart

```bash
make compose-up                 # Postgres + pgvector via docker compose
make run                        # workbench-mcp on 127.0.0.1:7777
curl localhost:7777/healthz     # → "ok"
curl localhost:7777/readyz      # → "ready" once the DB is reachable
make test                       # in-memory + integration tests
make smoke                      # boot the binary + curl /healthz, /readyz, /mcp initialize
```

To attach a client (Claude Code, Codex), point the MCP `transport.url` at
`http://127.0.0.1:7777/mcp`. The server is bound to localhost; v0 has no
authentication.

## What v0 exposes

The MCP surface mutates as the agent narrows scope by calling `refresh`
with selection arguments. Visibility rules:

| Surface | Visibility | Tools |
|---|---|---|
| Always-on | always | `refresh`, `ask`, `note.{add,list,search,get,update,delete}`, `namespace.create`, `namespace.list`, `backlog.{add,list,get,update,delete,take_next}` |
| Namespace selected | `refresh(namespace_id=...)` | `namespace.{get,update,delete}`, `project.{create,list}` |
| Project selected | `refresh(project_id=...)` | `project.{get,update,delete}`, `artifact.{create,list,get,update,delete,attach,sign_off,archive}`, `skill.*`, `prompt.*`, `blueprint.{create,list,get,update,delete}` |
| Blueprint selected | `refresh(blueprint_id=...)` | `mode.{create,list,get,update,delete}` |

The `backlog.*` tools proxy over HTTP to the separate
[`backlog-service`](https://github.com/manuelibar/backlog-service) binary
(default `http://127.0.0.1:7778`). The service must be running for those
tools to succeed; the rest of the workbench surface is unaffected if it's
down.

`refresh` is a patch: setting a deeper level preserves the higher levels
and clears the deeper ones. Selecting a project auto-resolves its
namespace; selecting a blueprint auto-resolves its project (and through
it, the namespace). `refresh(clear=true)` wipes the selection entirely.

Always-on resources:

- `workbench://skill` — agent onboarding (read first)

Templated resources (resolution scoped to current selection):

- `workbench:///notes/{id}`
- `workbench:///artifacts/{id}` and `workbench:///artifacts/{id}/{version}`
- `workbench:///skills/{name}`
- `workbench:///prompts/{name}`
- `workbench:///blueprints/{name}/{version}`

## Configuration

| Env var | Default | Notes |
|---|---|---|
| `WORKBENCH_BIND` | `127.0.0.1:7777` | TCP bind address |
| `WORKBENCH_DB_URL` | `postgres://workbench:workbench@127.0.0.1:5432/workbench?sslmode=disable` | Postgres DSN |
| `WORKBENCH_LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |
| `WORKBENCH_BACKLOG_URL` | `http://127.0.0.1:7778` | Base URL of the separate `backlog-service`. Empty disables the `backlog.*` tools. |

## Layout

`golang-standards/project-layout` subset:

- `cmd/workbench-mcp/` — main package
- `internal/` — implementation packages (not importable externally)
- `scripts/` — operational scripts (`smoke.sh`, etc.)
- `build/` — packaging artifacts
- `docker-compose.yml` — local Postgres + pgvector

## License

MIT — see [LICENSE](LICENSE).
