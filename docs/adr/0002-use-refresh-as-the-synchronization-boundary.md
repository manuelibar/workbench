# 2. Use refresh as the synchronization and navigation boundary

Date: 2026-05-19

## Status

Accepted

## Context

Workbench is an MCP harness/facade where namespace, project, role, board, task, and session selection change the context and capabilities available to an agent. MCP clients need one reliable operation that tells the server to apply the requested scope and reconcile the dynamic surface.

Earlier drafts treated `refresh` as a way to return a full context pack including inline skills and capability links. That is too much responsibility for a single tool response and duplicates MCP's native capability-discovery model.

## Decision

`refresh` is the permanent synchronization and navigation tool.

Clients call it:

- at session start
- after context compaction
- when namespace/project/role/task selection changes
- after a meaningful mutation that may affect available context

On each call, Workbench:

1. Applies the requested selection through the configured session/scope manager.
2. Reconciles manager-owned resources, tools, and prompts for the selected scope.
3. Emits MCP list-changed notifications for each capability family that changed.
4. Waits for observed client `tools/list` / `resources/list` calls on the one-to-one stdio connection, bounded by a timeout.
5. Returns a concise navigation briefing plus capability sync status. If the client did not relist before timeout, the response includes a fallback capability index.

The refresh response should orient the agent, not replace MCP discovery under normal conditions. It may include selected scope, a short project/context summary, recent task or knowledge highlights, and recommended resource URIs. The authoritative capability surface is discovered through MCP list methods after notifications. If Workbench cannot observe those list calls before timeout, it returns a temporary inline capability index as a compatibility fallback.

Entity mutation tools such as project, namespace, task, role, board, feedback, or knowledge operations are not the synchronization boundary. They mutate state. `refresh` reconciles the working surface.

## Consequences

Agents get a simple mental model: call `refresh`, then let the MCP client list capabilities and read resources. Workbench can add or remove scoped tools/resources/prompts without stuffing every capability into the refresh payload.

The main risk is client compatibility. MCP clients must handle list-changed notifications and refresh visible capability lists. Workbench now detects that behavior when possible. If a client cannot do that yet, Workbench includes a fallback index in `refresh()` rather than blocking indefinitely.
