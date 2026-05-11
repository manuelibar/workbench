package mcpserver

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/manuelibar/workbench/internal/domain"
)

// toolGroup is a named collection of MCP tools whose visibility on the live
// [*mcp.Server] is controlled by [toolGroup.visible] applied to the current
// [domain.Selection]. [toolGroup.register] registers every tool in the
// group; the SDK's `AddTool` replaces existing tools by name, so re-registering
// an already-registered group is safe and idempotent.
type toolGroup struct {
	name         string
	names        []string
	descriptions map[string]string
	visible      func(domain.Selection) bool
	register     func(s *Server, srv *mcp.Server)
}

// anySelection is the visibility predicate for tools that should be exposed
// regardless of the current selection.
func anySelection(_ domain.Selection) bool { return true }

// namespaceSelected reports whether [domain.Selection.NamespaceID] is set.
func namespaceSelected(sel domain.Selection) bool { return sel.NamespaceID != nil }

// projectSelected reports whether [domain.Selection.ProjectID] is set.
func projectSelected(sel domain.Selection) bool { return sel.ProjectID != nil }

// blueprintSelected reports whether [domain.Selection.BlueprintID] is set.
func blueprintSelected(sel domain.Selection) bool { return sel.BlueprintID != nil }

// toolGroups returns the static catalog of every tool group the server knows
// about, in stable order. The catalog is derived per-instance because each
// group's `register` closes over the parent [*Server] (for store/user/log
// access in the handlers).
func (s *Server) toolGroups() []toolGroup {
	return []toolGroup{
		{
			name:  "core",
			names: []string{"refresh", "ask"},
			descriptions: map[string]string{
				"refresh": "Sync workbench state and (optionally) change the selection. Returns the new selection plus visible tools/resources/prompts and recent events.",
				"ask":     "Ask the user a question via MCP elicitation.",
			},
			visible: anySelection,
			register: func(s *Server, srv *mcp.Server) {
				s.registerRefresh(srv)
				s.registerAsk(srv)
			},
		},
		{
			name:  "notes",
			names: []string{"note.add", "note.list", "note.search", "note.get", "note.update", "note.delete"},
			descriptions: map[string]string{
				"note.add":    "Capture a quick markdown note (Zettelkasten primitive).",
				"note.list":   "List notes most-recent-first, with optional filters.",
				"note.search": "Substring search across note bodies.",
				"note.get":    "Fetch a single note by id.",
				"note.update": "Patch a note's body and/or tags.",
				"note.delete": "Delete a note by id.",
			},
			visible:  anySelection,
			register: func(s *Server, srv *mcp.Server) { s.registerNotes(srv) },
		},
		{
			name:  "backlog",
			names: []string{"backlog.add", "backlog.list", "backlog.get", "backlog.update", "backlog.delete", "backlog.take_next"},
			descriptions: map[string]string{
				"backlog.add":       "Create an issue in the backlog. project_id defaults to the currently-selected project. source_refs link back to originating notes.",
				"backlog.list":      "List issues with optional filters. No filter = master backlog across all projects.",
				"backlog.get":       "Read a single issue (including its version, needed for backlog.update).",
				"backlog.update":    "Patch an issue. expected_version (OCC) is required.",
				"backlog.delete":    "Delete an issue by id.",
				"backlog.take_next": "Atomically claim the next todo issue (priority+age), assigning it to the workbench actor.",
			},
			visible:  anySelection,
			register: func(s *Server, srv *mcp.Server) { s.registerBacklog(srv) },
		},
		{
			name:  "namespace-bootstrap",
			names: []string{"namespace.create", "namespace.list"},
			descriptions: map[string]string{
				"namespace.create": "Create a new namespace (root or under a parent).",
				"namespace.list":   "List namespaces under a parent (or root namespaces).",
			},
			visible:  anySelection,
			register: func(s *Server, srv *mcp.Server) { s.registerNamespaceBootstrap(srv) },
		},
		{
			name:  "namespace-scoped",
			names: []string{"namespace.get", "namespace.update", "namespace.delete", "project.create", "project.list"},
			descriptions: map[string]string{
				"namespace.get":    "Fetch the selected (or named) namespace.",
				"namespace.update": "Patch the selected (or named) namespace.",
				"namespace.delete": "Delete the selected (or named) namespace; children cascade.",
				"project.create":   "Create a project under the selected (or named) namespace.",
				"project.list":     "List projects under the selected (or named) namespace.",
			},
			visible: namespaceSelected,
			register: func(s *Server, srv *mcp.Server) {
				s.registerNamespaceScoped(srv)
				s.registerProjectBootstrap(srv)
			},
		},
		{
			name: "project-scoped",
			names: []string{
				"project.get", "project.update", "project.delete",
				"artifact.create", "artifact.list", "artifact.get", "artifact.update", "artifact.delete",
				"artifact.attach", "artifact.sign_off", "artifact.archive",
				"skill.create", "skill.list", "skill.get", "skill.update", "skill.delete",
				"prompt.create", "prompt.list", "prompt.get", "prompt.update", "prompt.delete",
				"blueprint.create", "blueprint.list", "blueprint.get", "blueprint.update", "blueprint.delete",
			},
			descriptions: map[string]string{
				"project.get":    "Fetch the currently-selected project.",
				"project.update": "Patch the currently-selected project.",
				"project.delete": "Delete the currently-selected project; cascades to artifacts/skills/prompts/blueprints.",

				"artifact.create":   "Create a typed, versioned artifact in the selected project.",
				"artifact.list":     "List artifacts in the selected project.",
				"artifact.get":      "Fetch an artifact and one of its versions.",
				"artifact.update":   "Append a new version and/or change status.",
				"artifact.delete":   "Delete an artifact and all its versions.",
				"artifact.attach":   "Attach a parent artifact to record lineage.",
				"artifact.sign_off": "Move an artifact to status='signed_off'.",
				"artifact.archive":  "Move an artifact to status='archived'.",

				"skill.create": "Create a markdown skill in the selected project.",
				"skill.list":   "List skills in the selected project.",
				"skill.get":    "Fetch a skill by id or name.",
				"skill.update": "Patch a skill's name and/or body.",
				"skill.delete": "Delete a skill by id.",

				"prompt.create": "Create a prompt template in the selected project.",
				"prompt.list":   "List prompts in the selected project.",
				"prompt.get":    "Fetch a prompt by id or name.",
				"prompt.update": "Patch a prompt's metadata or body.",
				"prompt.delete": "Delete a prompt by id.",

				"blueprint.create": "Create a brand-new blueprint at v1 in the selected project.",
				"blueprint.list":   "List blueprint rows in the selected project (latest_only=true collapses).",
				"blueprint.get":    "Fetch a blueprint by id, or by (name, version?).",
				"blueprint.update": "Append a new immutable version of an existing blueprint name.",
				"blueprint.delete": "Delete a single blueprint version (and its modes).",
			},
			visible: projectSelected,
			register: func(s *Server, srv *mcp.Server) {
				s.registerProjectFull(srv)
				s.registerArtifacts(srv)
				s.registerSkills(srv)
				s.registerPrompts(srv)
				s.registerBlueprints(srv)
			},
		},
		{
			name:  "blueprint-scoped",
			names: []string{"mode.create", "mode.list", "mode.get", "mode.update", "mode.delete"},
			descriptions: map[string]string{
				"mode.create": "Create a mode inside the currently-selected blueprint version (latest only).",
				"mode.list":   "List modes inside the currently-selected blueprint version.",
				"mode.get":    "Fetch a mode by id or name within the currently-selected blueprint.",
				"mode.update": "Patch a mode's metadata. Only allowed on the latest blueprint version.",
				"mode.delete": "Delete a mode by id. Only allowed on the latest blueprint version.",
			},
			visible:  blueprintSelected,
			register: func(s *Server, srv *mcp.Server) { s.registerModes(srv) },
		},
	}
}

