// Command workbench-mcp runs one stdio MCP Workbench process.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/manuelibar/workbench/internal/mcp"
)

func main() {
	os.Exit(run())
}

func run() int {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: parseLogLevel(os.Getenv("WORKBENCH_LOG_LEVEL")),
	}))
	slog.SetDefault(logger)

	rootCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	server, err := mcp.New(mcp.Options{
		ArtifactDir: artifactDir(),
		SyncTimeout: syncTimeout(),
		Logger:      logger,
	})
	if err != nil {
		logger.Error("workbench init failed", "err", err)
		return 1
	}
	if err := server.Run(rootCtx, &mcpsdk.StdioTransport{}); err != nil && rootCtx.Err() == nil {
		logger.Error("workbench stopped with error", "err", err)
		return 1
	}
	return 0
}

func artifactDir() string {
	if dir := strings.TrimSpace(os.Getenv("WORKBENCH_ARTIFACT_DIR")); dir != "" {
		return dir
	}
	return filepath.Join("docs", "artifacts")
}

func syncTimeout() time.Duration {
	raw := strings.TrimSpace(os.Getenv("WORKBENCH_CONTEXT_SYNC_TIMEOUT"))
	if raw == "" {
		return 5 * time.Second
	}
	d, err := time.ParseDuration(raw)
	if err == nil {
		return d
	}
	if seconds, err := strconv.Atoi(raw); err == nil {
		return time.Duration(seconds) * time.Second
	}
	return 5 * time.Second
}

func parseLogLevel(raw string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
