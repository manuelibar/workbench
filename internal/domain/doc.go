// Package domain holds pure types shared by the storage and server layers:
// users, work sessions, selection state, and (in later phases) namespaces,
// projects, blueprints, modes, artifacts, notes, skills, prompts, events.
//
// This package has no dependencies on pgx, the MCP SDK, or net/http; it is
// the boundary type set that both [github.com/manuelibar/workbench/internal/pgstore]
// and [github.com/manuelibar/workbench/internal/mcpserver] consume.
package domain
