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
	"github.com/manuelibar/workbench/internal/storage"
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

	opts := mcp.Options{
		ArtifactDir: artifactDir(),
		SyncTimeout: syncTimeout(),
		Logger:      logger,
	}
	if client, ok, err := storageClient(); err != nil {
		logger.Error("workbench storage init failed", "err", err)
		return 1
	} else if ok {
		opts.StorageClient = client
		opts.StorageOrgID = envDefault("WORKBENCH_STORAGE_ORG_ID", "local")
		opts.StorageProjectID = envDefault("WORKBENCH_STORAGE_PROJECT_ID", "workbench")
		opts.StorageResourceType = envDefault("WORKBENCH_STORAGE_RESOURCE_TYPE", "artifacts")
	}
	server, err := mcp.New(opts)
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

func storageClient() (*storage.Client, bool, error) {
	baseURL := strings.TrimSpace(os.Getenv("WORKBENCH_STORAGE_URL"))
	if baseURL == "" {
		return nil, false, nil
	}
	client, err := storage.NewClient(storage.ClientOptions{
		BaseURL: baseURL,
		Timeout: durationEnv("WORKBENCH_STORAGE_TIMEOUT", 30*time.Second),
	})
	if err != nil {
		return nil, false, err
	}
	return client, true, nil
}

func durationEnv(name string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	d, err := time.ParseDuration(raw)
	if err == nil {
		return d
	}
	if seconds, err := strconv.Atoi(raw); err == nil {
		return time.Duration(seconds) * time.Second
	}
	return fallback
}

func envDefault(name, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
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
