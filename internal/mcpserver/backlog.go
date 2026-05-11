package mcpserver

import (
	"context"
	"errors"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/manuelibar/workbench/internal/backlogclient"
)

// errBacklogNotConfigured is returned by every backlog handler when the
// server was constructed without a [backlogclient.Client]. The MCP
// transports surface this as a tool error so the agent can self-correct
// (typically by setting `WORKBENCH_BACKLOG_URL`).
var errBacklogNotConfigured = errors.New("backlog service not configured: set WORKBENCH_BACKLOG_URL and start backlog-service")

// resolveBacklogProjectID returns raw if set, otherwise the currently-
// selected project's UUID (as string), otherwise an error. Used by
// `backlog.add` where a project_id is mandatory.
func (s *Server) resolveBacklogProjectID(raw string) (string, error) {
	if raw != "" {
		return raw, nil
	}
	sel := s.currentSelection()
	if sel.ProjectID == nil {
		return "", errors.New("project_id required (no project selected; pass project_id or call refresh(project_id=...))")
	}
	return sel.ProjectID.String(), nil
}

// backlog.add ------------------------------------------------------------

// BacklogAddArgs are the args of `backlog.add`. ProjectID falls back to the
// currently-selected project; the remaining fields are forwarded as-is.
type BacklogAddArgs struct {
	ProjectID  string   `json:"project_id,omitempty" jsonschema:"target project (opaque string; defaults to currently-selected project's UUID)"`
	Title      string   `json:"title"                jsonschema:"issue title"`
	Body       string   `json:"body,omitempty"       jsonschema:"longer markdown description"`
	Type       string   `json:"type,omitempty"       jsonschema:"issue type; defaults to 'task'. Free-form (e.g. 'bug', 'story', 'epic')"`
	Priority   string   `json:"priority,omitempty"   jsonschema:"urgent|high|med|low; defaults to med"`
	Labels     []string `json:"labels,omitempty"`
	ParentID   string   `json:"parent_id,omitempty"  jsonschema:"UUID of the parent issue (e.g. an epic this story belongs to)"`
	Assignee   string   `json:"assignee,omitempty"   jsonschema:"explicit initial assignee; otherwise unassigned until take_next claims it"`
	SourceRefs []string `json:"source_refs,omitempty" jsonschema:"opaque URIs back to the source material — typically workbench:///notes/{id} when the issue is raised from a captured note. The source is never modified."`
}

// BacklogAddResult is the return shape of `backlog.add`.
type BacklogAddResult struct {
	Issue backlogclient.Issue `json:"issue"`
}

func (s *Server) handleBacklogAdd(ctx context.Context, _ *mcp.CallToolRequest, args BacklogAddArgs) (*mcp.CallToolResult, BacklogAddResult, error) {
	if s.backlog == nil {
		return nil, BacklogAddResult{}, errBacklogNotConfigured
	}
	if args.Title == "" {
		return nil, BacklogAddResult{}, errors.New("backlog.add: title required")
	}
	projectID, err := s.resolveBacklogProjectID(args.ProjectID)
	if err != nil {
		return nil, BacklogAddResult{}, err
	}
	issue, err := s.backlog.Create(ctx, backlogclient.IssueCreateRequest{
		ProjectID:  projectID,
		Title:      args.Title,
		Body:       args.Body,
		Type:       args.Type,
		Priority:   args.Priority,
		Labels:     args.Labels,
		ParentID:   args.ParentID,
		Assignee:   args.Assignee,
		SourceRefs: args.SourceRefs,
	})
	if err != nil {
		return nil, BacklogAddResult{}, err
	}
	return nil, BacklogAddResult{Issue: *issue}, nil
}

// backlog.list -----------------------------------------------------------

// BacklogListArgs are the args of `backlog.list`. Every field is optional;
// no auto-fill from selection — a missing ProjectID means "master backlog
// across all projects".
type BacklogListArgs struct {
	ProjectID string `json:"project_id,omitempty"`
	Status    string `json:"status,omitempty"     jsonschema:"todo|in_progress|blocked|done|archived"`
	Type      string `json:"type,omitempty"`
	Priority  string `json:"priority,omitempty"`
	Label     string `json:"label,omitempty"      jsonschema:"single label to filter by; case-insensitive"`
	ParentID  string `json:"parent_id,omitempty"`
	Assignee  string `json:"assignee,omitempty"`
	SourceRef string `json:"source_ref,omitempty" jsonschema:"exact-match URI to find every issue that traces back to this source (e.g. workbench:///notes/{id})"`
	Q         string `json:"q,omitempty"          jsonschema:"substring search across title + body"`
	Order     string `json:"order,omitempty"      jsonschema:"created_at|updated_at|priority"`
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
}

// BacklogListResult is the return shape of `backlog.list`.
type BacklogListResult struct {
	Items      []backlogclient.Issue `json:"items"`
	Count      int                   `json:"count"`
	NextOffset int                   `json:"next_offset,omitempty"`
}

