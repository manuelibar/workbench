# Changelog

All notable changes to this project will be documented in this file.
The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.1.0] - 2026-05-10

### Added

- Initial scaffold: `cmd/workbench-mcp` entrypoint with `/healthz` and stub
  `/mcp` route, env-var-driven configuration in `internal/config`,
  `docker-compose.yml` for local Postgres + pgvector, Makefile targets for
  build/run/test/compose, `golang-standards/project-layout` skeleton.
- `internal/domain` boundary types: `User`, `WorkSession`, `Selection`.
- `internal/pgstore` package: pgxpool-backed `Store` with `Open`, `Close`,
  `Ping`, `WithTx`, hand-rolled migration runner with `embed.FS`-bundled
  SQL, and `EnsureSingletonUser` / `EnsureOpenWorkSession` /
  `UpdateSelection` / `CloseWorkSession` for v0 bootstrap.
- Migrations: `0001_init` (users, work_sessions with partial unique index
  enforcing one open session per user, four-part audit columns),
  `0002_pgvector` (`CREATE EXTENSION vector` for later semantic memory).
- `cmd/workbench-mcp` now opens Postgres on boot (with up-to-30s retry),
  applies migrations, and ensures the singleton user + open WorkSession
  before binding the listener. Adds a `/readyz` endpoint (1-second-cached
  DB ping).
- Dependencies: `github.com/jackc/pgx/v5` v5.9.2, `github.com/google/uuid`
  v1.6.0.
- `internal/id`: UUIDv7 helper plus a four-part `Audit` set (request /
  correlation / causation / idempotency) attached to a request
  `context.Context`.
- `internal/onboarding`: embedded `SKILL.md` exposed as the
  `workbench://skill` resource — the agent's first port of call.
- `internal/mcpserver`: streamable-HTTP MCP front door.
  - `Server.Handler` mounts on `/mcp` and serves the SDK's streamable
    HTTP transport, with a fresh per-protocol-session `*mcp.Server` from
    `Server.NewSessionServer`.
  - Always-on tools: `refresh` (sync state + optionally change selection;
    returns selection + tool/resource/prompt summaries inline) and `ask`
    (wraps `ServerSession.Elicit` to ask the user a question).
  - Receiving middleware: four-part-ID propagation (`middleware.IDs`)
    and structured request logging (`middleware.Slog`).
- `cmd/workbench-mcp` mounts the MCP handler at `/mcp` after
  bootstrapping Postgres + the singleton user + open WorkSession.
- Dependencies: `github.com/modelcontextprotocol/go-sdk` v1.6.0.
- `internal/domain.Note` and `internal/domain.Event` boundary types.
- Migration 0003: `notes` table (Zettelkasten primitive: user-scoped
  markdown body, tags, capture-time namespace/project hints, idempotency
  partial unique index) and `events` table (append-only episodic log
  scoped to a WorkSession, indexed by occurred_at desc).
- `pgstore` notes API: `AddNote` (idempotency-key replay returns the
  original row), `GetNote`, `GetNoteByIdempotency`, `ListNotes` (filter
  by tag / namespace / project / since), `SearchNotes` (substring,
  case-insensitive — semantic search deferred), `UpdateNote`,
  `DeleteNote`.
- `pgstore` events API: `RecordEvent`, `RecentEvents`.
- MCP tools (always-on): `note.add`, `note.list`, `note.search`,
  `note.get`, `note.update`, `note.delete`. Tool wire shapes use string
  UUIDs so the auto-generated output schemas are unambiguous.
- `internal/mcpserver/middleware.Events` records every `tools/call` (with
  the tool name as `subject_id`) into the events log, attributed to the
  WorkSession; `tool.call` on success, `tool.failed` on error or
  `IsError=true`.
- `refresh` tool result now includes `recent_events` populated from the
  events log (most-recent-first, default 20).
- Migration runner now takes a Postgres advisory lock, making `Migrate`
  safe under concurrent invocation (e.g. parallel test packages, or two
  workbench-mcp processes booting at once).
- `internal/domain.Namespace` and `internal/domain.Project` boundary types.
- Migrations 0004 (`namespaces`, tree-shaped with `parent_id` self-FK and
  `NULLS NOT DISTINCT` unique-name-per-parent) and 0005 (`projects`,
  optional `namespace_id`, unique name per namespace).
- `pgstore` namespace + project APIs: `CreateNamespace`, `GetNamespace`,
  `GetNamespaceByIdempotency`, `ListNamespaces`, `UpdateNamespace`,
  `DeleteNamespace` (cascades children); same shape for projects.
- MCP tools: `namespace.create` and `namespace.list` (always-on
  bootstrap); `namespace.get`, `namespace.update`, `namespace.delete`,
  `project.create`, `project.list` (visible after a namespace is
  selected). All `id` arguments default to the currently-selected
  namespace.
