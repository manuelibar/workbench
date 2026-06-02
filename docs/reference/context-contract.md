# Context Contract

## Tool

`contextualize` is the only context mutation entry point on `main`.

Input fields:

| Field | Type | Behavior |
|---|---|---|
| `focus` | string, null, or omitted | omitted preserves, null clears, string sets |
| `artifact_id` | string, null, or omitted | omitted preserves, null clears, string selects an existing artifact |

Unknown fields are rejected.

## Result

`contextualize` returns:

- `context_document`: the raw `ContextDocument`, byte-for-byte equal to the
  `workbench:///context` resource.
- `focus`: current selected focus, omitted when empty.
- `artifact_id`: current selected artifact id, omitted when empty.
- `sync`: changed list categories, observed list calls, status, generation,
  and timeout flag.
- `fallback_capabilities`: current MCP surface snapshot when sync times out.

The context document is intentionally state-focused. Capability listings belong
to MCP list responses, not duplicated in `workbench:///context`.

## Sync

Capability categories map to MCP list methods:

| Category | MCP list method |
|---|---|
| `tools` | `tools/list` |
| `resources` | `resources/list` |
| `prompts` | `prompts/list` |

The default sync timeout is `5s` and can be changed with
`WORKBENCH_CONTEXT_SYNC_TIMEOUT`. When the client does not observe the required
list methods before the timeout, `contextualize` returns
`fallback_capabilities` so the agent can recover without stale local state.
