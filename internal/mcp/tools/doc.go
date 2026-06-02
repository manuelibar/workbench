// Package tools defines Workbench MCP tool descriptors, handlers, and the
// protocol payloads owned by those handlers.
//
// Tools register themselves at init time. The MCP server supplies live behavior
// through the Host interface when a registered tool is bound.
package tools
