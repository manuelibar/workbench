# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Changed

- Re-founded `main` as a stdio MCP context and artifact kernel.
- Replaced the old `refresh` selection flow with the `contextualize` tool.
- Replaced Postgres-backed artifacts with flat Markdown files under
  `docs/artifacts/`.

### Removed

- HTTP server, health/readiness endpoints, Postgres store and migrations,
  backlog, notes, projects, namespaces, roles, skills, prompts, blueprints,
  modes, onboarding, and snapshot planning surfaces from `main`.

### Added

- In-memory concurrent context state with tri-state `focus` and `artifact_id`
  patch semantics.
- Deterministic capability planning and list-relist synchronization with a
  configurable timeout and fallback capability index.
- File-backed artifact tools: `artifact.begin`, `artifact.list`,
  `artifact.get`, `artifact.update`, `artifact.guidance`, and
  `artifact.validate`.
- Resources `workbench:///context` and `workbench:///artifacts/{id}`.

### Breaking

- `workbench-mcp` now speaks MCP over stdio only. Existing HTTP, Postgres, and
  feature-tool clients must move to the relevant epic branches or adapt to the
  new context/artifact kernel.

## [v0.1.0] - 2026-05-10

### Added

- Initial Workbench MCP prototype with HTTP transport, Postgres persistence,
  and early feature surfaces.