- `internal/mcpserver` refactor: a single `*mcp.Server` is shared across
  every MCP-protocol session (returned for every `getServer` call). A
  `sync.Mutex`-protected selection + tool-group registry drives surface
  mutation; on selection change, the server's tool surface is rebuilt
  via `RemoveTools` / `AddTool`, and the SDK fans out
  `notifications/tools/list_changed` to every connected session for free.
- `refresh` tool result now uses `SelectionWire` (string UUIDs) so the
  auto-generated JSON-Schema matches the runtime payload. The `tools`
  list is derived from the active tool-group set, so it tracks the
  live surface.
- `namespace.delete` clears the workbench's selection if the deleted
  namespace was currently selected.
- `internal/domain.Artifact`, `ArtifactVersion`, `Skill`, `Prompt`,
  `PromptArg` boundary types.
- Migrations 0006 (`artifacts` + `artifact_versions`, append-only versions
  with `latest_version` head pointer + a status CHECK constraint),
  0007 (`skills`, project-scoped markdown), 0008 (`prompts`, project-scoped
  templates with declared args).
- `pgstore` artifact API: `CreateArtifact` (writes head + version 1 in one
  tx), `GetArtifact`, `GetArtifactVersion` (version 0 = latest),
  `ListArtifacts` (filter by type / status), `AppendArtifactVersion`
  (txn-protected version bump), `SetArtifactStatus`,
  `AttachArtifactParent`, `DeleteArtifact`. Plus `pgstore.skills.*` and
  `pgstore.prompts.*`.
- MCP tools (visible after selecting a project):
  `project.{get,update,delete}`, `artifact.{create,list,get,update,delete}`,
  `skill.{create,list,get,update,delete}`,
  `prompt.{create,list,get,update,delete}`. `artifact.update` appends a new
  version; status changes are an orthogonal field on the same call.
- `refresh` auto-resolves the namespace from the project when the caller
  passes `project_id` without `namespace_id`, so the namespace-scoped
  surface is always co-resident with the project surface.
- Templated MCP resources, scoped to the currently-selected project:
  `workbench:///artifacts/{id}`, `workbench:///artifacts/{id}/{version}`,
  `workbench:///skills/{name}`, `workbench:///prompts/{name}`. Skill /
  prompt resources error politely when no project is selected.
- `project.delete` clears the project / blueprint / mode parts of the
  selection if the deleted project was currently selected.
- `internal/domain.Blueprint` and `internal/domain.Mode` boundary types.
- Migrations 0009 (`blueprints`, versioned via `(project_id, name,
  version)` unique with `CHECK (version > 0)`; updates write a new row
  rather than mutating) and 0010 (`modes`, nested per blueprint version
  with `system_prompt` and `capabilities_jsonb`).
- `pgstore` blueprint API: `CreateBlueprint`, `AppendBlueprintVersion`
  (server-monotonic `MAX(version)+1`), `GetBlueprint`, `GetLatestBlueprint`,
  `GetBlueprintByVersion`, `ListBlueprints` (with `LatestOnly` filter),
  `DeleteBlueprint`, `IsLatestBlueprint`.
- `pgstore` mode API: `CreateMode`, `GetMode`, `GetModeByName`,
  `ListModes`, `UpdateMode`, `DeleteMode`. Mutations enforce
  "latest blueprint version only" via an in-transaction
  `assertBlueprintIsLatestTx` helper; older versions return `ErrConflict`
  with a "blueprint.update first" message.
- MCP tools (visible after selecting a project):
  `blueprint.{create,list,get,update,delete}`, `backlog.{add,list,take_next}`.
  `blueprint.update` writes a new version; `blueprint.list latest_only=true`
  collapses to one row per name.
- MCP tools (visible after selecting a blueprint via
  `refresh(blueprint_id=...)`): `mode.{create,list,get,update,delete}`.
- Templated MCP resource `workbench:///blueprints/{name}/{version}` returns
  the JSON definition for that specific blueprint version.
- `blueprint.delete` clears the blueprint / mode parts of the selection
  if the deleted version was currently selected.
- Artifact lifecycle verbs: `artifact.attach` (records lineage,
  idempotent on duplicate parents), `artifact.sign_off`
  (status='signed_off'), `artifact.archive` (status='archived'). All three
  visible after selecting a project.
- `refresh` is now a patch-with-cascade rather than a full replacement:
  `refresh(blueprint_id=X)` keeps the current namespace and project (and
  auto-resolves them from the blueprint if missing); `refresh(project_id=Y)`
  keeps the namespace and clears blueprint/mode. `refresh(clear=true)`
  still wipes everything.
- `scripts/smoke.sh` boots the binary and exercises `/healthz`, `/readyz`,
  and an MCP `initialize` round-trip over real streamable HTTP. Wired
  into the Makefile as `make smoke`.
- End-to-end integration test covering the full
  namespace → project → blueprint → mode selection chain, blueprint
  immutability (mode mutations rejected on stale versions),
  `workbench:///blueprints/{name}/{version}` resource reads, the backlog
  verbs, and `artifact.sign_off`.
