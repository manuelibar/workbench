// Package middleware holds the workbench's MCP receiving middlewares. v0
// ships:
//
//   - [IDs] — extracts (or generates) the four-part audit identifiers
//     (request_id, correlation_id, causation_id, idempotency_key) and
//     attaches them to the request [context.Context].
//   - [Slog] — structured logging of every inbound request (method,
//     session id, request id, duration, error).
//
// Phase 4 will add an `events` middleware that records every tool call to
// the events table for episodic memory.
package middleware
