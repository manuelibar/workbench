// Command workbench-mcp is the workbench MCP server entrypoint.
//
// At boot, the binary:
//
//  1. Loads configuration from the process environment via [config.Load].
//  2. Connects to Postgres ([pgstore.Open]) and applies any pending
//     migrations.
//  3. Ensures a singleton user and one open WorkSession exist.
//  4. Mounts the streamable-HTTP MCP transport at /mcp ([mcpserver.New])
//     and serves /healthz (always 200) plus /readyz (200 only when the DB
//     is reachable).
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/manuelibar/workbench/internal/config"
	"github.com/manuelibar/workbench/internal/domain"
	"github.com/manuelibar/workbench/internal/mcpserver"
	"github.com/manuelibar/workbench/internal/pgstore"
)

const (
	dbPingTimeout = 30 * time.Second
	dbPingEvery   = time.Second
)

func main() {
	os.Exit(run())
}

func run() int {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load failed", "err", err)
		return 1
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: cfg.SlogLevel(),
	}))
	slog.SetDefault(logger)

	rootCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	store, user, ws, err := bootstrap(rootCtx, cfg)
	if err != nil {
		slog.Error("bootstrap failed", "err", err)
		return 1
	}
	defer store.Close()
	slog.Info("workbench bootstrapped",
		"user_id", user.ID,
		"work_session_id", ws.ID,
		"work_session_name", ws.Name,
	)

	mcpsrv := mcpserver.New(store, user, ws, logger)
	srv := newHTTPServer(cfg.Bind, store, mcpsrv)

	listenErr := make(chan error, 1)
	go func() {
		slog.Info("workbench-mcp listening", "addr", cfg.Bind, "log_level", cfg.LogLevel)
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			listenErr <- err
		}
		close(listenErr)
	}()

	select {
	case <-rootCtx.Done():
		slog.Info("shutdown signal received")
	case err, ok := <-listenErr:
		if ok && err != nil {
			slog.Error("listen failed", "err", err)
			return 1
		}
		return 0
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown failed", "err", err)
		return 1
	}
	slog.Info("workbench-mcp stopped")
	return 0
}

// bootstrap connects to Postgres (with retry), runs migrations, and ensures
// the singleton user + open WorkSession.
func bootstrap(ctx context.Context, cfg config.Config) (*pgstore.Store, domain.User, domain.WorkSession, error) {
	store, err := openWithRetry(ctx, cfg.DBURL)
	if err != nil {
		return nil, domain.User{}, domain.WorkSession{}, err
	}
	if err := store.Migrate(ctx); err != nil {
		store.Close()
		return nil, domain.User{}, domain.WorkSession{}, err
	}
	user, err := store.EnsureSingletonUser(ctx, "")
	if err != nil {
		store.Close()
		return nil, domain.User{}, domain.WorkSession{}, err
	}
	ws, err := store.EnsureOpenWorkSession(ctx, user.ID, "")
	if err != nil {
		store.Close()
		return nil, domain.User{}, domain.WorkSession{}, err
	}
	return store, user, ws, nil
}

// openWithRetry calls [pgstore.Open] up to dbPingTimeout, sleeping dbPingEvery
// between attempts. Returns the first successful Store, or the last error
// after the deadline.
func openWithRetry(ctx context.Context, dsn string) (*pgstore.Store, error) {
	deadline := time.Now().Add(dbPingTimeout)
	var lastErr error
	for {
		attemptCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		store, err := pgstore.Open(attemptCtx, dsn)
		cancel()
		if err == nil {
			return store, nil
		}
		lastErr = err
		if time.Now().After(deadline) {
			return nil, lastErr
		}
		slog.Warn("postgres unreachable, retrying", "err", err)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(dbPingEvery):
		}
	}
}

// newHTTPServer wires the HTTP routes for /healthz, /readyz, and /mcp.
func newHTTPServer(addr string, store *pgstore.Store, mcpsrv *mcpserver.Server) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthz)
	mux.Handle("/readyz", readyz(store))
	mux.Handle("/mcp", mcpsrv.Handler())

	return &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}

// readyz returns an http.Handler that pings the DB on each request. Results
// are cached for 1 second to avoid hammering Postgres on probe storms.
func readyz(store *pgstore.Store) http.Handler {
	type cached struct {
		ok        bool
		expiresAt time.Time
		err       string
	}
	var atomicState atomic.Pointer[cached]
	const ttl = time.Second

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		c := atomicState.Load()
		if c == nil || time.Now().After(c.expiresAt) {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			err := store.Ping(ctx)
			next := &cached{ok: err == nil, expiresAt: time.Now().Add(ttl)}
			if err != nil {
				next.err = err.Error()
			}
			atomicState.Store(next)
			c = next
		}
		if !c.ok {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("not ready: " + c.err + "\n"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready\n"))
	})
}
