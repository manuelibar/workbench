# Scope Contract

## Tool

`contextualize` is the only scope mutation entry point on `main`.

Input fields:

| Field | Type | Behavior |
|---|---|---|
| `focus` | string, null, or omitted | omitted preserves, null clears, string sets |
| `artifact_id` | string, null, or omitted | omitted preserves, null clears, string places an existing artifact in scope |

Unknown fields are rejected.

When `artifact_id` is set to a string, Workbench validates that the artifact
exists, materializes its full Markdown to a server-managed local scoped file,
and exposes `workbench:///artifacts/<id>` as the artifact resource. Storage is
unchanged until `artifact.upload` persists a full Markdown replacement.

## Result

`contextualize` returns:

- `scope_document`: the raw scope document, byte-for-byte equal to the
  `workbench:///scope` resource.
- `focus`: current focus, omitted when empty.
- `artifact_id`: current artifact id, omitted when empty.
- `sync`: changed list categories, observed list calls, status, generation,
  and timeout flag.
- `fallback_capabilities`: current MCP surface snapshot when sync times out.

The scope document is intentionally state-focused. Capability listings belong
to MCP list responses, not duplicated in `workbench:///scope`.

## Sync

Capability categories map to MCP list methods:

| Category | MCP list method |
|---|---|
| `tools` | `tools/list` |
| `resources` | `resources/list` |
| `prompts` | `prompts/list` |

The default sync timeout is `5s` and can be changed with
`WORKBENCH_SCOPE_SYNC_TIMEOUT`. When the client does not observe the required
list methods before the timeout, `contextualize` returns
`fallback_capabilities` so the agent can recover without stale local state.
