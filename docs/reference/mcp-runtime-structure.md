# MCP Runtime Structure

Workbench keeps MCP transport wiring separate from MCP surface definitions.
The goal is a scalable package shape, not a strict layered architecture.

## Ownership

`internal/mcp` owns the runtime kernel:

- server construction and MCP SDK wiring
- context state and patch parsing
- artifact tool/resource orchestration
- capability planning and sync
- MCP error boundary handling
- handler execution

`internal/artifacts` owns the artifact domain:

- typed artifact contracts and section specs
- Markdown rendering and parsing
- file-backed artifact storage and the storage-service artifact adapter
- artifact validation and artifact-specific classified errors

`internal/storageclient` owns the narrow HTTP client used to consume the
autonomous storage service. It mirrors only the storage API fields Workbench
needs for artifact Markdown and does not contain S3, normalization, indexing,
or service runtime code.

MCP handlers translate between protocol request/response payloads and
`internal/artifacts` commands/results. Context-selection fields such as
`select` or defaulting an update to the selected artifact are implemented by
tool handlers but call back into the runtime through a narrow interface.

`internal/mcp/tools` owns tool protocol-facing metadata and typed
handler:

- short tool name and optional group prefix
- description
- typed input and output structs inferred by the MCP SDK as JSON schemas,
  kept beside the tool handler that owns them
- runtime behavior

Each tool registers itself at package init time. The tools registry is only the
compiled tool catalog; the context planner still decides which tools are active
on the MCP server surface.

`internal/mcp/resources` owns resource descriptors, URI conventions, and
embedded static Markdown. Each resource registers itself at package init time.
Runtime read handlers remain in `internal/mcp` because they depend on context
and artifact stores.

Definitions use MCP-native identity where possible: tool name and resource URI.
The selected artifact resource is a contextual slot whose concrete URI and
display metadata are built from the selected artifact at registration time.

`internal/jsonschema` owns small Workbench schema primitives over
`github.com/google/jsonschema-go/jsonschema` for cases where reflected schemas
are not expressive enough.

## Dependency Direction

Descriptor packages do not import `internal/mcp`.

The dependency direction is:

```text
internal/mcp -> internal/artifacts
internal/mcp -> internal/mcp/tools
internal/mcp -> internal/mcp/resources
internal/artifacts -> internal/storageclient
```

This keeps protocol metadata and behavior decentralized while avoiding
speculative ports or interfaces. Add an interface only when a package consumes
behavior that needs real substitution.

## Adding A Tool

1. Add a Go file under `internal/mcp/tools`, named for the tool concept
   such as `artifact_begin.go` or `contextualize.go`.
2. Implement `Name`, `Group`, `Description`, and typed `Handle`.
3. Register the tool from the file with `register[Input, Output]`.
4. Add focused tests for registration validity and behavior.

## Adding A Resource

1. Add a descriptor under `internal/mcp/resources/`.
2. Add URI parsing or construction helpers there if the URI has structure.
3. Add static Markdown beside the descriptor when the content is build-time
   knowledge.
4. Register the descriptor from the file.
5. Bind the descriptor to the runtime read handler in `internal/mcp`.

Generated daily artifacts stay on disk under `docs/artifacts/` unless
Workbench is configured with `WORKBENCH_STORAGE_URL`; only static resource
bodies, prompt bodies, and templates should be embedded.

## Future MCP Surface

Prompts, sampling, and elicitations should follow the same rule: static bodies
can live beside their capability code, while runtime orchestration and
client/session-dependent behavior stay in `internal/mcp`.
