// Command storage-service runs the generic S3-backed Markdown storage API.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/manuelibar/workbench/internal/storage"
	"github.com/manuelibar/workbench/internal/storage/s3store"
)

func main() {
	os.Exit(run())
}

func run() int {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: parseLogLevel(os.Getenv("STORAGE_LOG_LEVEL")),
	}))
	slog.SetDefault(logger)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	bucket := strings.TrimSpace(os.Getenv("STORAGE_BUCKET"))
	if bucket == "" {
		logger.Error("storage init failed", "err", "STORAGE_BUCKET is required")
		return 1
	}
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Error("storage init failed", "err", err)
		return 1
	}
	objectStore, err := s3store.New(bucket, s3.NewFromConfig(awsCfg))
	if err != nil {
		logger.Error("storage init failed", "err", err)
		return 1
	}
	service, err := storage.NewService(storage.ServiceOptions{
		Objects: objectStore,
		Normalizer: storage.MarkItDownNormalizer{
			Command: envDefault("STORAGE_MARKITDOWN_COMMAND", "markitdown"),
			TempDir: strings.TrimSpace(os.Getenv("STORAGE_TEMP_DIR")),
		},
		PresignTTL: durationEnv("STORAGE_PRESIGN_TTL", 15*time.Minute),
	})
	if err != nil {
		logger.Error("storage init failed", "err", err)
		return 1
	}
	server := &http.Server{
		Addr:              envDefault("STORAGE_ADDR", ":8080"),
		Handler:           storage.NewHandler(service).Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		logger.Info("storage service listening", "addr", server.Addr)
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("storage shutdown failed", "err", err)
			return 1
		}
		return 0
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return 0
		}
		logger.Error("storage service stopped", "err", err)
		return 1
	}
}

func durationEnv(name string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	if d, err := time.ParseDuration(raw); err == nil {
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
