// Package mcp implements the Workbench MCP runtime kernel.
//
// It owns server wiring, context state, capability planning, sync, and handler
// execution. Artifact contracts and Markdown persistence live in
// internal/artifacts. MCP surface metadata lives in child packages such as tools
// and resources so new protocol features can be added without concentrating
// every schema and description in the server.
package mcp
