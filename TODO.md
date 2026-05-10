# TODO

Living list of work deferred from v0.1.0. Grouped by horizon — top of each
section is highest priority.

## v0.1.x — polish on the current surface

- **Repo hygiene.** No `git init` yet. Once initialised, push `v0.1.0` as
  the first tag and add the GitHub remote.
- **godoc pass.** Every exported symbol has a godoc comment, but a strict
  pass (consistent voice, complete sentences, `[SymbolName]` cross-refs)
  hasn't been done over the whole tree. Useful before any external reader
  shows up.
- **`example_test.go` for `internal/mcpserver`.** A runnable doc example
  that boots a server, connects an in-memory client, and exercises
  `refresh` + `note.add` would make the package's intended use obvious.
- **`select_blueprint` by name + version.** Today `refresh(blueprint_id=…)`
  expects the UUID. Convenience aliases `blueprint_name` /
  `blueprint_version` would let the agent select a blueprint without
  first looking up its UUID.
- **Note search: query trimming.** `note.search` errors on a whitespace
  query — fine, but a clearer error message saying so would help the
  agent self-correct.
- **`backlog.take_next` ordering.** v0 reuses `ListArtifacts` (which is
  `ORDER BY updated_at DESC`) and walks backwards. A dedicated query
  (`ORDER BY created_at ASC LIMIT 1`) belongs in `pgstore`. Trivial swap.
- **Resource: `workbench:///notes/{id}`.** Listed in the README's
  resource surface but not yet implemented. Add a handler that fetches
  the note by id and returns the markdown body.
- **WorkSession rotation.** v0 has a singleton open WorkSession and no
  rotation tool. Once the user wants per-day boundaries, add
  `worksession.close` (always-on) and have the server open a new one on
  next boot. Schema already supports this via the partial unique index.
- **Events pagination + filtering.** `pgstore.RecentEvents` returns at
  most 200 events of any type. A `since` / `until` filter and a
  `type IN (...)` filter would let `status`-style views work.

## v0.2 — features the kernel implies but didn't ship

- **Bearer-token auth.** v0 binds to `127.0.0.1` with no auth. Next
  iteration: `WORKBENCH_TOKEN` env var, SHA-256 hashed, constant-time
  compared. The `auth/` package already has the skeleton.
- **User-authored prompts via `prompts/list_changed`.** Today
  `prompt.create` writes to Postgres but the MCP `prompts/list` surface
  only ever returns an empty list. Wire `Server.AddPrompt` /
  `Server.RemovePrompts` into the selection-driven flow so the agent can
  use `prompts/get` directly.
- **Mode-driven surface narrowing.** v0 lets the user record
  `capabilities` on a mode but doesn't actually filter the tool surface
  based on them. Once a mode is selected, the surface should honour the
  capability list (subset of always-on + scoped tools, plus any
  user-defined tools/resources/prompts).
- **`refresh` sync-wait.** Phase 5 deliberately skipped the "block until
  the host's `tools/list` re-call arrives" mechanic per the
  avoid-races directive. If the small race window proves problematic in
  practice, add an inbound-method observer that closes a per-session ack
  channel; refresh waits up to ~2s before returning `synced: true/false`.
- **Notes promotion workflow.** The `note.promoted_to` column is in the
  schema but no tool / mode uses it. Build a mode whose workflow walks
  the inbox and materialises each note into a typed artifact, linking
  back via that column.
- **Semantic search (pgvector).** Migration 0002 created the extension
  but no embeddings table exists. Add an `embeddings` table, an
  embedding service interface (HTTP boundary), and rewrite
  `note.search` + a new `recall` tool to use RRF (BM25 + vector).
- **Conventional Commits trailers.** Manifesto reference exists; nothing
  in v0 writes them. Add a `git_commit` tool plus `Ripple-Run-ID`,
  `Ripple-Mode`, `Ripple-Correlation`, … trailers (only relevant once
  the Execution Engine lands).
- **Multi-WorkSession.** v0 enforces one open session per user via a
  partial unique index. Multi-session unlocks parallel daily threads
  (e.g. work + personal). Mostly a schema relaxation + a session
  picker tool.

## Out of scope until further notice

(Mentioned in the plan; not in v0.)

- **Live agent-to-agent topics.** Cross-machine pub/sub channels for
  collaborative agent discussions. Needs a separate transport layer
  (probably the NATS bus the Execution Engine will host).
- **Headless Claude / Codex executions from inside the MCP.** A tool
  that shells out to `claude -p ...` or `codex ...` for sub-agent
  delegation. Easy to add later; needs to live behind a configurable
  binary path + execution sandbox.
- **NATS execution engine + AFK runs.** Manifesto's `Workbench / Core /
  Execution Engine` triad — only Workbench is in this repo. Core
  (HTTP) and Execution Engine (NATS JetStream) are separate repos.
- **`ripplectl` CLI.** A management CLI for namespace/project/blueprint
  ops, alternative to driving everything through MCP. Lower priority
  since the MCP surface already does what the CLI would.
- **Agent-pick tool registration** (`tools.search` / `tools.register`).
  Discussed and deferred; selection-driven surface seems sufficient.
- **OAuth / multi-tenant / ABAC.** Workbench is single-user local-first
  by design; this is a Phase 6+ concern from the manifesto.
- **`workbench://status` resource and a separate `status` tool.**
  Folded into `refresh`'s response inline. Re-add only if a passive
  read path becomes valuable.

## Manifesto features intentionally not started

These are bigger-than-one-iteration commitments from
`github.com/manuelibar/ripple` that v0 does *not* claim to address.

- **Three-phase iteration** (Prepare → Execute → Wrap-up). Execution
  Engine concern.
- **Run lifecycle, Iteration table, Signal vocabulary.** Same.
- **Persistent-session ("Ralph") mode.** Phase 4 of the manifesto roadmap.
- **Sandbox per iteration / per run.** Docker / gVisor / Firecracker.
  Execution Engine concern.
- **OTel export.** Slog-only for v0.
- **Knowledge-graph layer.** NTH in the manifesto.
- **Scheduling / trigger policies.** AFK Phase 3.
- **Cost / token budget enforcement.** Phase 3 once Run plumbing exists.

## Cross-cutting watch-list

- **MCP spec churn.** Spec revision `2026-06-30` is in flight (HTTP
  header standardisation, OAuth `application_type`). The SDK is already
  absorbing pieces; revisit when v1.7.x lands.
- **`MCPGODEBUG` toggles.** `disablelocalhostprotection` and
  `enableoriginverification` are scheduled for removal in SDK v1.8.0.
  We don't depend on either, but document this if anyone adds env vars
  for them.
- **Output schema vs `uuid.UUID`.** jsonschema-go reflects `*uuid.UUID`
  as `[16]byte`, producing a schema that doesn't match the runtime
  string form. v0 sidesteps this by using `string` IDs in every wire
  type. If we add new tool result types, keep doing that — or, better,
  contribute a `uuid.UUID` reflector to `google/jsonschema-go`.
