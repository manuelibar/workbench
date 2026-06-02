package mcp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/manuelibar/workbench/internal/artifacts"
	"github.com/manuelibar/workbench/internal/errs"
	mcpresources "github.com/manuelibar/workbench/internal/mcp/resources"
	"github.com/manuelibar/workbench/internal/mcp/tools"
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
	Planner             Planner
}

type Server struct {
	sdk       *mcpsdk.Server
	log       *slog.Logger
	context   *ContextStore
	artifacts *artifacts.Store
	catalog   CapabilityCatalog
	planner   Planner
	sync      *CapabilitySync

	surfaceMu sync.Mutex
	active    activeSurface
}

type activeSurface struct {
	tools             map[string]bool
	resources         map[string]bool
	resourceTemplates map[string]bool
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
	planner := opts.Planner
	if planner == nil {
		planner = deterministicPlanner{}
	}
	s := &Server{
		log:       log,
		context:   NewContextStore(),
		artifacts: artifactStore,
		catalog:   NewCapabilityCatalog(),
		planner:   planner,
		sync:      NewCapabilitySync(opts.SyncTimeout),
		active: activeSurface{
			tools:             map[string]bool{},
			resources:         map[string]bool{},
			resourceTemplates: map[string]bool{},
		},
	}
	s.sdk = mcpsdk.NewServer(
		&mcpsdk.Implementation{Name: implName, Version: implVersion},
		&mcpsdk.ServerOptions{
			Instructions: "Workbench is a stdio MCP context and artifact kernel. Call contextualize to read or patch focus/artifact selection; use artifact.begin/list/get and the selected-artifact tools for typed Markdown artifacts.",
			Logger:       log,
		},
	)
	s.installCapabilityListMiddleware()
	s.installMCPErrorBoundaryMiddleware()
	plan, err := s.plan(context.Background(), s.context.Snapshot())
	if err != nil {
		return nil, err
	}
	s.applyPlan(plan)
	return s, nil
}

func (s *Server) Run(ctx context.Context, transport mcpsdk.Transport) error {
	return s.sdk.Run(ctx, transport)
}

func (s *Server) SDKServer() *mcpsdk.Server {
	return s.sdk
}

func (s *Server) ContextStore() *ContextStore {
	return s.context
}

func (s *Server) ArtifactStore() *artifacts.Store {
	return s.artifacts
}

func (s *Server) SetSyncTimeout(timeout time.Duration) {
	s.sync.SetTimeout(timeout)
}

func (s *Server) installCapabilityListMiddleware() {
	s.sdk.AddReceivingMiddleware(func(next mcpsdk.MethodHandler) mcpsdk.MethodHandler {
		return func(ctx context.Context, method string, req mcpsdk.Request) (mcpsdk.Result, error) {
			res, err := next(ctx, method, req)
			if err == nil {
				s.sync.MarkObserved(method)
			}
			return res, err
		}
	})
}

type publicMCPError struct {
	Title     string
	Code      errs.Code
	Retryable bool
	Sentinel  error
}

func (s *Server) installMCPErrorBoundaryMiddleware() {
	s.sdk.AddReceivingMiddleware(func(next mcpsdk.MethodHandler) mcpsdk.MethodHandler {
		return func(ctx context.Context, method string, req mcpsdk.Request) (mcpsdk.Result, error) {
			res, err := next(ctx, method, req)
			if err != nil {
				pub, ok := publicError(err)
				if !ok {
					return res, err
				}
				s.logClassifiedMCPError(method, err, pub)
				if method == "resources/read" {
					return nil, jsonrpcErrorFor(pub)
				}
				return nil, jsonrpcErrorFor(pub)
			}
			if method != "tools/call" {
				return res, nil
			}
			toolResult, ok := res.(*mcpsdk.CallToolResult)
			if !ok || !toolResult.IsError {
				return res, nil
			}
			pub, ok := publicError(toolResult.GetError())
			if !ok {
				return res, nil
			}
			s.logClassifiedMCPError(method, toolResult.GetError(), pub)
			return sanitizedToolResult(pub), nil
		}
	})
}

