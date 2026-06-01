# workbench

Workbench is a local stdio MCP server for one agent process. `main` is now
the small context and artifact kernel: it keeps current agent context in
memory, stores typed Markdown artifacts either under `docs/artifacts/` or
through the storage service, and changes the MCP surface through a
deterministic `context` tool.

MCP tool implementations are decentralized in self-registering `internal/mcp`
files, while resources are defined under `internal/mcp/resources`; the runtime
kernel binds those capabilities to sync and error boundaries. Artifact
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
tool result includes fallback capabilities so the agent can recover
without relying on stale local state.

## Configuration

| Env var | Default | Notes |
|---|---|---|
| `WORKBENCH_ARTIFACT_DIR` | `docs/artifacts` | Flat Markdown artifact directory |
| `WORKBENCH_STORAGE_URL` | unset | When set, Workbench stores artifacts through the storage service instead of the filesystem |
| `WORKBENCH_STORAGE_ORG_ID` | `local` | Storage org namespace for Workbench artifacts |
| `WORKBENCH_STORAGE_PROJECT_ID` | `workbench` | Storage project namespace for Workbench artifacts |
| `WORKBENCH_STORAGE_RESOURCE_TYPE` | `artifacts` | Generic storage resource type used for artifacts |
| `WORKBENCH_STORAGE_TIMEOUT` | `30s` | Storage HTTP client timeout |
| `WORKBENCH_CONTEXT_SYNC_TIMEOUT` | `5s` | Duration or integer seconds |
| `WORKBENCH_LOG_LEVEL` | `info` | `debug`, `info`, `warn`, or `error` |

## Storage Service

`cmd/storage-service` is a generic S3-backed Markdown storage service. It does
not know about Workbench artifacts. It accepts resource uploads for multiple
MIME types, uses MarkItDown to normalize each object into Markdown, prepends a
YAML byte-offset index, and stores the indexed Markdown at:

```text
s3://<bucket>/<org_id>/<project_id>/<resource_type>/<resource_id>.md
```

Raw uploads go through presigned S3 URLs under:

```text
s3://<bucket>/inbox/<org_id>/<project_id>/<raw_filename.ext>
```

The service exposes:

- `POST /api/storage/upload-url`
- `GET /api/storage/resources`
- `GET /api/storage/{id}/stats`
- `GET /api/storage/{id}/download-url`
- `POST /api/storage/{id}/update-url`
- `POST /api/storage/multipart/start`
- `POST /internal/webhook/s3-event`

Storage service configuration:

| Env var | Default | Notes |
|---|---|---|
| `STORAGE_BUCKET` | required | S3 bucket with versioning enabled |
| `STORAGE_ADDR` | `:8080` | HTTP listen address |
| `STORAGE_PRESIGN_TTL` | `15m` | Presigned URL lifetime |
| `STORAGE_MARKITDOWN_COMMAND` | `markitdown` | MarkItDown executable |
| `STORAGE_TEMP_DIR` | system temp | Temporary raw-file staging directory |
| `STORAGE_LOG_LEVEL` | `info` | `debug`, `info`, `warn`, or `error` |

AWS credentials and region are loaded with the AWS SDK default chain. The S3
bucket must have object versioning enabled and CORS rules that allow the
presigned PUT/GET flows used by the client.

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
