package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// Defaults applied by [Load] when an environment variable is unset or empty.
const (
	DefaultBind       = "127.0.0.1:7777"
	DefaultDBURL      = "postgres://workbench:workbench@127.0.0.1:5432/workbench?sslmode=disable"
	DefaultLogLevel   = "info"
	DefaultBacklogURL = "http://127.0.0.1:7778"
)

// ErrInvalidLogLevel is returned by [Load] when WORKBENCH_LOG_LEVEL is set to
// a value other than debug, info, warn, or error.
var ErrInvalidLogLevel = errors.New("config: invalid WORKBENCH_LOG_LEVEL")

// Config is the typed view of the workbench-mcp environment.
type Config struct {
	// Bind is the TCP address the HTTP server listens on, including port.
	Bind string
	// DBURL is the Postgres connection string passed to pgx.
	DBURL string
	// LogLevel is the slog level: "debug", "info", "warn", or "error".
	LogLevel string
	// BacklogURL is the HTTP base URL of the separate backlog-service. The
	// MCP `backlog.*` tools proxy to it. Empty disables the surface (handlers
	// return a clear error).
	BacklogURL string
}

// SlogLevel returns the [slog.Level] corresponding to [Config.LogLevel].
// Validation happens during [Load]; SlogLevel does not error and falls back to
// info for any unrecognised value.
func (c Config) SlogLevel() slog.Level {
	switch c.LogLevel {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Load reads the workbench-mcp configuration from the process environment,
// returning a [Config] populated with defaults for any unset or empty fields.
//
// Load returns [ErrInvalidLogLevel] (wrapped) if WORKBENCH_LOG_LEVEL is set
// to anything other than debug, info, warn, or error.
func Load() (Config, error) {
	c := Config{
		Bind:       env("WORKBENCH_BIND", DefaultBind),
		DBURL:      env("WORKBENCH_DB_URL", DefaultDBURL),
		LogLevel:   strings.ToLower(env("WORKBENCH_LOG_LEVEL", DefaultLogLevel)),
		BacklogURL: env("WORKBENCH_BACKLOG_URL", DefaultBacklogURL),
	}
	switch c.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return Config{}, fmt.Errorf("%w: %q", ErrInvalidLogLevel, c.LogLevel)
	}
	return c, nil
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
