package mcp

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/manuelibar/workbench/internal/artifacts"
	"github.com/manuelibar/workbench/internal/errs"
	"github.com/manuelibar/workbench/internal/mcp/middleware"
	"github.com/manuelibar/workbench/internal/storageclient"
)

const (
	implName    = "workbench"
	implVersion = "v0.2.0-foundation"
)

type Options struct {
	ArtifactDir         string
	StorageClient       *storageclient.Client
	StorageOrgID        string
	StorageProjectID    string
	StorageResourceType string
	SyncTimeout         time.Duration
	Logger              *slog.Logger
}

type Server struct {
	sdk       *mcpsdk.Server
	log       *slog.Logger
	scope     *scopeStore
	scopeDir  string
	artifacts *artifacts.Store
	catalog   capabilityCatalog
	planner   planner
	sync      *capabilitySync
	surface   *surfaceSynchronizer
}

func New(opts Options) (*Server, error) {
	log := opts.Logger
	if log == nil {
		log = slog.Default()
	}
	registry := artifacts.NewRegistry()
	var artifactStore *artifacts.Store
	var err error
	if opts.StorageClient != nil {
		artifactStore, err = artifacts.NewStorageStore(artifacts.StorageBackendOptions{
			Client:       opts.StorageClient,
			OrgID:        opts.StorageOrgID,
			ProjectID:    opts.StorageProjectID,
			ResourceType: opts.StorageResourceType,
		}, registry)
	} else {
		artifactStore, err = artifacts.NewStore(opts.ArtifactDir, registry)
	}
	if err != nil {
		return nil, err
	}
	scopeDir, err := os.MkdirTemp("", "workbench-scope-*")
	if err != nil {
		return nil, err
	}
	s := &Server{
		log:       log,
		scope:     newScopeStore(),
		scopeDir:  scopeDir,
		artifacts: artifactStore,
		catalog:   newCapabilityCatalog(),
		planner:   deterministicPlanner{},
		sync:      newCapabilitySync(opts.SyncTimeout),
		surface:   newSurfaceSynchronizer(),
	}
	s.sdk = mcpsdk.NewServer(
		&mcpsdk.Implementation{Name: implName, Version: implVersion},
		&mcpsdk.ServerOptions{
			Instructions: "Workbench is a stdio MCP scope and artifact kernel. Use artifact.find or artifact.create to locate drafts, contextualize with artifact_id to place one artifact in scope, read the artifact resource, and call artifact.upload to persist full Markdown changes.",
			Logger:       log,
		},
	)
	middleware.ObserveCapabilityLists(s.sdk, s.sync.MarkObserved)
	middleware.SanitizeClassifiedErrors(s.sdk, publicError, s.logClassifiedMCPError)
	plan, err := s.plan(context.Background(), s.scope.Snapshot())
	if err != nil {
		_ = os.RemoveAll(scopeDir)
		return nil, err
	}
	s.surface.Synchronize(s, plan.Active)
	return s, nil
}

func (s *Server) Run(ctx context.Context, transport mcpsdk.Transport) error {
	defer os.RemoveAll(s.scopeDir)
	session, err := s.sdk.Connect(ctx, transport, nil)
	if err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return err
	}
	done := make(chan error, 1)
	go func() {
		done <- session.Wait()
	}()
	select {
	case <-ctx.Done():
		_ = session.Close()
		<-done
		return nil
	case err := <-done:
		if ctx.Err() != nil {
			return nil
		}
		return err
	}
}

func publicError(err error) (middleware.PublicError, bool) {
	sentinel := errs.SentinelOf(err)
	code := errs.CodeOf(err)
	if sentinel == nil && code == "" {
		return middleware.PublicError{}, false
	}
	if code == "" {
		code = defaultPublicCode(sentinel)
	}
	title := publicTitle(err)
	if title == "" {
		title = defaultPublicTitle(sentinel, code)
	}
	return middleware.PublicError{
		Title:     title,
		Code:      code,
		Retryable: errs.IsRetryable(err),
		Sentinel:  sentinel,
	}, true
}

func publicTitle(err error) string {
	var classified *errs.Error
	if errors.As(err, &classified) && classified.Message() != "" {
		return classified.Message()
	}
	var multi *errs.Multi
	if errors.As(err, &multi) && multi.Message() != "" {
		return multi.Message()
	}
	return ""
}

func (s *Server) logClassifiedMCPError(method string, err error, pub middleware.PublicError) {
	s.log.Debug(
		"classified MCP error",
		"method", method,
		"title", pub.Title,
		"code", pub.Code.String(),
		"retryable", pub.Retryable,
		"attrs", errs.AttrsOf(err),
		"err", err,
	)
}

func (s *Server) plan(ctx context.Context, state scopeState) (capabilityPlan, error) {
	attrs := map[string]any{"operation": "capability.plan"}
	plan, err := s.planner.Plan(ctx, state, s.catalog)
	if err == nil {
		if validateErr := s.catalog.ValidatePlan(state, plan); validateErr == nil {
			return plan, nil
		} else {
			s.log.Warn("planner returned invalid capability plan; using deterministic fallback", "err", validateErr)
		}
	} else {
		s.log.Warn("planner failed; using deterministic fallback", "err", err)
	}
	plan, err = deterministicPlanner{}.Plan(ctx, state, s.catalog)
	if err != nil {
		return capabilityPlan{}, errs.New(
			"planner unavailable",
			errs.WithSentinel(errs.ErrUnavailable),
			errs.WithCode(errCodePlannerUnavailable),
			errs.WithSeverity(errs.SeverityError),
			errs.WithCause(err),
			errs.WithAttrs(attrs),
			errs.WithRetryable(true),
		)
	}
	return plan, nil
}
