#!/usr/bin/env bash
# scripts/smoke.sh — boot the workbench-mcp binary and verify the live HTTP
# surface: /healthz, /readyz, and an MCP `initialize` round-trip over the
# streamable HTTP transport. Tool-surface coverage lives in the Go
# integration tests (`go test ./...`); this script is the "real wire"
# sanity check for releases.
#
# Requires: a healthy docker-compose stack (use `make compose-up`).

set -euo pipefail

cd "$(dirname "$0")/.."

BIND="${WORKBENCH_SMOKE_BIND:-127.0.0.1:7780}"
URL="http://${BIND}"
BACKLOG_URL="${WORKBENCH_BACKLOG_URL:-http://127.0.0.1:7778}"

echo "[smoke] building workbench-mcp..."
go build -o build/_output/workbench-mcp ./cmd/workbench-mcp

echo "[smoke] starting binary on ${BIND} (backlog at ${BACKLOG_URL})..."
WORKBENCH_BIND="${BIND}" WORKBENCH_LOG_LEVEL=warn WORKBENCH_BACKLOG_URL="${BACKLOG_URL}" \
  ./build/_output/workbench-mcp >/tmp/workbench-smoke.log 2>&1 &
PID=$!
HEADERS="$(mktemp)"
cleanup() {
  kill -INT "$PID" 2>/dev/null || true
  wait "$PID" 2>/dev/null || true
  rm -f "$HEADERS"
}
trap cleanup EXIT

echo "[smoke] waiting for /healthz..."
for _ in $(seq 1 60); do
  if curl -sf "${URL}/healthz" >/dev/null; then
    break
  fi
  sleep 0.25
done
curl -sf "${URL}/healthz" >/dev/null || { echo "FAIL: /healthz never came up"; cat /tmp/workbench-smoke.log; exit 1; }
echo "  /healthz OK"

echo "[smoke] /readyz..."
curl -sf "${URL}/readyz" >/dev/null || { echo "FAIL: /readyz returned non-2xx"; cat /tmp/workbench-smoke.log; exit 1; }
echo "  /readyz OK"

echo "[smoke] MCP initialize..."
curl -sS -D "$HEADERS" -o /dev/null -X POST "${URL}/mcp" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json, text/event-stream" \
  -H "MCP-Protocol-Version: 2025-11-25" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"smoke","version":"0"}}}'

SID="$(grep -i '^Mcp-Session-Id:' "$HEADERS" | awk '{print $2}' | tr -d '\r')"
if [ -z "$SID" ]; then
  echo "FAIL: initialize did not return Mcp-Session-Id"
  cat "$HEADERS"
  cat /tmp/workbench-smoke.log
  exit 1
fi
echo "  session: ${SID}"

# Optional backlog round-trip: only if the separate backlog-service is up
# on BACKLOG_URL. Probes /readyz first so this stays a non-fatal addition.
if curl -fsS "${BACKLOG_URL}/readyz" >/dev/null 2>&1; then
  echo "[smoke] backlog.add via MCP tools/call..."
  TOOL_RESP=$(curl -sS -X POST "${URL}/mcp" \
    -H "Content-Type: application/json" \
    -H "Accept: application/json, text/event-stream" \
    -H "MCP-Protocol-Version: 2025-11-25" \
    -H "Mcp-Session-Id: ${SID}" \
    -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"backlog.add","arguments":{"project_id":"smoke","title":"smoke ran"}}}')
  echo "${TOOL_RESP}" | grep -q '"smoke ran"' || {
    echo "FAIL: backlog.add did not echo the title"
    echo "${TOOL_RESP}"
    cat /tmp/workbench-smoke.log
    exit 1
  }
  echo "  backlog.add OK"
else
  echo "[smoke] backlog-service not running on ${BACKLOG_URL}; skipping backlog round-trip"
fi

echo "[smoke] OK"
