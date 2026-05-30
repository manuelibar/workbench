package mcpserver

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/manuelibar/workbench/internal/domain"
)

// RefreshArgs are the optional selection-mutation parameters of the
// `refresh` tool. Empty struct = "just sync, don't change selection".
type RefreshArgs struct {
	NamespaceID   string `json:"namespace_id,omitempty" jsonschema:"select this namespace by UUID; ignored if empty"`
	ProjectID     string `json:"project_id,omitempty"   jsonschema:"select this project by UUID; ignored if empty"`
	ArtifactID    string `json:"artifact_id,omitempty"  jsonschema:"select this artifact by UUID; resolves its project/namespace and clears blueprint/mode"`
	BlueprintID   string `json:"blueprint_id,omitempty" jsonschema:"select this blueprint by UUID; ignored if empty"`
	ModeName      string `json:"mode_name,omitempty"    jsonschema:"select this mode (within the selected blueprint); ignored if empty"`
	Focus         string `json:"focus,omitempty"        jsonschema:"persist a short steering string for the next workbench capability surface"`
	Clear         bool   `json:"clear,omitempty"        jsonschema:"if true, clear the entire selection; takes precedence over the other fields"`
	ClearArtifact bool   `json:"clear_artifact,omitempty" jsonschema:"if true, clear only the selected artifact"`
	ClearFocus    bool   `json:"clear_focus,omitempty"    jsonschema:"if true, clear only the persisted focus"`
}

// HasMutation reports whether args carry any selection change.
func (a RefreshArgs) HasMutation() bool {
	return a.Clear || a.ClearArtifact || a.ClearFocus ||
		a.NamespaceID != "" || a.ProjectID != "" || a.ArtifactID != "" ||
		a.BlueprintID != "" || a.ModeName != "" || a.Focus != ""
}

// RefreshResult is what `refresh` returns: the live selection plus a
// summary of what's currently visible to the agent.
type RefreshResult struct {
	Selection    SelectionWire     `json:"selection"`
	Tools        []ToolSummary     `json:"tools"`
	Resources    []ResourceSummary `json:"resources"`
	Prompts      []PromptSummary   `json:"prompts"`
	RecentEvents []EventSummary    `json:"recent_events"`
	Synced       bool              `json:"synced"`
}

// SelectionWire is the JSON-friendly projection of [domain.Selection]
// used in tool results. UUIDs travel as canonical strings so the
// auto-generated output schema matches the runtime payload (a *uuid.UUID
// reflects to `[16]byte` which jsonschema-go would describe as type
// "array", colliding with the actual JSON string form).
type SelectionWire struct {
	NamespaceID string `json:"namespace_id,omitempty"`
	ProjectID   string `json:"project_id,omitempty"`
	ArtifactID  string `json:"artifact_id,omitempty"`
	BlueprintID string `json:"blueprint_id,omitempty"`
	ModeName    string `json:"mode_name,omitempty"`
	Focus       string `json:"focus,omitempty"`
}

// IsEmpty reports whether no selection field is set.
func (s SelectionWire) IsEmpty() bool {
	return s.NamespaceID == "" && s.ProjectID == "" && s.ArtifactID == "" && s.BlueprintID == "" && s.ModeName == "" && s.Focus == ""
}

func selectionToWire(sel domain.Selection) SelectionWire {
	w := SelectionWire{ModeName: sel.ModeName, Focus: sel.Focus}
	if sel.NamespaceID != nil {
		w.NamespaceID = sel.NamespaceID.String()
	}
	if sel.ProjectID != nil {
		w.ProjectID = sel.ProjectID.String()
	}
	if sel.ArtifactID != nil {
		w.ArtifactID = sel.ArtifactID.String()
	}
	if sel.BlueprintID != nil {
		w.BlueprintID = sel.BlueprintID.String()
	}
	return w
}

