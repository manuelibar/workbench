package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/manuelibar/workbench/internal/mcpserver/skills"
)

type refreshIn struct {
	ProjectID   string `json:"project_id,omitempty"`
	NamespaceID string `json:"namespace_id,omitempty"`
	RoleID      string `json:"role_id,omitempty"`
	BoardID     string `json:"board_id,omitempty"`
}

// handleRefresh is the synchronization point for capability updates.
func (s *Server) handleRefresh(ctx context.Context, _ *mcp.CallToolRequest, in refreshIn) (*mcp.CallToolResult, RefreshResult, error) {
	s.mu.Lock()

	var sel selection
	if in.ProjectID != "" {
		id, err := uuid.Parse(in.ProjectID)
		if err != nil {
			s.mu.Unlock()
			return nil, RefreshResult{}, fmt.Errorf("invalid project_id: %w", err)
		}
		if _, err := s.store.Get(id); err != nil {
			s.mu.Unlock()
			if errors.Is(err, ErrProjectNotFound) {
				return nil, RefreshResult{}, fmt.Errorf("project not found: %s", in.ProjectID)
			}
			return nil, RefreshResult{}, err
		}
		sel.ProjectID = &id
	}
	if in.NamespaceID != "" {
		id, err := uuid.Parse(in.NamespaceID)
		if err != nil {
			s.mu.Unlock()
			return nil, RefreshResult{}, fmt.Errorf("invalid namespace_id: %w", err)
		}
		if _, ok := s.namespaces[id]; !ok {
			s.mu.Unlock()
			return nil, RefreshResult{}, fmt.Errorf("namespace not found: %s", in.NamespaceID)
		}
		sel.NamespaceID = &id
	}
	if in.RoleID != "" {
		id, err := uuid.Parse(in.RoleID)
		if err != nil {
			s.mu.Unlock()
			return nil, RefreshResult{}, fmt.Errorf("invalid role_id: %w", err)
		}
		if _, ok := s.roles[id]; !ok {
			s.mu.Unlock()
			return nil, RefreshResult{}, fmt.Errorf("role not found: %s", in.RoleID)
		}
		sel.RoleID = &id
	}
	if in.BoardID != "" {
		id, err := uuid.Parse(in.BoardID)
		if err != nil {
			s.mu.Unlock()
			return nil, RefreshResult{}, fmt.Errorf("invalid board_id: %w", err)
		}
		if _, ok := s.boards[id]; !ok {
			s.mu.Unlock()
			return nil, RefreshResult{}, fmt.Errorf("board not found: %s", in.BoardID)
		}
		sel.BoardID = &id
	}

	s.sel = sel
	s.refreshCapabilities(sel) // triggers debounced list-changed notifications
	tracker := s.startCapabilitySync([]string{"tools/list", "resources/list"})
	projCtx := s.projectContext(sel)
	bundles := s.skills.For(sel.ProjectID != nil)
	navigation := s.composeNavigation(sel)
	dynamicTools := append([]string(nil), s.dynamicTools...)
	overview := s.composeScopeOverview(sel, navigation, dynamicTools)
	capIndex := s.composeCapabilityIndex(sel, bundles, navigation, dynamicTools)

	s.mu.Unlock()

	// Wait for the SDK's 10 ms notification debounce to fire before pinging.
	time.Sleep(15 * time.Millisecond)

	capSync := s.waitForCapabilityRelist(ctx, tracker, capIndex)

	skillWires := make([]SkillWire, 0, len(bundles))
	for _, b := range bundles {
		instructions := ""
		for _, f := range b.Files {
			if f.RelPath == "SKILL.md" {
				instructions = string(f.Content(projCtx))
				break
			}
		}
		skillWires = append(skillWires, SkillWire{
			Name:         b.Name,
			Description:  b.Description,
			Version:      b.Version,
			Instructions: instructions,
		})
	}

	return nil, RefreshResult{
		Selection:      selectionToWire(sel),
		OverviewURI:    ScopeOverviewURI,
		Overview:       overview,
		Skills:         skillWires,
		Navigation:     navigation,
		CapabilitySync: capSync,
	}, nil
}

func (s *Server) composeCapabilityIndex(sel selection, bundles []skills.Bundle, navigation Navigation, dynamicTools []string) CapabilityIndexWire {
	tools := toolNames(dynamicTools)
	skillNames := make([]string, 0, len(bundles))
	for _, b := range bundles {
		skillNames = append(skillNames, b.Name)
	}
	sort.Strings(skillNames)
	return CapabilityIndexWire{Tools: tools, Resources: navigation.Available, Skills: skillNames}
}

