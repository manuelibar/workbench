# workbench

Personal MCP server for the agent-native development workflow described
in [manuelibar/ripple](https://github.com/manuelibar/ripple). Single-user,
local-first, streamable HTTP, stateful.

This is the **Workbench MCP** — the front door of a three-layer system
(Workbench / Core / Execution Engine). v0 implements the primitives
directly so the user can dogfood; subsystems extract into separate
services as the design proves out.
[`backlog-service`](https://github.com/manuelibar/backlog-service) is the
first such extraction.

## Status

`v0.x.y`. Pre-1.0. The MCP and DB surfaces are unstable and may change
between minor releases. Breaking changes are flagged in
[CHANGELOG.md](CHANGELOG.md).

## Quickstart

```bash
# in workbench/
make compose-up                 # Postgres + pgvector via docker compose
make run                        # workbench-mcp on 127.0.0.1:7777
curl localhost:7777/healthz     # → "ok"
curl localhost:7777/readyz      # → "ready" once the DB is reachable
make test                       # in-memory + integration tests
make smoke                      # boot the binary + exercise /healthz,
                                #   /readyz, MCP initialize, optional
                                #   backlog.add round-trip
```

The `backlog.*` MCP tools proxy to the separate
[`backlog-service`](https://github.com/manuelibar/backlog-service) binary.
For the full surface, bring both up:

```bash
# in backlog-service/
make compose-up && make run     # backlog-service on 127.0.0.1:7778
```

To attach a client (Claude Code, Codex), point the MCP `transport.url`
at `http://127.0.0.1:7777/mcp`. The server is bound to localhost; v0 has
no authentication.

## Architecture

```
        MCP client (Claude Code, Codex, ...)
                       │
                       │ streamable HTTP (/mcp)
                       ▼
        ┌─────────────────────────┐  HTTP    ┌──────────────────┐
        │     workbench-mcp       │ ───────► │  backlog-service │
        │     127.0.0.1:7777      │          │  127.0.0.1:7778  │
        │                         │          └─────────┬────────┘
        │  - selection state      │                    │
        │  - notes, namespaces,   │                    ▼
        │    projects, artifacts, │             backlog DB
        │    skills, prompts,     │             (Postgres 5433)
        │    blueprints, modes    │
        └───────────┬─────────────┘
                    │
                    ▼
              workbench DB
              (Postgres 5432)
```

Two binaries, two databases, one MCP front door. The MCP surface mutates
as the agent narrows scope by calling `refresh` with selection
arguments; the SDK fans out `notifications/tools/list_changed` to every
attached MCP session. The `backlog.*` tools forward calls (with the
audit headers) to `backlog-service` over HTTP; the rest of the surface
is in-process against workbench's own Postgres.

### Two session layers

| Layer | Identity | Lifetime | Owns |
|---|---|---|---|
| **WorkSession** | `work_session_id` (UUID) | Days; persisted | Selection state, recent events log, the singleton "main thread" |
| **MCP protocol session** | `Mcp-Session-Id` | One client connection | A view onto the singleton `*mcp.Server`'s tool surface |

v0 enforces one open WorkSession at a time (partial unique index `WHERE
closed_at IS NULL`). Every MCP-protocol session attaches to it; the
SDK's `getServer` factory returns the same `*mcp.Server` for every new
connection, so two clients see the same state.

## Tool surface

The MCP surface mutates as the agent narrows scope by calling `refresh`
with selection arguments. Visibility rules:

| Surface | Visibility | Tools |
|---|---|---|
| Always-on | always | `refresh`, `ask`, `note.{add,list,search,get,update,delete}`, `namespace.create`, `namespace.list`, `backlog.{add,list,get,update,delete,take_next}` |
| Namespace selected | `refresh(namespace_id=...)` | `namespace.{get,update,delete}`, `project.{create,list}` |
| Project selected | `refresh(project_id=...)` | `project.{get,update,delete}`, `artifact.{begin,create,list,get,update,delete,attach,sign_off,archive}`, `skill.*`, `prompt.*`, `blueprint.{create,list,get,update,delete}` |
| Artifact selected | `refresh(artifact_id=...)` or `artifact.begin` | `artifact.{guidance,validate,elicit}` |
| Blueprint selected | `refresh(blueprint_id=...)` | `mode.{create,list,get,update,delete}` |

`refresh` is a **patch with cascade**: setting a deeper level preserves
the higher levels and clears the deeper ones. Selecting a project
auto-resolves its namespace; selecting a blueprint auto-resolves its
project (and through it, the namespace). Selecting an artifact
auto-resolves its project and namespace, clears blueprint/mode, and
persists `artifact_id`. `focus` is a persisted steering string returned
with selection state; clear it with `refresh(clear_focus=true)`.
`refresh(clear=true)` wipes the selection entirely.

Typed artifact authoring uses the existing artifact tables. `artifact.begin`
creates a draft from a registered contract type and writes normalized
`content_jsonb` plus a Markdown `content_text` projection with YAML
frontmatter. `artifact.elicit` asks the human through MCP elicitation and
appends a new artifact version only when accepted.

The `backlog.*` tools are always-on; `project_id` is auto-resolved from
selection on `backlog.add` but stays explicit on `backlog.list` and
`backlog.take_next` (the default is the **master backlog across all
projects**). If `backlog-service` is unreachable, only those six tools
fail — the rest of the surface keeps working.

### Notes are immutable

Notes (the Zettelkasten primitive) are pure capture-time records. They
are never modified to "promote" them into a different artifact type.
When an agent raises an issue from a note (or any other downstream
artifact), the **derived artifact** carries a `source_refs` link back
to the note URI (`workbench:///notes/{id}`). The note stays untouched
and queryable forever. This applies to backlog issues, future
artifact types, and anything else derived from captured context.

## Resources

Always-on:

- `workbench://skill` — agent onboarding (read first).

Templated (resolution scoped to current selection):

- `workbench:///notes/{id}`
- `workbench:///artifacts/{id}` and `workbench:///artifacts/{id}/{version}`
- `workbench:///skills/{name}`
- `workbench:///prompts/{name}`
- `workbench:///blueprints/{name}/{version}`

## Events log

Every MCP tool call is recorded in the `events` table on the WorkSession,
along with audit IDs and a `subject_kind`/`subject_id` pointing at the
acted-upon row. `refresh`'s result echoes the most recent events inline
so agents can see what just happened across the session.

The shape is the **unified events contract**:

| Column | Notes |
|---|---|
| `id`, `occurred_at`, `type`, `subject_kind`, `subject_id` | core shape |
| `payload_jsonb` | type-specific body |
| `work_session_id`, `mcp_session_id` | workbench-specific scope |
| `request_id`, `correlation_id`, `causation_id`, `idempotency_key` | four-part audit |

`backlog-service` mirrors this contract (with `project_id` + `actor`
scope columns instead of `work_session_id` + `mcp_session_id`). Future
extracted services should adopt the same shape; do **not** invent
per-entity `*_events` tables.

## Audit headers (workbench → backlog-service)

The backlog client (`internal/backlogclient`) pulls these from `ctx` on
every outbound call:

| Header | Source | Purpose |
|---|---|---|
| `X-Request-Id` | `id.EnsureRequest(ctx)` | Unique to one MCP tool call |
| `X-Correlation-Id` | inbound MCP request | Higher-level operation grouping |
| `X-Causation-Id` | inbound MCP request | Direct cause |
| `Idempotency-Key` | inbound MCP request | Dedup for safe retries |
| `X-Workbench-Actor` | `user.ID` at startup | Actor identity (default assignee on `take_next`, `events.actor`) |

The four-part audit schema is set by `internal/mcpserver/middleware/ids.go`
and lives in `ctx` for the duration of the tool call.

## Configuration

| Env var | Default | Notes |
|---|---|---|
| `WORKBENCH_BIND` | `127.0.0.1:7777` | TCP bind address |
| `WORKBENCH_DB_URL` | `postgres://workbench:workbench@127.0.0.1:5432/workbench?sslmode=disable` | Postgres DSN |
| `WORKBENCH_LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |
| `WORKBENCH_BACKLOG_URL` | `http://127.0.0.1:7778` | Base URL of the separate `backlog-service`. Empty disables the `backlog.*` tools (handlers return a clear "not configured" error). |

## Layout

`golang-standards/project-layout` subset:

```
cmd/workbench-mcp/        — main package
internal/
├── backlogclient/        — typed HTTP client for backlog-service
├── config/               — env vars → typed Config
├── domain/               — boundary types (Note, Artifact, Project, ...)
├── id/                   — UUIDv7 + four-part Audit on ctx
├── mcpserver/            — streamable HTTP MCP server
│   ├── middleware/       — ids / slog / events
│   └── ...
├── onboarding/           — embedded SKILL.md
└── pgstore/              — pgxpool + hand-rolled migrations
    └── migrations/       — embed.FS-bundled SQL
scripts/                  — operational scripts (smoke.sh)
build/                    — packaging
docker-compose.yml        — local Postgres + pgvector
```

Three runtime deps: `github.com/modelcontextprotocol/go-sdk`,
`github.com/jackc/pgx/v5`, `github.com/google/uuid`.

## Tests

- `make test` runs unit tests under the race detector (skips integration
  with `-short`).
- `make test-integration` runs everything; requires `make compose-up`.
- `make smoke` boots the binary + exercises `/healthz`, `/readyz`, MCP
  `initialize`, and (if `backlog-service` is reachable) one
  `tools/call backlog.add` round-trip.

The `internal/mcpserver` package has black-box integration tests
(`mcpserver_test.go`) and a backlog-specific test
(`TestServer_BacklogViaHTTPClient`) that spins up an `httptest.NewServer`
stub of backlog-service and asserts the full round-trip including
audit-header propagation and OCC version conflicts.

## License

MIT — see [LICENSE](LICENSE).
