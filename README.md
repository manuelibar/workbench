# workbench

Workbench is a local stdio MCP server for one agent process. `main` is now
the small context and artifact kernel: it keeps current agent context in
memory, stores typed Markdown artifacts under `docs/artifacts/`, and changes
the MCP surface through a deterministic `context` tool.

Backlog, notes, namespaces, roles, memory, knowledge, sessions, AFK, skills,
and Artifact B2 workflows continue on epic branches that define their contracts
through self-contained docs packets before runtime work.

## Quickstart

```bash
make test
make build
WORKBENCH_LOG_LEVEL=warn ./build/_output/workbench-mcp
```

Configure an MCP client to launch the binary over stdio. The server has no
HTTP listener and no database requirement.

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
tool result includes a full fallback capability index so the agent can recover
without relying on stale local state.

## Configuration

| Env var | Default | Notes |
|---|---|---|
| `WORKBENCH_ARTIFACT_DIR` | `docs/artifacts` | Flat Markdown artifact directory |
| `WORKBENCH_CONTEXT_SYNC_TIMEOUT` | `5s` | Duration or integer seconds |
| `WORKBENCH_LOG_LEVEL` | `info` | `debug`, `info`, `warn`, or `error` |

## Documentation

Start with [docs/README.md](docs/README.md). The foundation docs cover:

- context-window thesis
- progressive disclosure
- context contract
- artifact conventions
- epic branch workflow

## Verification

```bash
go test ./...
go test -race ./...
```