func toolNames(dynamicTools []string) []string {
	tools := []string{"refresh", "feedback", "ask"}
	tools = append(tools, dynamicTools...)
	sort.Strings(tools)
	return tools
}

// composeScopeOverview builds the canonical scope briefing used by both
// refresh() and workbench:///scope/overview. It does not mutate selection or
// synchronize MCP capabilities.
func (s *Server) composeScopeOverview(sel selection, navigation Navigation, dynamicTools []string) ScopeOverview {
	overview := ScopeOverview{
		Selection:            selectionToWire(sel),
		Navigation:           navigation,
		RecommendedResources: append([]NavHint(nil), navigation.Available...),
		RecommendedPrompts:   []NavHint{},
		RecommendedTools:     toolNames(dynamicTools),
		KnowledgeHighlights:  []string{},
		TaskHighlights:       []string{},
		SelfResourceURI:      ScopeOverviewURI,
		CapabilityStateNote:  "Read-only snapshot. Call refresh() to change scope or synchronize MCP tool/resource/prompt lists; use MCP list methods for full capability manifests.",
	}
	if sel.ProjectID == nil {
		overview.Summary = "No project selected. Call refresh(project_id=<id>) to select one."
		return overview
	}
	p, err := s.store.Get(*sel.ProjectID)
	if err != nil {
		overview.Summary = "Project selection is set, but project details are unavailable. Call refresh() with a valid project_id to resynchronize."
		return overview
	}
	wire := projectToWire(p)
	overview.ActiveProject = &wire
	overview.Summary = "Project selected: " + p.Name
	return overview
}

// projectContext builds a skills.ProjectContext for the given selection.
// Must be called with s.mu held.
func (s *Server) projectContext(sel selection) skills.ProjectContext {
	if sel.ProjectID == nil {
		return skills.ProjectContext{}
	}
	p, err := s.store.Get(*sel.ProjectID)
	if err != nil {
		return skills.ProjectContext{}
	}
	return skills.ProjectContext{
		Name:         p.Name,
		Description:  p.Description,
		SystemPrompt: p.SystemPrompt,
	}
}

// composeNavigation builds navigation hints for the given selection.
// Safe to call concurrently; reads only sel (a value) and s.skills (immutable).
func (s *Server) composeNavigation(sel selection) Navigation {
	hasProject := sel.ProjectID != nil

	hints := []NavHint{
		{
			URI:         ScopeOverviewURI,
			Description: "Session overview and next recommended resources",
			Purpose:     "orient",
			MIMEType:    "application/json",
		},
		{
			URI:         "workbench:///knowledge",
			Description: "Queryable feedback and project knowledge",
			Purpose:     "knowledge",
			MIMEType:    "application/json",
		},
	}

	if hasProject {
		hints = append(hints, NavHint{
			URI:         "workbench:///tasks",
			Description: "Task state machine view for implementation work",
			Purpose:     "tasks",
			MIMEType:    "application/json",
		})
		hints = append(hints, NavHint{
			URI:         "workbench:///context/project-snapshot",
			Description: "README/docs/tech-stack snapshot for the configured project root",
			Purpose:     "context",
			MIMEType:    "application/json",
		})
		hints = append(hints, NavHint{
			URI:         "workbench:///projects/" + sel.ProjectID.String(),
			Description: "Details and system prompt for the selected project",
			Purpose:     "context",
			MIMEType:    "application/json",
		})
		hints = append(hints, NavHint{
			URI:         "workbench:///github/config",
			Description: "GitHub organization config supplied by MCP stdio environment",
			Purpose:     "integration",
			MIMEType:    "application/json",
		})
	}
	if sel.RoleID != nil {
		hints = append(hints, NavHint{URI: "workbench:///roles/" + sel.RoleID.String(), Description: "Selected role definition", Purpose: "role", MIMEType: "application/json"})
	}
	if sel.BoardID != nil {
		hints = append(hints, NavHint{URI: "workbench:///boards/" + sel.BoardID.String(), Description: "Selected board details", Purpose: "board", MIMEType: "application/json"})
	}

	for _, b := range s.skills.For(hasProject) {
		hints = append(hints, NavHint{
			URI:         "skill://" + b.Name + "/manifest",
			Description: b.Description,
			Purpose:     "skill",
			MIMEType:    "application/json",
		})
	}

	return Navigation{
		EntryURI:  ScopeOverviewURI,
		Available: hints,
	}
}