// applyVisibility brings the singleton [*mcp.Server]'s tool surface in line
// with sel, using `RemoveTools` for tools that are no longer visible and
// `AddTool` (via each group's register fn) for tools that should be visible.
//
// MUST be called with [Server.selectionMu] held.
func (s *Server) applyVisibility(sel domain.Selection) {
	desired := map[string]bool{}
	for _, g := range s.toolGroups() {
		if g.visible(sel) {
			for _, n := range g.names {
				desired[n] = true
			}
		}
	}

	var toRemove []string
	for n := range s.activeTools {
		if !desired[n] {
			toRemove = append(toRemove, n)
		}
	}
	if len(toRemove) > 0 {
		s.sdkServer.RemoveTools(toRemove...)
	}

	for _, g := range s.toolGroups() {
		if g.visible(sel) {
			g.register(s, s.sdkServer)
		}
	}
	s.activeTools = desired
}

// toolSummariesForSelection returns the [ToolSummary] list reflecting which
// tools are visible at sel. Used by [refresh] to populate `tools` in the
// result.
//
// MUST be called with [Server.selectionMu] held when sel is read from
// [Server.selection]; the function itself does not take the lock.
func (s *Server) toolSummariesForSelection(sel domain.Selection) []ToolSummary {
	var out []ToolSummary
	for _, g := range s.toolGroups() {
		if !g.visible(sel) {
			continue
		}
		for _, n := range g.names {
			out = append(out, ToolSummary{Name: n, Description: g.descriptions[n]})
		}
	}
	return out
}
