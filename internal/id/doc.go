// Package id holds the workbench's identifier helpers: UUIDv7 generation
// plus the four-part audit schema (request_id, correlation_id, causation_id,
// idempotency_key) carried on every state-bearing row and propagated through
// every tool invocation via [context.Context].
//
// The four-part schema is taken verbatim from the manifesto's
// docs/reference/id-schema.md. Generation is centralised here so that pgstore
// rows, MCP middleware, and event records never disagree about the format.
package id
