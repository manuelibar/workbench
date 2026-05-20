package mcpserver

import "time"

// ProjectWire is the JSON shape for a project.
type ProjectWire struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	SystemPrompt string `json:"system_prompt,omitempty"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// SelectionWire is the JSON shape for a session selection.
type SelectionWire struct {
	ProjectID   string `json:"project_id,omitempty"`
	NamespaceID string `json:"namespace_id,omitempty"`
	RoleID      string `json:"role_id,omitempty"`
	BoardID     string `json:"board_id,omitempty"`
}

// NavHint is a single navigation suggestion surfaced by scope/overview.
type NavHint struct {
	URI         string `json:"uri"`
	Description string `json:"description"`
	Purpose     string `json:"purpose"`
	MIMEType    string `json:"mime_type,omitempty"`
}

// Navigation is the navigation block surfaced by scope/overview.
type Navigation struct {
	EntryURI  string    `json:"entry_uri"`
	Available []NavHint `json:"available"`
}

// SkillWire is a skill bundle with its instructions rendered for the session.
type SkillWire struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Version      string `json:"version"`
	Instructions string `json:"instructions"` // rendered SKILL.md content
}

const ScopeOverviewURI = "workbench:///scope/overview"

// ScopeOverview is the shared read model returned inline by refresh() and by
// the read-only workbench:///scope/overview resource.
type ScopeOverview struct {
	Selection            SelectionWire `json:"selection"`
	Summary              string        `json:"summary"`
	ActiveProject        *ProjectWire  `json:"active_project,omitempty"`
	Navigation           Navigation    `json:"navigation"`
	RecommendedResources []NavHint     `json:"recommended_resources"`
	RecommendedPrompts   []NavHint     `json:"recommended_prompts"`
	RecommendedTools     []string      `json:"recommended_tools"`
	KnowledgeHighlights  []string      `json:"knowledge_highlights"`
	TaskHighlights       []string      `json:"task_highlights"`
	CapabilityStateNote  string        `json:"capability_state_note"`
	SelfResourceURI      string        `json:"self_resource_uri"`
}

// RefreshResult is the output of the refresh tool.
// It carries the full working context in one response: the active selection,
// a shared overview read model, and every skill bundle relevant to that
// selection with instructions inlined. Resources remain available for
// on-demand re-fetch after context compaction.
type RefreshResult struct {
	Selection      SelectionWire      `json:"selection"`
	OverviewURI    string             `json:"overview_uri"`
	Overview       ScopeOverview      `json:"overview"`
	Skills         []SkillWire        `json:"skills"`
	Navigation     Navigation         `json:"navigation"`
	CapabilitySync CapabilitySyncWire `json:"capability_sync"`
}

type CapabilitySyncWire struct {
	Generation int64                `json:"generation"`
	Status     string               `json:"status"`
	Required   []string             `json:"required"`
	Observed   []string             `json:"observed"`
	TimedOut   bool                 `json:"timed_out"`
	Index      *CapabilityIndexWire `json:"index,omitempty"`
}

type CapabilityIndexWire struct {
	Tools     []string  `json:"tools"`
	Resources []NavHint `json:"resources"`
	Skills    []string  `json:"skills"`
}

func projectToWire(p Project) ProjectWire {
	return ProjectWire{
		ID:           p.ID.String(),
		Name:         p.Name,
		Description:  p.Description,
		SystemPrompt: p.SystemPrompt,
		CreatedAt:    p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    p.UpdatedAt.Format(time.RFC3339),
	}
}

func selectionToWire(sel selection) SelectionWire {
	out := SelectionWire{}
	if sel.ProjectID != nil {
		out.ProjectID = sel.ProjectID.String()
	}
	if sel.NamespaceID != nil {
		out.NamespaceID = sel.NamespaceID.String()
	}
	if sel.RoleID != nil {
		out.RoleID = sel.RoleID.String()
	}
	if sel.BoardID != nil {
		out.BoardID = sel.BoardID.String()
	}
	return out
}
