#!/usr/bin/env bash
# Build the stdio MCP server and run the command-transport integration smoke.

set -euo pipefail

cd "$(dirname "$0")/.."

echo "[smoke] building workbench-mcp..."
go build -o build/_output/workbench-mcp ./cmd/workbench-mcp

echo "[smoke] running stdio MCP integration..."
go test ./cmd/workbench-mcp -run TestStdioMCPContextIntegration -count=1

echo "[smoke] OK"
