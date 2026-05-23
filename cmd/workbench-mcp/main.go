package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/manuelibar/workbench/internal/mcpserver"
	"github.com/manuelibar/workbench/internal/mcpserver/skills"
	pgstorage "github.com/manuelibar/workbench/internal/mcpserver/storage/postgres"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	databaseURL := os.Getenv("WORKBENCH_DATABASE_URL")
	storeBackend := os.Getenv("WORKBENCH_STORE_BACKEND")
	if databaseURL != "" || storeBackend == "postgres" {
		if databaseURL == "" {
			log.Error("postgres backend requested but WORKBENCH_DATABASE_URL is empty")
			os.Exit(1)
		}
		pgDB, err := pgstorage.Open(context.Background(), databaseURL)
		if err != nil {
			log.Error("postgres init failed", "err", err)
			os.Exit(1)
		}
		defer pgDB.Close()
		if err := pgstorage.Migrate(context.Background(), pgDB.SQL); err != nil {
			log.Error("postgres migration failed", "err", err)
			os.Exit(1)
		}
		log.Info("postgres migrations applied", "backend", storeBackend)
	}

	storePath := os.Getenv("WORKBENCH_STORE")
	if storePath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Error("home dir unavailable", "err", err)
			os.Exit(1)
		}
		storePath = filepath.Join(home, ".workbench", "projects.json")
	}

	store, err := mcpserver.NewFileProjectStore(storePath)
	if err != nil {
		log.Error("store init failed", "path", storePath, "err", err)
		os.Exit(1)
	}

	registry := skills.SkillRegistry(skills.NewEmbeddedRegistry())
	if skillsDir := os.Getenv("WORKBENCH_SKILLS_DIR"); skillsDir != "" {
		fsRegistry, err := skills.NewFilesystemRegistry(skillsDir)
		if err != nil {
			log.Error("skills registry init failed", "path", skillsDir, "err", err)
			os.Exit(1)
		}
		registry = skills.NewOverlayRegistry(fsRegistry, registry)
	}
	srv := mcpserver.New(log, store, registry)
	srv.SetProjectRoot(firstNonEmpty(os.Getenv("WORKBENCH_PROJECT_ROOT"), mustGetwd()))
	srv.SetGitHubConfig(mcpserver.GitHubConfig{
		Organization: os.Getenv("WORKBENCH_GITHUB_ORG"),
		Token:        firstNonEmpty(os.Getenv("WORKBENCH_GITHUB_TOKEN"), os.Getenv("GITHUB_TOKEN")),
	})
	if kbURL := os.Getenv("WORKBENCH_KB_URL"); kbURL != "" {
		srv.SetKBRetriever(mcpserver.NewHTTPKBRetriever(kbURL))
		srv.SetAskSynthesizer(mcpserver.NewCodexAskSynthesizer(firstNonEmpty(os.Getenv("WORKBENCH_CODEX_COMMAND"), "codex")))
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	defer stop()

	if err := srv.Run(ctx); err != nil {
		log.Error("server error", "err", err)
		os.Exit(1)
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}
