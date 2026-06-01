package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/manuelibar/workbench/internal/errs"
)

const (
	implName    = "workbench"
	implVersion = "v0.2.0-foundation"
)

type Options struct {
	ArtifactDir string
	SyncTimeout time.Duration
	Logger      *slog.Logger
	Planner     Planner
}

type Server struct {
	sdk       *mcp.Server
	log       *slog.Logger
	context   *ContextStore
	artifacts *ArtifactStore
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
	registry := NewContractRegistry()
	artifacts, err := NewArtifactStore(opts.ArtifactDir, registry)
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
		artifacts: artifacts,
		catalog:   NewCapabilityCatalog(),
		planner:   planner,
		sync:      NewCapabilitySync(opts.SyncTimeout),
		active: activeSurface{
			tools:             map[string]bool{},
			resources:         map[string]bool{},
			resourceTemplates: map[string]bool{},
		},
	}
	s.sdk = mcp.NewServer(
		&mcp.Implementation{Name: implName, Version: implVersion},
		&mcp.ServerOptions{
			Instructions: "Workbench is a stdio MCP context and artifact kernel. Call context to read or patch focus/artifact selection; use artifact.begin/list/get and the selected-artifact tools for typed Markdown artifacts.",
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

func (s *Server) Run(ctx context.Context, transport mcp.Transport) error {
	return s.sdk.Run(ctx, transport)
}

func (s *Server) SDKServer() *mcp.Server {
	return s.sdk
}

func (s *Server) ContextStore() *ContextStore {
	return s.context
}

func (s *Server) ArtifactStore() *ArtifactStore {
	return s.artifacts
}

func (s *Server) SetSyncTimeout(timeout time.Duration) {
	s.sync.SetTimeout(timeout)
}

func (s *Server) installCapabilityListMiddleware() {
	s.sdk.AddReceivingMiddleware(func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
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
	s.sdk.AddReceivingMiddleware(func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
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
			toolResult, ok := res.(*mcp.CallToolResult)
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

func sanitizedToolResult(pub publicMCPError) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: pub.Title}},
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
		code = mcp.CodeResourceNotFound
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

	desiredTools := toolsFromIndex(plan.Index)
	desiredResources := resourcesFromIndex(plan.Index)
	desiredTemplates := templatesFromIndex(plan.Index)
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
			s.sdk.RemoveTools(name)
			delete(s.active.tools, name)
		}
	}
	for uri := range s.active.resources {
		if !desiredResources[uri] {
			s.sdk.RemoveResources(uri)
			delete(s.active.resources, uri)
		}
	}
	for uriTemplate := range s.active.resourceTemplates {
		if !desiredTemplates[uriTemplate] {
			s.sdk.RemoveResourceTemplates(uriTemplate)
			delete(s.active.resourceTemplates, uriTemplate)
		}
	}
	for _, tool := range plan.Index.Tools {
		if s.active.tools[tool.Name] {
			continue
		}
		s.registerTool(tool.Name)
		s.active.tools[tool.Name] = true
	}
	for _, resource := range plan.Index.Resources {
		if s.active.resources[resource.URI] {
			continue
		}
		s.registerResource(resource.URI)
		s.active.resources[resource.URI] = true
	}
	for _, template := range plan.Index.ResourceTemplates {
		if s.active.resourceTemplates[template.URITemplate] {
			continue
		}
		s.registerResourceTemplate(template.URITemplate)
		s.active.resourceTemplates[template.URITemplate] = true
	}
	return changed
}

func (s *Server) registerTool(name string) {
	switch name {
	case "context":
		mcp.AddTool(s.sdk, &mcp.Tool{Name: name, Description: "Read or patch focus/artifact context and return raw context plus capability sync status."}, s.handleContext)
	case "artifact.begin":
		mcp.AddTool(s.sdk, &mcp.Tool{Name: name, Description: "Create a typed Markdown artifact draft under docs/artifacts."}, s.handleArtifactBegin)
	case "artifact.list":
		mcp.AddTool(s.sdk, &mcp.Tool{Name: name, Description: "List file-backed artifacts in docs/artifacts."}, s.handleArtifactList)
	case "artifact.get":
		mcp.AddTool(s.sdk, &mcp.Tool{Name: name, Description: "Read one artifact by stable id."}, s.handleArtifactGet)
	case "artifact.update":
		mcp.AddTool(s.sdk, &mcp.Tool{Name: name, Description: "Update selected artifact metadata or section bodies."}, s.handleArtifactUpdate)
	case "artifact.guidance":
		mcp.AddTool(s.sdk, &mcp.Tool{Name: name, Description: "Return deterministic contract guidance for the selected artifact."}, s.handleArtifactGuidance)
	case "artifact.validate":
		mcp.AddTool(s.sdk, &mcp.Tool{Name: name, Description: "Validate the selected artifact against its type contract."}, s.handleArtifactValidate)
	default:
		panic(fmt.Sprintf("unknown tool %q", name))
	}
}

func (s *Server) registerResource(uri string) {
	switch {
	case uri == "workbench:///context":
		s.sdk.AddResource(&mcp.Resource{
			URI:         uri,
			Name:        "context",
			Title:       "Workbench Context",
			Description: "Current raw Workbench context document.",
			MIMEType:    "text/markdown",
		}, s.readContextResource)
	case artifactIDFromURI(uri) != "":
		s.sdk.AddResource(&mcp.Resource{
			URI:         uri,
			Name:        "selected_artifact",
			Title:       "Selected Artifact",
			Description: "Selected artifact Markdown.",
			MIMEType:    "text/markdown",
		}, s.readArtifactResource)
	default:
		panic(fmt.Sprintf("unknown resource %q", uri))
	}
}

func (s *Server) registerResourceTemplate(uriTemplate string) {
	switch uriTemplate {
	case "workbench:///artifacts/{id}":
		s.sdk.AddResourceTemplate(&mcp.ResourceTemplate{
			URITemplate: uriTemplate,
			Name:        "artifact",
			Title:       "Artifact",
			Description: "Read an artifact Markdown file by stable id.",
			MIMEType:    "text/markdown",
		}, s.readArtifactResource)
	default:
		panic(fmt.Sprintf("unknown resource template %q", uriTemplate))
	}
}

func toolsFromIndex(index CapabilityIndex) map[string]bool {
	out := map[string]bool{}
	for _, tool := range index.Tools {
		out[tool.Name] = true
	}
	return out
}

func resourcesFromIndex(index CapabilityIndex) map[string]bool {
	out := map[string]bool{}
	for _, resource := range index.Resources {
		out[resource.URI] = true
	}
	return out
}

func templatesFromIndex(index CapabilityIndex) map[string]bool {
	out := map[string]bool{}
	for _, template := range index.ResourceTemplates {
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