// ToolSummary, ResourceSummary, PromptSummary, EventSummary are the
// agent-facing projections returned inline in [RefreshResult] so the agent
// has the new shape without round-tripping `tools/list` etc.
type ToolSummary struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ResourceSummary struct {
	URI         string `json:"uri"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	MIMEType    string `json:"mime_type,omitempty"`
}

type PromptSummary struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type EventSummary struct {
	OccurredAt  time.Time `json:"occurred_at"`
	Type        string    `json:"type"`
	SubjectKind string    `json:"subject_kind,omitempty"`
	SubjectID   string    `json:"subject_id,omitempty"`
}

func (s *Server) registerRefresh(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "refresh",
		Description: "Sync workbench state and (optionally) change the selection. Returns the current selection plus a summary of currently-visible tools, resources, prompts, and recent events. The server also emits notifications/{tools,resources,prompts}/list_changed when the surface changes.",
	}, s.handleRefresh)
}

func (s *Server) handleRefresh(ctx context.Context, _ *mcp.CallToolRequest, args RefreshArgs) (*mcp.CallToolResult, RefreshResult, error) {
	res, err := s.Refresh(ctx, args)
	return nil, res, err
}

// Refresh applies any selection mutation in args (persisting to the store and
// rebuilding the singleton [*mcp.Server]'s tool surface), then returns the
// current state. It is the single entry point for selection changes; tool
// handlers route through it.
func (s *Server) Refresh(ctx context.Context, args RefreshArgs) (RefreshResult, error) {
	s.selectionMu.Lock()
	defer s.selectionMu.Unlock()

	if args.HasMutation() {
		newSel, err := s.applyArgsPatch(ctx, s.selection, args)
		if err != nil {
			return RefreshResult{}, err
		}
		if err := s.store.UpdateSelection(ctx, s.user.ID, newSel); err != nil {
			return RefreshResult{}, fmt.Errorf("refresh: persist selection: %w", err)
		}
		s.selection = newSel
		s.applyVisibility(s.selection)
	}

	events, err := s.store.RecentEvents(ctx, s.workSessionID, 0)
	if err != nil {
		return RefreshResult{}, fmt.Errorf("refresh: read recent events: %w", err)
	}

	return RefreshResult{
		Selection:    selectionToWire(s.selection),
		Tools:        s.toolSummariesForSelection(s.selection),
		Resources:    s.alwaysOnResourceSummaries(),
		Prompts:      []PromptSummary{},
		RecentEvents: toEventSummaries(events),
		Synced:       true,
	}, nil
}

// applyArgsPatch merges args into the current selection with these rules:
//
//   - args.Clear: returns the empty selection.
//   - args.NamespaceID set: jumps to a fresh namespace selection
//     (project / artifact / blueprint / mode are cleared; focus is preserved).
//   - args.ProjectID set: overrides ProjectID, clears blueprint/mode, and
//     auto-resolves the namespace from the project if NamespaceID isn't
//     also in args and isn't already set.
//   - args.ArtifactID set: overrides ArtifactID, auto-resolves project and
//     namespace from the artifact, and clears blueprint/mode.
//   - args.BlueprintID set: overrides BlueprintID, clears mode, and
//     auto-resolves project (and namespace through project) if missing.
//   - args.ModeName set: overrides ModeName, leaves the rest alone.
//   - args.Focus set: persists the supplied steering string.
//   - args.ClearArtifact / args.ClearFocus clear only those fields.
//
// Order of application: namespace → project → artifact → blueprint → mode,
// then clear flags/focus, so passing
// multiple fields in one args produces the union the caller asked for.
func (s *Server) applyArgsPatch(ctx context.Context, base domain.Selection, args RefreshArgs) (domain.Selection, error) {
	if args.Clear {
		return domain.Selection{}, nil
	}
	out := base
	if args.NamespaceID != "" {
		nsID, err := uuid.Parse(args.NamespaceID)
		if err != nil {
			return domain.Selection{}, fmt.Errorf("refresh: namespace_id: %w", err)
		}
		out = domain.Selection{NamespaceID: &nsID, Focus: out.Focus}
	}
	if args.ProjectID != "" {
		pID, err := uuid.Parse(args.ProjectID)
		if err != nil {
			return domain.Selection{}, fmt.Errorf("refresh: project_id: %w", err)
		}
		out.ProjectID = &pID
		out.ArtifactID = nil
		out.BlueprintID = nil
		out.ModeName = ""
		if out.NamespaceID == nil {
			if p, err := s.store.GetProject(ctx, pID); err == nil && p.NamespaceID != nil {
				out.NamespaceID = p.NamespaceID
			}
		}
	}
	if args.ArtifactID != "" {
		aID, err := uuid.Parse(args.ArtifactID)
		if err != nil {
			return domain.Selection{}, fmt.Errorf("refresh: artifact_id: %w", err)
		}
		a, err := s.store.GetArtifact(ctx, aID)
		if err != nil {
			return domain.Selection{}, fmt.Errorf("refresh: artifact: %w", err)
		}
		pID := a.ProjectID
		out.ProjectID = &pID
		out.ArtifactID = &aID
		out.BlueprintID = nil
		out.ModeName = ""
		if p, err := s.store.GetProject(ctx, pID); err == nil && p.NamespaceID != nil {
			out.NamespaceID = p.NamespaceID
		}
	}
	if args.BlueprintID != "" {
		bID, err := uuid.Parse(args.BlueprintID)
		if err != nil {
			return domain.Selection{}, fmt.Errorf("refresh: blueprint_id: %w", err)
		}
		out.ArtifactID = nil
		out.BlueprintID = &bID
		out.ModeName = ""
		if b, err := s.store.GetBlueprint(ctx, bID); err == nil {
			if out.ProjectID == nil {
				pID := b.ProjectID
				out.ProjectID = &pID
			}
			if out.NamespaceID == nil {
				if p, err := s.store.GetProject(ctx, b.ProjectID); err == nil && p.NamespaceID != nil {
					out.NamespaceID = p.NamespaceID
				}
			}
		}
	}
	if args.ModeName != "" {
		out.ModeName = args.ModeName
	}
	if args.ClearArtifact {
		out.ArtifactID = nil
	}
	if args.Focus != "" {
		out.Focus = strings.TrimSpace(args.Focus)
	}
	if args.ClearFocus {
		out.Focus = ""
	}
	return out, nil
}

// alwaysOnResourceSummaries lists resources that should appear in
// `refresh.resources` regardless of selection. v0 has just one: the
// embedded onboarding skill at workbench://skill.
func (s *Server) alwaysOnResourceSummaries() []ResourceSummary {
	return []ResourceSummary{
		{
			URI:         skillResourceURI,
			Name:        "Workbench onboarding skill",
			Description: "Top-level agent onboarding document; read this first.",
			MIMEType:    "text/markdown",
		},
	}
}

// toEventSummaries converts a [domain.Event] slice into the agent-facing
// [EventSummary] projection used in [RefreshResult.RecentEvents].
func toEventSummaries(events []domain.Event) []EventSummary {
	out := make([]EventSummary, len(events))
	for i, e := range events {
		out[i] = EventSummary{
			OccurredAt:  e.OccurredAt,
			Type:        e.Type,
			SubjectKind: e.SubjectKind,
			SubjectID:   e.SubjectID,
		}
	}
	return out
}