func publicError(err error) (publicMCPError, bool) {
	sentinel := errs.SentinelOf(err)
	code := errs.CodeOf(err)
	if sentinel == nil && code == "" {
		return publicMCPError{}, false
	}
	if code == "" {
		code = defaultPublicCode(sentinel)
	}
	title := publicTitle(err)
	if title == "" {
		title = defaultPublicTitle(sentinel, code)
	}
	return publicMCPError{
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

func sanitizedToolResult(pub publicMCPError) *mcpsdk.CallToolResult {
	return &mcpsdk.CallToolResult{
		Content: []mcpsdk.Content{&mcpsdk.TextContent{Text: pub.Title}},
		StructuredContent: map[string]any{
			"error": map[string]any{
				"title":     pub.Title,
				"code":      pub.Code.String(),
				"retryable": pub.Retryable,
			},
		},
		IsError: true,
	}
}

func jsonrpcErrorFor(pub publicMCPError) error {
	code := int64(jsonrpc.CodeInternalError)
	switch {
	case errors.Is(pub.Sentinel, errs.ErrNotFound):
		code = mcpsdk.CodeResourceNotFound
	case errors.Is(pub.Sentinel, errs.ErrInvalid):
		code = jsonrpc.CodeInvalidParams
	}
	return &jsonrpc.Error{
		Code:    code,
		Message: pub.Title,
	}
}

func (s *Server) logClassifiedMCPError(method string, err error, pub publicMCPError) {
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

func (s *Server) plan(ctx context.Context, state ContextState) (CapabilityPlan, error) {
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
		return CapabilityPlan{}, errs.New(
			"Planner unavailable",
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

func (s *Server) applyPlan(plan CapabilityPlan) []string {
	s.surfaceMu.Lock()
	defer s.surfaceMu.Unlock()

	desiredTools := toolsFromSurface(plan.Active)
	desiredResources := resourcesFromSurface(plan.Active)
	desiredTemplates := templatesFromSurface(plan.Active)
	var changed []string
	if !sameStringSet(s.active.tools, desiredTools) {
		changed = append(changed, "tools")
	}
	if !sameStringSet(s.active.resources, desiredResources) {
		changed = append(changed, "resources")
	}
	if !sameStringSet(s.active.resourceTemplates, desiredTemplates) {
		changed = append(changed, "resource_templates")
	}

	for name := range s.active.tools {
		if !desiredTools[name] {
			s.removeTool(name)
			delete(s.active.tools, name)
		}
	}
	for uri := range s.active.resources {
		if !desiredResources[uri] {
			s.removeResource(uri)
			delete(s.active.resources, uri)
		}
	}
	for uriTemplate := range s.active.resourceTemplates {
		if !desiredTemplates[uriTemplate] {
			s.removeResourceTemplate(uriTemplate)
			delete(s.active.resourceTemplates, uriTemplate)
		}
	}
	for _, tool := range plan.Active.Tools {
		if s.active.tools[tool.Name] {
			continue
		}
		s.addTool(tool.Name)
		s.active.tools[tool.Name] = true
	}
	for _, resource := range plan.Active.Resources {
		if s.active.resources[resource.URI] {
			continue
		}
		s.addResource(resource.URI)
		s.active.resources[resource.URI] = true
	}
	for _, template := range plan.Active.ResourceTemplates {
		if s.active.resourceTemplates[template.URITemplate] {
			continue
		}
		s.addResourceTemplate(template.URITemplate)
		s.active.resourceTemplates[template.URITemplate] = true
	}
	return changed
}

func (s *Server) addTool(name string) {
	tool, ok := tools.ByName(name)
	if !ok {
		panic(fmt.Sprintf("unknown tool %q", name))
	}
	tool.AddTo(s.sdk, s)
}

func (s *Server) removeTool(name string) {
	s.sdk.RemoveTools(name)
}

func (s *Server) addResource(uri string) {
	def, ok := mcpresources.ByURI(uri)
	if !ok {
		panic(fmt.Sprintf("unknown resource %q", uri))
	}
	switch {
	case uri == mcpresources.ContextURI:
		s.sdk.AddResource(&mcpsdk.Resource{
			URI:         uri,
			Name:        def.Name(),
			Title:       def.Title(),
			Description: def.Description(),
			MIMEType:    def.MIMEType(),
		}, s.readContextResource)
	case artifactIDFromURI(uri) != "":
		id := artifactIDFromURI(uri)
		if artifact, err := s.artifacts.GetContext(context.Background(), id); err == nil {
			def = mcpresources.NewSelectedArtifactResource(selectedArtifactResource(artifact.Summary))
		}
		s.addResourceDefinition(def, s.readArtifactResource)
	default:
		panic(fmt.Sprintf("unknown resource %q", uri))
	}
}

func (s *Server) removeResource(uri string) {
	s.sdk.RemoveResources(uri)
}

func (s *Server) refreshSelectedArtifactResource(artifact artifacts.Summary) {
	uri := mcpresources.ArtifactURI(artifact.ID)
	s.surfaceMu.Lock()
	active := s.active.resources[uri]
	s.surfaceMu.Unlock()
	if !active {
		return
	}
	s.addSelectedArtifactResource(selectedArtifactResource(artifact))
}

func (s *Server) addSelectedArtifactResource(selected mcpresources.SelectedArtifact) {
	s.addResourceDefinition(mcpresources.NewSelectedArtifactResource(selected), s.readArtifactResource)
}

func (s *Server) addResourceDefinition(def mcpresources.Definition, handler mcpsdk.ResourceHandler) {
	s.sdk.AddResource(&mcpsdk.Resource{
		URI:         def.URI(),
		Name:        def.Name(),
		Title:       def.Title(),
		Description: def.Description(),
		MIMEType:    def.MIMEType(),
	}, handler)
}

func (s *Server) addResourceTemplate(uriTemplate string) {
	def, ok := mcpresources.TemplateByURITemplate(uriTemplate)
	if !ok {
		panic(fmt.Sprintf("unknown resource template %q", uriTemplate))
	}
	switch uriTemplate {
	case mcpresources.ArtifactTemplateURI:
		s.sdk.AddResourceTemplate(&mcpsdk.ResourceTemplate{
			URITemplate: uriTemplate,
			Name:        def.Name(),
			Title:       def.Title(),
			Description: def.Description(),
			MIMEType:    def.MIMEType(),
		}, s.readArtifactResource)
	default:
		panic(fmt.Sprintf("unknown resource template %q", uriTemplate))
	}
}

func (s *Server) removeResourceTemplate(uriTemplate string) {
	s.sdk.RemoveResourceTemplates(uriTemplate)
}

func toolsFromSurface(surface tools.CapabilitySurface) map[string]bool {
	out := map[string]bool{}
	for _, tool := range surface.Tools {
		out[tool.Name] = true
	}
	return out
}

func resourcesFromSurface(surface tools.CapabilitySurface) map[string]bool {
	out := map[string]bool{}
	for _, resource := range surface.Resources {
		out[resource.URI] = true
	}
	return out
}

func templatesFromSurface(surface tools.CapabilitySurface) map[string]bool {
	out := map[string]bool{}
	for _, template := range surface.ResourceTemplates {
		out[template.URITemplate] = true
	}
	return out
}

func sameStringSet(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for key := range a {
		if !b[key] {
			return false
		}
	}
	return true
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for key := range m {
		out = append(out, key)
	}
	slices.Sort(out)
	return out
}
