# workbench

Workbench is a local stdio MCP server for one agent process. `main` is now
the small context and artifact kernel: it keeps current agent context in
memory, stores typed Markdown artifacts either under `docs/artifacts/` or
through the external storage HTTP service, and changes the MCP surface through
a deterministic `context` tool.

MCP tools live in the self-registering `internal/mcp/tools` package, while
resources live in the self-registering `internal/mcp/resources` package; the
runtime kernel binds those capabilities to sync and error boundaries. Artifact
contracts and Markdown persistence live in `internal/artifacts`.

Backlog, notes, namespaces, roles, memory, knowledge, sessions, AFK, skills,
and Artifact v2 workflows continue on epic branches that define their contracts
through self-contained docs packets before runtime work.

## Quickstart

```bash
make test
make build
WORKBENCH_LOG_LEVEL=warn ./build/_output/workbench-mcp
```

Configure an MCP client to launch the binary over stdio. The server has no
HTTP listener, database requirement, S3 client, or MarkItDown runtime.

## MCP Surface

Always available:

- `context`
- `artifact.begin`
- `artifact.list`
- `artifact.get`
- resource `workbench:///context`
- resource template `workbench:///artifacts/{id}`

Available after `context` selects an `artifact_id`:

- `artifact.update`
- `artifact.guidance`
- `artifact.validate`
- resource `workbench:///artifacts/<selected-id>`

`context` accepts tri-state patch fields:

- omitted field: preserve the current value
- `null`: clear the value
- string: set the value

Current fields are `focus` and `artifact_id`.

When a context change alters tool or resource visibility, Workbench emits MCP
list-changed notifications and waits for the changed capability lists to be
observed before returning. The default wait is five seconds. On timeout, the
tool result includes fallback capabilities so the agent can recover
without relying on stale local state.

## Configuration

| Env var | Default | Notes |
|---|---|---|
| `WORKBENCH_ARTIFACT_DIR` | `docs/artifacts` | Flat Markdown artifact directory |
| `WORKBENCH_STORAGE_URL` | unset | When set, Workbench stores artifacts through the autonomous storage service |
| `WORKBENCH_STORAGE_ORG_ID` | `local` | Storage org namespace for Workbench artifacts |
| `WORKBENCH_STORAGE_PROJECT_ID` | `workbench` | Storage project namespace for Workbench artifacts |
| `WORKBENCH_STORAGE_RESOURCE_TYPE` | `artifacts` | Generic storage resource type used for artifacts |
| `WORKBENCH_STORAGE_TIMEOUT` | `30s` | Storage HTTP client timeout |
| `WORKBENCH_CONTEXT_SYNC_TIMEOUT` | `5s` | Duration or integer seconds |
| `WORKBENCH_LOG_LEVEL` | `info` | `debug`, `info`, `warn`, or `error` |

## Storage Service

The storage service lives outside this repository at `../storage`. Workbench
does not own S3, normalization, or the storage HTTP server; it only uses the
storage API to list, read, and write artifact Markdown when
`WORKBENCH_STORAGE_URL` is configured.

## Documentation

Start with [docs/README.md](docs/README.md). The foundation docs cover:

- context-window thesis
- progressive disclosure
- context contract
- MCP runtime structure
- artifact conventions
- epic branch workflow

## Verification

```bash
go test ./...
go test -race ./...
```
