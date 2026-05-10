// Package mcpserver implements the workbench's MCP front door.
//
// At the package boundary, [New] constructs a [Server] from a [pgstore.Store]
// and the bootstrapped [domain.User] / [domain.WorkSession]. The Server
// exposes [Server.Handler] for mounting on a net/http mux and
// [Server.NewSessionServer] for in-memory testing.
//
// The package's responsibilities split into:
//
//   - lifecycle and per-protocol-session [*mcp.Server] construction (this
//     file plus mcpserver.go);
//   - tool handlers — refresh.go, ask.go, and (later) the `tools/` subpackage
//     for each CRUD-namespaced resource;
//   - resource handlers — resources.go (always-on `workbench://skill` plus
//     templated namespace/project/artifact resolutions in later phases);
//   - middleware — `middleware/` subpackage for ids, slog, events.
//
// Phase 3 ships the always-on `refresh` and `ask` tools plus the embedded
// onboarding `workbench://skill` resource. Selection-driven surface
// mutation arrives in phase 5.
package mcpserver
