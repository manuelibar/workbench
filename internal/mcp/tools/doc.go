// Package tools defines Workbench MCP tool descriptors.
//
// A descriptor owns the MCP-facing identity, visibility, description, and JSON
// schemas for one tool. Runtime handlers remain in package mcp, where the
// context store, artifact store, sync boundary, and error boundary live.
package tools
