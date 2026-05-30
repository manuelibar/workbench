# Workbench MCP — agent onboarding

Welcome. This MCP is the daily workbench for a single-user, agent-driven
development workflow. It exposes CRUD primitives — never built-in content —
and changes its visible surface as you select a working scope.

## Core concepts

- **WorkSession**: a daily-scoped, persistent session owned by the workbench.
  Every client connected to this server (Claude Code, Codex, …) shares the
  same WorkSession; selection state is shared across them.
- **Selection**: the currently-active scope: namespace → project → artifact
  or blueprint → mode, plus optional `focus`. Tools and resources visible at
  any moment depend on what's selected.
- **Refresh**: the single sync verb. Call `refresh` to re-evaluate state.
  Pass selection arguments to change scope; the response includes the new
  tool/resource list inline. The server also emits
  `notifications/{tools,resources,prompts}/list_changed`.

## Always-on tools (any state)

- `refresh(namespace_id?, project_id?, artifact_id?, blueprint_id?,
  mode_name?, focus?, clear?, clear_artifact?, clear_focus?)` — sync +
  (optionally) change selection.
- `ask(question, schema?)` — ask the user a question via elicitation.

(More CRUD primitives — `note.*`, `namespace.*`, etc. — surface in later
phases. Always call `tools/list` if unsure what's available.)

## Surface evolution

- Select a namespace → namespace mutations + `project.{create,list,…}`.
- Select a project → `artifact.*`, `skill.*`, `prompt.*`, `blueprint.*`,
  `backlog.*`.
- Select an artifact → `artifact.{guidance,validate,elicit}` for typed
  authoring.
- Select a blueprint → `mode.*`.
- Select a mode → user-defined tools/resources/prompts surface (post-v0).

## Discipline

- Never assume a tool exists. Call `tools/list` if unsure.
- Treat the response from `refresh` as the source of truth for state.
- Use `artifact.begin` for durable project/process assets such as RFCs, ADRs,
  PRDs, specs, plans, runbooks, and postmortems.
- Notes are the universal Zettelkasten capture primitive (once available);
  use them liberally. Promotion to typed artifacts is a separate workflow you
  trigger explicitly.
