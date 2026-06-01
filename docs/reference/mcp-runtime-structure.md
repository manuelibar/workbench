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
- file-backed artifact storage
- artifact validation and artifact-specific classified errors

MCP handlers translate between protocol request/response payloads and
`internal/artifacts` commands/results. Context-selection fields such as
`select` or defaulting an update to the selected artifact stay in `internal/mcp`.

Tool files in `internal/mcp` own their protocol-facing metadata and typed
handler:

- short tool name and optional group prefix
- description
- typed input and output structs inferred by the MCP SDK as JSON schemas
- runtime behavior

Each tool registers itself at package init time. The registry is only the
compiled capability catalog; the context planner still decides which tools are
active on the MCP server surface.

`internal/mcp/resources` owns resource descriptors, URI conventions, and
embedded static Markdown. Runtime read handlers remain in `internal/mcp`
because they depend on context and artifact stores.

Definitions use MCP-native identity where possible: tool name, resource URI,
and resource template URI. The selected artifact resource is a contextual slot
whose concrete URI and display metadata are built from the selected artifact at
registration time.

`internal/jsonschema` owns small Workbench schema primitives over
`github.com/google/jsonschema-go/jsonschema` for cases where reflected schemas
are not expressive enough.

## Dependency Direction

Descriptor packages do not import `internal/mcp`.

The dependency direction is:

```text
internal/mcp -> internal/artifacts
internal/mcp -> internal/mcp/resources
```

This keeps protocol metadata and behavior decentralized while avoiding
speculative ports or interfaces. Add an interface only when a package consumes
behavior that needs real substitution.

## Adding A Tool

1. Add a `tool_*.go` file under `internal/mcp`.
2. Implement `Name`, `Group`, `Description`, and typed `Handle`.
3. Register the tool from the file with `registerTool[Input, Output]`.
4. Add focused tests for registration validity and behavior.

## Adding A Resource

1. Add a descriptor under `internal/mcp/resources/`.
2. Add URI parsing or construction helpers there if the URI has structure.
3. Add static Markdown beside the descriptor when the content is build-time
   knowledge.
4. Add the descriptor to `resources.All()` or `resources.Templates()`.
5. Bind the descriptor to the runtime read handler in `internal/mcp`.

Generated daily artifacts stay on disk under `docs/artifacts/`; only static
resource bodies, prompt bodies, and templates should be embedded.

## Future MCP Surface

Prompts, sampling, and elicitations should follow the same rule: static bodies
can live beside their capability code, while runtime orchestration and
client/session-dependent behavior stay in `internal/mcp`.
