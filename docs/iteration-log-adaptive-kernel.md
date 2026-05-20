# Iteration Log: Adaptive Workbench Kernel

Date: 2026-05-19

This log records the first five implementation/documentation iterations after ADR 0006 and RFC 0002. Each iteration keeps Workbench moving toward an adaptive, manager-oriented MCP server while preserving a working build.

## Iteration 1 — Project context snapshot

Implemented `ProjectSnapshot` indexing for a configured project root. Workbench can now detect:

- README title and short brief
- docs/*.md index
- basic tech stack from common files such as go.mod, package.json, pyproject.toml, Cargo.toml, Dockerfile

New resource:

- `workbench:///context/project-snapshot`

Configuration:

- `WORKBENCH_PROJECT_ROOT`, defaulting to the server working directory

Why this matters: refresh can guide the agent toward README/docs/tech-stack state without embedding all docs directly in the tool response.

## Iteration 2 — Filesystem skill registry overlay

Implemented a filesystem registry and overlay registry for skills.

Configuration:

- `WORKBENCH_SKILLS_DIR`

Behavior:

- Each direct child directory is a skill bundle if it contains `SKILL.md`.
- Filesystem skills are layered before embedded skills.
- Name collisions resolve in favor of the higher-priority registry.

Why this matters: skills are no longer only compiled into the binary. Local filesystem is the first concrete `SkillRegistry` implementation; future S3/Git/HTTP/IPFS registries can implement the same contract.

## Iteration 3 — Task state machine

Added a deterministic task state model.

States:

- proposed
- ready
- in_progress
- blocked
- review
- done
- cancelled

New project-scoped tools:

- `task.create`
- `task.list`
- `task.transition`

New resource:

- `workbench:///tasks`

Why this matters: agents can create and update work items, but Go owns the valid transitions. Headless agents may suggest transitions later, but deterministic server logic validates them.

## Iteration 4 — Knowledge and ask

Feedback is now stored as queryable knowledge instead of being only acknowledged.

Core tool:

- `ask`

Existing tool behavior changed:

- `feedback` stores feedback as `KnowledgeItem` records.

New resource:

- `workbench:///knowledge`

Current query behavior is deterministic substring matching. Future versions can add graph edges and semantic routing while preserving cited knowledge item outputs.

## Iteration 5 — Refresh/navigation contract refinement

Refresh now points agents at richer resources:

- knowledge
- tasks
- project snapshot
- project details
- integration config
- skills via MCP capability/resource discovery

The intended contract remains:

1. `refresh` updates selected scope and reconciles capabilities.
2. MCP list-changed notifications tell the client what changed.
3. The client lists capabilities.
4. The agent reads resources as needed.
5. The refresh response serves as a navigation briefing, not as the only source of truth.

## Verification

The iteration added tests for:

- project snapshot indexing
- filesystem skill loading
- task state transitions
- feedback-to-knowledge querying

Run:

```sh
gofmt -w cmd/workbench-mcp/main.go internal/mcpserver/*.go internal/mcpserver/skills/*.go
go test ./...
make build
make test
hermes mcp test workbench
```
