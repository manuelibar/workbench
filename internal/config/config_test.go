package config_test

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/manuelibar/workbench/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("WORKBENCH_BIND", "")
	t.Setenv("WORKBENCH_DB_URL", "")
	t.Setenv("WORKBENCH_LOG_LEVEL", "")

	c, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.Bind != config.DefaultBind {
		t.Errorf("Bind = %q, want %q", c.Bind, config.DefaultBind)
	}
	if c.DBURL != config.DefaultDBURL {
		t.Errorf("DBURL = %q, want %q", c.DBURL, config.DefaultDBURL)
	}
	if c.LogLevel != config.DefaultLogLevel {
		t.Errorf("LogLevel = %q, want %q", c.LogLevel, config.DefaultLogLevel)
	}
}

func TestLoad_Override(t *testing.T) {
	t.Setenv("WORKBENCH_BIND", "127.0.0.1:9999")
	t.Setenv("WORKBENCH_DB_URL", "postgres://x@y/z")
	t.Setenv("WORKBENCH_LOG_LEVEL", "DEBUG") // case-insensitive

	c, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.Bind != "127.0.0.1:9999" {
		t.Errorf("Bind = %q", c.Bind)
	}
	if c.DBURL != "postgres://x@y/z" {
		t.Errorf("DBURL = %q", c.DBURL)
	}
	if c.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want debug", c.LogLevel)
	}
}

func TestLoad_InvalidLogLevel(t *testing.T) {
	t.Setenv("WORKBENCH_LOG_LEVEL", "verbose")

	_, err := config.Load()
	if !errors.Is(err, config.ErrInvalidLogLevel) {
		t.Fatalf("err = %v, want ErrInvalidLogLevel", err)
	}
}

func TestSlogLevel(t *testing.T) {
	t.Parallel()
	cases := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
		"":      slog.LevelInfo,
	}
	for level, want := range cases {
		level, want := level, want
		t.Run(level, func(t *testing.T) {
			t.Parallel()
			c := config.Config{LogLevel: level}
			if got := c.SlogLevel(); got != want {
				t.Errorf("SlogLevel(%q) = %v, want %v", level, got, want)
			}
		})
	}
}
