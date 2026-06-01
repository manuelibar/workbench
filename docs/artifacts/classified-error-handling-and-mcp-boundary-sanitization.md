---
id: "classified-error-handling-and-mcp-boundary-sanitization"
type: "adr"
title: "Classified Error Handling And MCP Boundary Sanitization"
status: "accepted"
created: "2026-06-01T00:00:00Z"
updated: "2026-06-01T00:00:00Z"
---

# Classified Error Handling And MCP Boundary Sanitization

## Context

Workbench returns errors from several layers: context patch parsing, artifact
lookup and storage, capability planning, and MCP resource or tool calls. Raw
low-level errors can be useful for local debugging, but they are not a stable
client contract and can expose private details such as filesystem paths,
wrapped cause text, or internal operation names.

The MCP server is the only transport boundary in this pass. Tools should report
ordinary execution failures as tool results with `isError=true`, while resource
reads report failures as JSON-RPC errors.

## Decision

Errors are classified at the layer that understands the failure, decorated with
private context while they bubble up, and sanitized once at the MCP boundary.

Workbench uses the internal `errs` package for programmatic classes, stable
codes, retryability, severity, private attributes, causes, and captured frames.
The MCP receiving middleware is responsible for converting classified errors
into public client responses.

For `tools/call`, classified failures return `isError=true` with a public title
and structured content containing only `title`, `code`, and `retryable`. For
`resources/read`, classified failures map to JSON-RPC codes:

- `ErrNotFound` maps to `mcp.CodeResourceNotFound`.
- `ErrInvalid` maps to `jsonrpc.CodeInvalidParams`.
- Other classified failures map to `jsonrpc.CodeInternalError`.

SDK-origin argument validation remains unchanged so clients can self-correct
from the SDK validation message.

## Consequences

Error construction is more explicit at failure boundaries. Callers can use
`errors.Is` for coarse classes and stable Workbench codes for logs and client
behavior.

Client responses no longer expose attributes, causes, stack frames, filesystem
paths, private URIs, or wrapped low-level error text. Server logs can still use
private attributes and the original classified error for debugging.

Future handlers must avoid returning raw low-level errors directly when the
failure is understood. They should classify the failure at the source, decorate
it while adding context, and let the MCP boundary decide what is public.

## Alternatives

Return raw errors directly from handlers. This keeps handlers small, but it
makes client behavior depend on unstable strings and risks private leakage.

Classify errors only in MCP middleware. This centralizes response handling, but
the boundary often lacks the domain context needed to distinguish invalid
input, missing artifacts, dependency failures, and temporary unavailability.

Add a shared external error dependency. This was rejected for the foundation
because Workbench only needs a small internal package and should stay
dependency-free for error handling.