func (s *Server) handleBacklogList(ctx context.Context, _ *mcp.CallToolRequest, args BacklogListArgs) (*mcp.CallToolResult, BacklogListResult, error) {
	if s.backlog == nil {
		return nil, BacklogListResult{}, errBacklogNotConfigured
	}
	resp, err := s.backlog.List(ctx, backlogclient.IssueListQuery{
		ProjectID: args.ProjectID,
		Status:    args.Status,
		Type:      args.Type,
		Priority:  args.Priority,
		Label:     args.Label,
		ParentID:  args.ParentID,
		Assignee:  args.Assignee,
		SourceRef: args.SourceRef,
		Q:         args.Q,
		Order:     args.Order,
		Limit:     args.Limit,
		Offset:    args.Offset,
	})
	if err != nil {
		return nil, BacklogListResult{}, err
	}
	items := resp.Items
	if items == nil {
		items = []backlogclient.Issue{}
	}
	return nil, BacklogListResult{
		Items:      items,
		Count:      len(items),
		NextOffset: resp.NextOffset,
	}, nil
}

// backlog.get ------------------------------------------------------------

// BacklogGetArgs are the args of `backlog.get`.
type BacklogGetArgs struct {
	ID string `json:"id" jsonschema:"issue UUID"`
}

// BacklogGetResult is the return shape of `backlog.get`.
type BacklogGetResult struct {
	Issue backlogclient.Issue `json:"issue"`
}

func (s *Server) handleBacklogGet(ctx context.Context, _ *mcp.CallToolRequest, args BacklogGetArgs) (*mcp.CallToolResult, BacklogGetResult, error) {
	if s.backlog == nil {
		return nil, BacklogGetResult{}, errBacklogNotConfigured
	}
	if args.ID == "" {
		return nil, BacklogGetResult{}, errors.New("backlog.get: id required")
	}
	issue, err := s.backlog.Get(ctx, args.ID)
	if err != nil {
		return nil, BacklogGetResult{}, err
	}
	return nil, BacklogGetResult{Issue: *issue}, nil
}

// backlog.update ---------------------------------------------------------

// BacklogUpdateArgs are the args of `backlog.update`. ExpectedVersion is
// mandatory: the caller must read first (or remember the version returned
// by add/take_next) so concurrent edits collide loudly via 409.
//
// Pointer-ish semantics: zero values for string fields mean "leave
// unchanged"; passing a non-empty string sets it. Empty-string clear
// semantics are explicit via the dedicated flags below.
type BacklogUpdateArgs struct {
	ID              string `json:"id"`
	ExpectedVersion int    `json:"expected_version"`

	Title    string `json:"title,omitempty"`
	Body     string `json:"body,omitempty"`
	Status   string `json:"status,omitempty"`
	Type     string `json:"type,omitempty"`
	Priority string `json:"priority,omitempty"`

	Labels      []string `json:"labels,omitempty"`
	SetLabels   bool     `json:"set_labels,omitempty"  jsonschema:"set to true to replace labels (defaults: only sent when non-nil)"`
	SourceRefs  []string `json:"source_refs,omitempty"`
	SetSources  bool     `json:"set_source_refs,omitempty"`

	ParentID      string `json:"parent_id,omitempty"`
	ClearParentID bool   `json:"clear_parent_id,omitempty" jsonschema:"clear parent_id; takes precedence over parent_id when true"`

	Assignee      string `json:"assignee,omitempty"`
	ClearAssignee bool   `json:"clear_assignee,omitempty"`
}

// BacklogUpdateResult is the return shape of `backlog.update`.
type BacklogUpdateResult struct {
	Issue backlogclient.Issue `json:"issue"`
}

func (s *Server) handleBacklogUpdate(ctx context.Context, _ *mcp.CallToolRequest, args BacklogUpdateArgs) (*mcp.CallToolResult, BacklogUpdateResult, error) {
	if s.backlog == nil {
		return nil, BacklogUpdateResult{}, errBacklogNotConfigured
	}
	if args.ID == "" {
		return nil, BacklogUpdateResult{}, errors.New("backlog.update: id required")
	}
	if args.ExpectedVersion <= 0 {
		return nil, BacklogUpdateResult{}, errors.New("backlog.update: expected_version required (fetch via backlog.get first)")
	}
	req := backlogclient.IssueUpdateRequest{ExpectedVersion: args.ExpectedVersion}
	if args.Title != "" {
		req.Title = strPtr(args.Title)
	}
	if args.Body != "" {
		req.Body = strPtr(args.Body)
	}
	if args.Status != "" {
		req.Status = strPtr(args.Status)
	}
	if args.Type != "" {
		req.Type = strPtr(args.Type)
	}
	if args.Priority != "" {
		req.Priority = strPtr(args.Priority)
	}
	if args.SetLabels || len(args.Labels) > 0 {
		labels := args.Labels
		if labels == nil {
			labels = []string{}
		}
		req.Labels = &labels
	}
	if args.SetSources || len(args.SourceRefs) > 0 {
		refs := args.SourceRefs
		if refs == nil {
			refs = []string{}
		}
		req.SourceRefs = &refs
	}
	switch {
	case args.ClearParentID:
		empty := ""
		req.ParentID = &empty
	case args.ParentID != "":
		req.ParentID = strPtr(args.ParentID)
	}
	switch {
	case args.ClearAssignee:
		empty := ""
		req.Assignee = &empty
	case args.Assignee != "":
		req.Assignee = strPtr(args.Assignee)
	}
	issue, err := s.backlog.Update(ctx, args.ID, req)
	if err != nil {
		return nil, BacklogUpdateResult{}, err
	}
	return nil, BacklogUpdateResult{Issue: *issue}, nil
}

