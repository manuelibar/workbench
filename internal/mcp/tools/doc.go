// Package tools defines Workbench MCP tool descriptors, handlers, and the
// protocol payloads owned by those handlers.
//
// Tools register themselves at init time. The MCP runtime supplies live server
// behavior through the Runtime interface when a registered tool is bound.
package tools