// backlog.delete ---------------------------------------------------------

// BacklogDeleteArgs are the args of `backlog.delete`.
type BacklogDeleteArgs struct {
	ID string `json:"id"`
}

// BacklogDeleteResult is the return shape of `backlog.delete`.
type BacklogDeleteResult struct {
	Deleted bool `json:"deleted"`
}

func (s *Server) handleBacklogDelete(ctx context.Context, _ *mcp.CallToolRequest, args BacklogDeleteArgs) (*mcp.CallToolResult, BacklogDeleteResult, error) {
	if s.backlog == nil {
		return nil, BacklogDeleteResult{}, errBacklogNotConfigured
	}
	if args.ID == "" {
		return nil, BacklogDeleteResult{}, errors.New("backlog.delete: id required")
	}
	if err := s.backlog.Delete(ctx, args.ID); err != nil {
		return nil, BacklogDeleteResult{}, err
	}
	return nil, BacklogDeleteResult{Deleted: true}, nil
}

// backlog.take_next ------------------------------------------------------

// BacklogTakeNextArgs are the args of `backlog.take_next`. ProjectID is
// optional; if omitted, the claim runs across the master backlog (every
// project).
type BacklogTakeNextArgs struct {
	ProjectID string `json:"project_id,omitempty"`
	Assignee  string `json:"assignee,omitempty" jsonschema:"explicit assignee override; defaults to the workbench actor (the X-Workbench-Actor header)"`
}

// BacklogTakeNextResult is the return shape of `backlog.take_next`. Found
// is false when no `todo` issue matched the filter; Issue is the
// zero-value in that case.
type BacklogTakeNextResult struct {
	Found   bool                `json:"found"`
	Issue   backlogclient.Issue `json:"issue,omitempty"`
	Message string              `json:"message,omitempty"`
}

func (s *Server) handleBacklogTakeNext(ctx context.Context, _ *mcp.CallToolRequest, args BacklogTakeNextArgs) (*mcp.CallToolResult, BacklogTakeNextResult, error) {
	if s.backlog == nil {
		return nil, BacklogTakeNextResult{}, errBacklogNotConfigured
	}
	issue, err := s.backlog.TakeNext(ctx, backlogclient.TakeNextRequest{
		ProjectID: args.ProjectID,
		Assignee:  args.Assignee,
	})
	if err != nil {
		return nil, BacklogTakeNextResult{}, err
	}
	if issue == nil {
		msg := "no todo issues available"
		if args.ProjectID != "" {
			msg = fmt.Sprintf("no todo issues available in project %q", args.ProjectID)
		}
		return nil, BacklogTakeNextResult{Found: false, Message: msg}, nil
	}
	return nil, BacklogTakeNextResult{Found: true, Issue: *issue}, nil
}

// registerBacklog wires the six backlog.* tools on srv.
func (s *Server) registerBacklog(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "backlog.add",
		Description: "Create an issue in the backlog. project_id defaults to the currently-selected project. Use source_refs to link back to originating notes (workbench:///notes/{id}) without modifying them.",
	}, s.handleBacklogAdd)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "backlog.list",
		Description: "List issues with optional filters. No filter = master backlog across all projects.",
	}, s.handleBacklogList)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "backlog.get",
		Description: "Read a single issue by id (including its current version, needed for backlog.update).",
	}, s.handleBacklogGet)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "backlog.update",
		Description: "Patch an issue. expected_version is required for OCC — fetch via backlog.get (or remember from add/take_next) first. Use clear_assignee / clear_parent_id to null those fields.",
	}, s.handleBacklogUpdate)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "backlog.delete",
		Description: "Delete an issue by id.",
	}, s.handleBacklogDelete)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "backlog.take_next",
		Description: "Atomically claim the next todo issue (highest priority, oldest first), assigning it to the workbench actor and flipping status to in_progress. Returns {found: false} if none available.",
	}, s.handleBacklogTakeNext)
}

func strPtr(s string) *string { return &s }
