# Workbench Architecture Specification

Date: 2026-05-28
Status: Draft handoff specification

This document captures the current Workbench project understanding and the architecture direction discussed so far. It is intended as a handoff artifact for another agent or contributor before restructuring the codebase.

Workbench should evolve into a pragmatic Go implementation of DDD plus hexagonal architecture, without importing heavy Java-style ceremony. The core idea is simple:

- Model the real domain primitives explicitly.
- Keep dependency flow inward.
- Keep MCP, filesystems, Postgres, HTTP APIs, Codex CLI, and future integrations at the edges.
- Use idiomatic Go package design, where package names are domain concepts or concrete adapter roles.
- Avoid preemptive interfaces and generic architecture folders when they do not earn their cost.

## Current Project Summary

Workbench is currently a Go MCP stdio server.

It gives AI agents a scoped, adaptive working context. Instead of exposing every tool and resource all the time, the server lets an agent call `refresh()` to select scope and reconcile the current capability surface.

Implemented capabilities include:

- MCP stdio server entry point in `cmd/workbench-mcp`.
- Core MCP tools:
  - `refresh`
  - `feedback`
  - `query`
- Dynamic tools before project selection:
  - project create/list/delete
  - namespace create/list
  - role create/list
- Dynamic tools after project selection:
  - project list
  - namespace list
  - role list
  - task create/list/transition
  - board create/list
- Static resources:
  - `workbench:///scope/overview`
  - `workbench:///github/config`
  - `workbench:///context/project-snapshot`
  - `workbench:///tasks`
  - `workbench:///knowledge`
  - templates for projects, roles, and boards
- Dynamic skill resources:
  - `skill://{name}/manifest`
  - `skill://{name}/SKILL.md`
- Embedded seed skills:
  - `workbench-orient`
  - `workbench-system-prompt`
  - `go-coding-guidelines`
- Filesystem skill overlay through `WORKBENCH_SKILLS_DIR`.
- Project metadata store:
  - memory implementation for tests
  - JSON file implementation for local durability
- Initial Postgres migration foundation.
- In-memory namespaces, roles, boards, tasks, and knowledge.
- Feedback-backed local knowledge.
- KB-backed `query` path that can call:
  - `/content/search`
  - `/knowledge/query`
  - headless Codex synthesis
- Codex CLI synthesizer for `query`.
- Project snapshot indexing for README/docs/tech-stack discovery.
- Tests for refresh synchronization, dynamic resources, tasks, knowledge, project snapshots, filesystem skills, and migrations.

Current important limitation:

- Only projects are meaningfully durable through the file project store.
- Namespace, role, board, task, session, and fallback knowledge state are process-local.
- Postgres schema exists as foundation but project CRUD still uses the file-backed store.
- The current package `internal/mcpserver` mixes inbound MCP adapter code, application use cases, domain state, and adapter implementations.

## Desired Architecture Direction

Workbench should become an adaptive capability kernel exposed through MCP.

The project is meant to be exposed primarily through an MCP server, so using `internal/server` instead of `internal/mcpserver` is acceptable. However, `server` must remain an inbound adapter. It should translate MCP requests and responses. It should not own domain rules.

Dependency flow must point inward:

```text
cmd -> server -> application/domain packages <- adapter implementations
```

Core packages must not import:

- `github.com/modelcontextprotocol/go-sdk`
- Postgres/pgx/sql driver details
- filesystem implementation details unless the package is explicitly an adapter
- HTTP client implementation details unless the package is explicitly an adapter
- `os/exec` or Codex CLI details
- GitHub/Jira/Linear API clients
- S3 SDKs

The core should be testable without MCP.

A good architectural test:

- Project creation can be tested without MCP.
- Task transitions can be tested without MCP.
- Session selection can be tested without stdio.
- Refresh planning can be tested without registering MCP resources.
- Skill visibility can be tested without the MCP SDK.
- Knowledge lookup decisions can be tested without a real HTTP server.
- Agent execution can be faked without invoking Codex.

## Pragmatic Go DDD Principles

Use DDD concepts where they clarify real behavior:

- namespaces
- projects
- sessions
- tasks
- artifacts
- trace links
- skills
- knowledge
- refresh/scope composition

Avoid heavy DDD ceremony:

- Do not create empty `domain`, `application`, `infrastructure`, or `ports` directories unless they materially help dependency direction.
- Do not use repository interfaces "just in case".
- Do not put all interfaces in a `ports` package.
- Do not force a package split before import direction is clear.
- Do not use generic names like `utils`, `helpers`, `common`, or `storage` as broad dumping grounds.

Interfaces should usually live near the consuming package, not near the implementation.

Example:

```text
internal/project/repository.go       # project.Repository interface used by project.Manager
internal/repo/file/project.go        # file adapter satisfying project.Repository
internal/repo/postgres/project.go    # postgres adapter satisfying project.Repository
```

This avoids a central `ports` package and keeps Go interfaces small and consumer-owned.

## Package Layout Candidate

Preferred candidate:

```text
cmd/workbench-mcp/
  main.go

internal/server/
  server.go
  refresh_handler.go
  capability_sync.go

internal/server/tools/
  project_create.go
  project_list.go
  task_create.go
  task_transition.go
  knowledge_query.go

internal/server/resources/
  scope_overview.go
  project_snapshot.go
  skill.go
  task.go
  knowledge.go

internal/server/prompts/
  registry.go

internal/refresh/
  service.go
  capability.go
  overview.go

internal/project/
  project.go
  namespace.go
  role.go
  board.go
  manager.go
  repository.go

internal/task/
  task.go
  state.go
  manager.go
  repository.go

internal/artifact/
  artifact.go
  kind.go
  state.go
  manager.go
  repository.go

internal/session/
  session.go
  scope.go
  manager.go
  repository.go

internal/trace/
  ref.go
  link.go
  relation.go
  manager.go
  repository.go

internal/skills/
  bundle.go
  registry.go
  source.go
  spec.go
  manager.go

internal/skills/embedded/
  source.go

internal/skills/filesystem/
  source.go

internal/knowledge/
  item.go
  manager.go
  repository.go

internal/librarian/
  librarian.go

internal/librarian/http/
  client.go

internal/agent/
  runner.go

internal/agent/codex/
  runner.go

internal/repo/file/
  project.go
  task.go
  artifact.go
  session.go
  trace.go
  knowledge.go

internal/repo/postgres/
  db.go
  migrations.go
  project.go
  task.go
  artifact.go
  session.go
  trace.go
  knowledge.go

internal/tracker/
  tracker.go

internal/tracker/github/
  tracker.go
```

This is a candidate, not a mandate. The restructuring should be incremental.

## Why No Root `workbench` Package

An umbrella `internal/workbench` directory was rejected for now.

Reasoning:

- The repo already is Workbench.
- A `workbench` directory would mostly be an architectural label, not a domain concept.
- It adds path depth without clarifying responsibility.
- Package names like `project`, `task`, `session`, `skills`, and `refresh` are more direct.

If later there is a genuine public library or reusable kernel package, a higher-level package can be reconsidered.

## Repositories

The term `storage` was rejected as too generic and infrastructure-oriented.

Two repository organization options were considered:

Option A:

```text
internal/projectrepo/
internal/taskrepo/
internal/sessionrepo/
```

Option B:

```text
internal/repo/file/
internal/repo/postgres/
```

Preferred option: `internal/repo/<backend>`.

Reasoning:

- A real backend often implements multiple repositories.
- Postgres transaction boundaries may cross projects, tasks, artifacts, trace links, sessions, and knowledge.
- Grouping by backend keeps connection pooling, migrations, transaction helpers, and backend-specific concerns together.
- Package names avoid underscores, matching Go conventions.

Repository interfaces should not live in `internal/repo`.

They should live in the consuming domain/application package:

```text
internal/project/repository.go
internal/task/repository.go
internal/session/repository.go
internal/artifact/repository.go
internal/trace/repository.go
internal/knowledge/repository.go
```

The adapters in `internal/repo/file` and `internal/repo/postgres` satisfy those interfaces implicitly.

Open question:

- Should file-backed repositories remain a first-class backend long term, or only a local development/prototype adapter?

## Server Package and MCP Transport

The project is primarily exposed as an MCP server, so `internal/server` is acceptable.

However:

- `internal/server` is an inbound adapter.
- It depends on application/domain packages.
- It may import the MCP SDK.
- Core packages must not import `internal/server`.

The server should own:

- MCP server construction
- MCP stdio/in-memory transport wiring
- tool registration
- resource registration
- prompt registration when implemented
- MCP list-changed notification behavior
- capability sync wait behavior
- JSON wire shapes that are MCP-specific
- conversion between MCP requests and application service calls

The server should not own:

- task state transition rules
- project/namespace invariants
- session persistence semantics
- artifact lifecycle rules
- skill specification validation
- repository implementation details
- Codex invocation details
- KB HTTP API details

## Server Subpackages: Tools, Resources, Prompts

The user prefers:

```text
internal/server/tools/
internal/server/resources/
internal/server/prompts/
```

This is acceptable if each package can depend on application services without needing private `server` internals.

This shape supports one file per tool/resource:

```text
internal/server/tools/project_create.go
internal/server/tools/task_transition.go
internal/server/resources/scope_overview.go
internal/server/resources/skill.go
```

Benefit:

- The full context of a tool can live in one file.
- Large `tools.go` files are avoided.
- The MCP adapter surface becomes easier to scan.

Risk:

- If every handler needs `*server.Server` private fields, subpackages will force awkward exports.
- This can cause adapter packages to grow strange dependencies.

Guideline:

- Start with subpackages only if handlers can be written around explicit dependencies.
- Otherwise, keep a single `internal/server` package with files like:
  - `tool_project_create.go`
  - `tool_task_transition.go`
  - `resource_scope_overview.go`
  - `resource_skill.go`

Both are idiomatic. Avoid a giant `tools.go`.

## Refresh Mechanism

`refresh` is central and should remain the synchronization boundary.

Today, `refresh()`:

1. Accepts optional selection IDs.
2. Validates project, namespace, role, and board IDs.
3. Updates in-memory selection.
4. Reconciles dynamic tools and skill resources.
5. Starts capability sync tracking.
6. Waits for MCP client relist calls or returns fallback index.
7. Returns selection, overview, rendered skills, navigation, and capability sync status.

Desired architecture:

```text
internal/refresh       # application service
internal/server        # MCP adapter applying refresh result to MCP SDK
```

`refresh.Service` should decide:

- current session selection
- valid scope
- visible tools
- visible resources
- visible prompts
- visible skills
- overview/navigation payload
- capability plan

The MCP server should then:

- apply the desired dynamic tool/resource/prompt registrations
- emit list-changed notifications through the MCP SDK
- observe `tools/list`, `resources/list`, and later `prompts/list`
- wait for client relist or return fallback capability index
- serialize the result into MCP tool output

Open question:

- Should capability sync remain entirely in the MCP adapter, or should the refresh application service return enough intent for other future transports to implement their own sync mechanism?

Current answer:

- Keep MCP list-observation mechanics in `internal/server`.
- Keep transport-neutral refresh planning in `internal/refresh`.

## Sessions

Session state should become durable and decoupled from MCP stdio process lifetime.

Today:

- Selection is process-local.
- One stdio process equals one agent session.
- Restart loses selection.

Desired:

- A `session.Manager` owns selected namespace, project, role, board, active task, and possibly active artifact/context.
- Sessions can be resumed by ID.
- Refresh updates or clears session selection.
- Selection should not be hidden inside the MCP server.

Candidate session model:

```go
type Scope struct {
    NamespaceID *NamespaceID
    ProjectID   *ProjectID
}

type Selection struct {
    Scope      Scope
    RoleID     *RoleID
    BoardID    *BoardID
    TaskID     *TaskID
    ArtifactID *ArtifactID
}

type Session struct {
    ID        SessionID
    Selection Selection
    Metadata  map[string]string
}
```

Open questions:

- Should session ID be supplied only by server config/env, by `refresh(session_id)`, or both?
- Should a session be tied to one MCP client process or be portable across clients?
- Should session selection include an active task, active artifact, or both?

## Scope Model: Namespace, Project, Task

The relationship between namespaces, projects, and tasks is intentionally flexible.

Rules discussed:

- A namespace may contain projects, tasks, roles, boards, policies, skills, and knowledge.
- A project may exist independently or may belong to a namespace.
- A task may exist independently, may belong to a namespace, may belong to a project, or both.
- Everything should not be forced to belong to a project.
- Scope should be progressively narrowed.

Candidate:

```go
type Scope struct {
    NamespaceID *NamespaceID
    ProjectID   *ProjectID
}

type Project struct {
    ID          ProjectID
    NamespaceID *NamespaceID
    Name        string
}

type Task struct {
    ID          TaskID
    NamespaceID *NamespaceID
    ProjectID   *ProjectID
    Title       string
    State       TaskState
}
```

Potential invariant:

- If a task has both `NamespaceID` and `ProjectID`, and the project also has a namespace, they should match unless explicit cross-namespace linking is supported.

Open question:

- Should `Scope` be embedded in all scoped objects, or should each object expose optional IDs directly?

Leaning:

- Use a small shared `Scope` value object if it remains simple.
- Avoid over-abstracting if object-specific rules diverge.

## Task vs Issue vs Artifact

This was an important design discussion.

Initial question:

- Do we need both issues and tasks?
- If Workbench aims for full traceability, should issues and tasks be separate primitives?
- The agent needs a typical to-do list. Are those tasks?
- Are issues actually specifications?

Current conclusion:

- Do not make `Issue` a core primitive yet.
- A task is Workbench-owned executable work.
- An issue is usually an external tracker object, such as GitHub Issue, Jira ticket, or Linear issue.
- Workbench should model external issues as references or projections, not as the primary domain primitive.
- The missing primitive is not `Issue`; it is `Artifact`.

Core primitives should include:

```text
Task
Artifact
TraceLink
```

Task:

- Executable work.
- The agent's to-do item.
- Has a lifecycle centered on doing work.
- Can be project-scoped, namespace-scoped, both, or neither.

Artifact:

- Truth-bearing work product.
- Represents specs, RFCs, ADRs, statements of work, plans, research notes, acceptance contracts, and decisions.
- Has a lifecycle centered on review, acceptance, supersession, and archival.
- Can be project-scoped, namespace-scoped, both, or neither.

External issue:

- A tracker representation of a task or artifact.
- Should be modeled as an external reference through `tracker`.

Candidate external reference:

```go
type ExternalRef struct {
    System string // github, jira, linear
    Kind   string // issue, pr, discussion
    ID     string
    URI    string
}
```

Open question:

- Do we need a separate `tracker` package now, or should external references live only in `trace` until real issue tracker sync is implemented?

Leaning:

- Add `tracker` only when creating or syncing external tracker records.
- Until then, use trace links or external refs on tasks/artifacts.

## Task State Machine

The existing task state machine is:

```text
proposed
ready
in_progress
blocked
review
done
cancelled
```

Existing transitions:

```text
proposed -> ready
proposed -> cancelled
ready -> in_progress
ready -> blocked
ready -> cancelled
in_progress -> blocked
in_progress -> review
in_progress -> cancelled
blocked -> ready
blocked -> in_progress
blocked -> cancelled
review -> in_progress
review -> done
review -> cancelled
done -> terminal
cancelled -> terminal
```

Concern:

- The state machine may already be too complex.
- Some concepts may be orthogonal dimensions rather than states.

Simpler candidate:

```text
proposed -> ready -> in_progress -> review -> done
             |             |
             +-> blocked <-+
any nonterminal -> cancelled
```

Possible separate dimensions:

- priority
- blocked reason
- assignee/agent
- confidence
- evidence completeness
- review status
- external tracker status

Lesson learned:

- Do not overload task state with every workflow concern.
- Keep task state about execution progress.
- Use evidence, metadata, trace links, and artifact states for everything else.

Open questions:

- Is `review` a task state or an artifact/output state?
- Should `blocked` be a state or a condition with reasons?
- Should tasks support parent/child decomposition?
- Should tasks support dependencies directly, or only through trace links?
- Should tasks be durable before artifact support lands?

## Artifact Model

Artifacts are needed for full traceability.

Likely artifact kinds:

```text
problem_statement
research_note
knowledge_gap
rfc
spec
adr
statement_of_work
plan
acceptance_contract
decision_record
iteration_log
runbook
```

The existing docs already include examples:

- problem statements
- knowledge gaps
- program charter
- RFCs
- ADRs
- implementation plans
- iteration logs

Candidate artifact model:

```go
type ArtifactKind string

const (
    ArtifactProblemStatement ArtifactKind = "problem_statement"
    ArtifactResearchNote     ArtifactKind = "research_note"
    ArtifactRFC              ArtifactKind = "rfc"
    ArtifactSpec             ArtifactKind = "spec"
    ArtifactADR              ArtifactKind = "adr"
    ArtifactSOW              ArtifactKind = "statement_of_work"
    ArtifactPlan             ArtifactKind = "plan"
)

type Artifact struct {
    ID      ArtifactID
    Scope   Scope
    Kind    ArtifactKind
    Title   string
    State   ArtifactState
    BodyURI string
}
```

Candidate artifact state:

```text
draft -> in_review -> accepted -> active -> superseded
draft/in_review -> rejected
any stable state -> archived
```

Open questions:

- Should artifact body live in the database, filesystem, object store, or external docs system?
- Should artifacts be immutable/versioned after acceptance?
- Should ADRs and RFCs have specialized models or be generic artifact kinds?
- Should artifact creation be a tool, a resource template, or both?
- How should an agent assign an artifact to a project after creation?

Current leaning:

- Use a generic `Artifact` primitive first.
- Specialize only when a lifecycle or schema actually differs.

## Traceability

Full traceability should be represented through explicit links, not by collapsing every concept into one primitive.

Core primitive:

```go
type Ref struct {
    Type string // namespace, project, task, artifact, commit, pr, external_issue, skill, knowledge
    ID   string
    URI  string
}

type TraceLink struct {
    From Ref
    To   Ref
    Rel  Relation
}
```

Candidate relations:

```text
describes
refines
decomposes_to
implements
evidences
blocks
supersedes
mirrors
references
produces
authorizes
contradicts
resolves
```

Example traces:

```text
Problem Statement -> motivates -> RFC
RFC -> refines -> Spec
Statement of Work -> authorizes -> Spec
Spec -> decomposes_to -> Task
Task -> produces -> ADR
Task -> evidences -> Commit
Task -> mirrors -> GitHub Issue
PR -> implements -> Task
Artifact -> references -> KnowledgeItem
```

Open questions:

- Should trace links be first-class queryable resources immediately?
- Should traces be validated by allowed relation/type pairs?
- Should trace links support directionality and reverse lookup?
- Should trace links be append-only?

Leaning:

- Make trace links first-class early because they preserve flexibility across tasks, artifacts, external issues, commits, PRs, knowledge, and skills.

## Skills

Skills are instruction bundles that can be surfaced dynamically as MCP resources.

Current implementation:

- Embedded seed registry.
- Filesystem registry.
- Overlay registry where filesystem can override embedded skills.
- Skill resources are exposed under `skill://...`.

Desired model:

```text
internal/skills/
  bundle.go
  registry.go
  source.go
  spec.go
  manager.go

internal/skills/embedded/
  source.go

internal/skills/filesystem/
  source.go

internal/skills/s3/
  source.go     # future
```

The logical registry should compose multiple sources:

- embedded core skills
- local filesystem bundles
- future S3 bucket
- future Git registry
- future HTTP registry

The skill manager/registry should handle:

- discovery
- validation
- specification compliance
- enabled/disabled state
- opt-out rules
- precedence/overrides
- scope visibility
- rendering with project/session context
- resource metadata generation

The MCP server should only surface selected skills as resources.

It should not own skill validation or discovery.

Core embedded skills desired as proof of concept:

```text
gitops
go-coding-guidelines
nodejs-coding-guidelines
typescript-coding-guidelines
tdd
ddd
rest-principles
```

Important rule:

- Shipping a skill in code does not mean it is active all the time.
- Users should be able to opt out of core skills in the logical skill registry.
- Skills should be activated based on scope, role, project tech stack, explicit configuration, and possibly refresh input.

Open questions:

- What is the exact Agent Skills specification to enforce?
- What metadata is required for compliance?
- How should opt-out configuration be stored?
- Should skill enablement be namespace-level, project-level, session-level, or all three?
- Should skill versions appear in URIs?
- How should dynamic resource updates notify clients when filesystem/S3 skill bundles change?
- Should skills ever contribute MCP tools/prompts, or only resources?

## Knowledge and the KB Interaction

The name `kbapi` was rejected as too implementation-specific.

The name `knowledge_retriever` was also disliked as mechanical.

Current preferred fantasy-but-obvious name:

```text
internal/librarian/
internal/librarian/http/
```

Reasoning:

- A librarian searches, fetches, cites, and returns knowledge.
- It is understandable.
- It avoids confusion with Oracle DB.
- It keeps `knowledge` free for Workbench-owned knowledge records.

Conceptual split:

```text
knowledge.Repository
```

Stores Workbench-owned knowledge:

- feedback
- notes
- facts
- decisions
- observations

```text
librarian.Librarian
```

Queries an external knowledge base:

- content search
- knowledge query
- evidence retrieval
- citation retrieval

Current KB-backed `query` does two things:

- chooses retrieval primitives based on query text
- synthesizes a final answer or ad hoc skill resources

Desired design:

- Keep Workbench-owned knowledge in `internal/knowledge`.
- Put external KB interaction behind `internal/librarian`.
- Put HTTP implementation in `internal/librarian/http`.
- Keep final answer generation behind `agent.Runner` or a dedicated synthesis service using `agent`.

Open questions:

- Is `query` part of `knowledge`, `librarian`, or a separate application use case?
- Should Codex synthesis be invoked by `librarian`, by `knowledge.Manager`, or by a separate `query` use case?
- Should ad hoc skill resources produced by `query` be persisted, temporary, or session-scoped?
- Should the librarian return raw evidence only, or also synthesized summaries?

Leaning:

- `librarian` retrieves.
- `agent` synthesizes.
- `knowledge` stores Workbench-owned facts/feedback.
- `query` can be an application use case that orchestrates all three.

## Agent Package

The name `agentrunner` was rejected.

Preferred:

```text
internal/agent/
internal/agent/codex/
```

Core abstraction:

```go
type Runner interface {
    Run(ctx context.Context, req RunRequest) (RunResult, error)
}
```

The Codex implementation should live at the edge:

```text
internal/agent/codex/
```

It may use:

- `os/exec`
- Codex CLI arguments
- timeouts
- JSON parsing of agent output

Core packages should only depend on `agent.Runner` where they consume it.

Open questions:

- Should `agent.Runner` be generic for all headless agent work, or should there be more specific interfaces for query synthesis, semantic routing, background jobs, etc.?
- Should prompts passed to Codex be versioned artifacts?
- Should all agent outputs be validated before mutating state?

Strong rule:

- Headless agents can propose or synthesize.
- Deterministic Go code owns validation and state mutation.

## Tracker and External Issues

An issue tracker may be needed later, but `Issue` should not become a core primitive prematurely.

Candidate package:

```text
internal/tracker/
internal/tracker/github/
```

The tracker package represents external systems:

- GitHub Issues
- GitHub PRs
- Jira
- Linear

Potential interface:

```go
type Tracker interface {
    Create(ctx context.Context, draft Draft) (ExternalRef, error)
    Get(ctx context.Context, ref ExternalRef) (Record, error)
    Sync(ctx context.Context, ref ExternalRef) (Record, error)
}
```

Open questions:

- Does Workbench need to create external issues, or only link to them?
- Should a GitHub issue mirror a task, an artifact, or both?
- Should issue comments become knowledge items?
- Should PRs be trace-linked to tasks/artifacts directly?

Leaning:

- Delay `tracker` until there is a concrete GitHub/Jira/Linear workflow.
- In the meantime, model external references and trace links.

## Prompts

Prompts are planned but not meaningfully implemented yet.

They should be treated as first-class capabilities alongside:

- tools
- resources
- skills

Open questions:

- Should prompts be static, dynamic, or both?
- Should prompts be generated from artifacts, roles, skills, and project context?
- Should prompts be registered as MCP prompts or served as resources first?
- Should prompt templates be artifacts?
- Should role system prompts and project system prompts become prompt templates?

Current leaning:

- Add `internal/server/prompts` only for MCP prompt adapter code.
- Keep prompt template/domain logic in an application/domain package if it becomes more than wire registration.

## Namespace and Project Manager

The user wants a project manager package that holds everything from project to namespaces.

Candidate:

```text
internal/project/
  project.go
  namespace.go
  role.go
  board.go
  manager.go
  repository.go
```

This is reasonable because in the current product:

- namespace management
- project management
- role management
- board management

are all part of scope/project organization.

Potential risk:

- The package may grow too broad if namespace becomes its own large domain with policies, skills, integrations, and hierarchy rules.

Open question:

- Should namespace remain inside `project`, or eventually become `internal/namespace`?

Leaning:

- Start with `internal/project` for project/namespace/role/board.
- Extract `internal/namespace` only if namespace-specific behavior becomes large enough.

## Naming Lessons

Agreed naming direction:

- Use `internal/server`, not necessarily `internal/mcpserver`.
- Use `internal/skills`, not `skillsregistry`.
- Use `internal/agent`, not `agentrunner`.
- Use `internal/librarian` for external KB interaction unless a better name emerges.
- Use `internal/repo/<backend>`, not generic `storage`.
- Avoid a central `ports` folder.
- Avoid an umbrella `workbench` folder.
- Avoid underscores in Go package names.
- Use plural package names only when the package naturally represents a set or collection, such as `skills`, `tools`, or `resources`.
- Use singular package names for domain concepts, such as `project`, `task`, `session`, `artifact`, `trace`, `knowledge`, `refresh`, `agent`.

## Suggested Migration Strategy

Do not rewrite everything at once.

Possible sequence:

1. Create domain/application packages without changing behavior:
   - `project`
   - `task`
   - `session`
   - `skills`
   - `knowledge`
   - `refresh`

2. Move pure domain rules first:
   - task states and transitions
   - project/namespace/role/board types
   - selection/scope types
   - skill bundle types

3. Move repository interfaces next:
   - define consumer-owned interfaces in domain/application packages
   - keep existing file/memory store adapters working

4. Extract skill registry:
   - core `skills` package
   - embedded adapter
   - filesystem adapter
   - overlay/logical registry

5. Extract `refresh.Service`:
   - it returns a transport-neutral capability/overview plan
   - MCP server applies the plan

6. Rename `internal/mcpserver` to `internal/server` only after core extractions reduce coupling.

7. Add artifact and trace packages:
   - generic artifact model
   - trace links between artifacts, tasks, projects, external refs, knowledge, commits, PRs, and skills

8. Add durable sessions:
   - session manager
   - repository
   - refresh integration

9. Replace in-memory tasks/knowledge/roles/boards with repositories as needed.

10. Add Postgres repository implementations behind existing interfaces.

11. Add richer skill specification validation and opt-out registry config.

12. Add tracker integration only when a real external issue workflow is ready.

## Acceptance Criteria for Restructure

The restructure should preserve current behavior:

- `make build` passes.
- `make test` passes.
- Current MCP refresh acceptance behavior remains intact.
- `refresh()` still returns overview and rendered skills.
- Dynamic tools/resources still update by scope.
- `workbench:///scope/overview` still mirrors the inline refresh overview.
- `workbench:///scope/capabilities` remains absent.
- Skill resources continue to use `skill://...` URIs.
- `query` still works in fallback local mode and KB-backed mode.
- Existing tests are migrated, not deleted.

New architecture acceptance:

- Domain packages do not import the MCP SDK.
- Domain packages do not import Postgres/HTTP/Codex adapter packages.
- Interfaces are defined by consumers.
- Repositories are grouped under `internal/repo/<backend>`.
- MCP-specific code lives under `internal/server`.
- Codex CLI invocation lives under `internal/agent/codex`.
- External KB HTTP calls live under `internal/librarian/http`.
- Skill discovery/validation is not owned by the MCP server.
- Task transition rules are not owned by the MCP server.
- Refresh planning is testable without MCP.

## Open Design Questions

Repository and persistence:

- Should file repositories remain long-term or only support local development?
- Should Postgres become the default backend once durable sessions/tasks/artifacts exist?
- How should cross-repository transactions be handled?

Server adapter:

- Should `tools`, `resources`, and `prompts` be subpackages or same-package split files?
- What dependency object should each tool/resource handler receive?
- How should MCP prompt support be introduced?

Refresh:

- What is the exact transport-neutral refresh result shape?
- Should prompt relisting be part of capability sync from the start?
- Should refresh include active artifact selection?

Sessions:

- How is session ID provided?
- Is session recovery explicit or automatic?
- Can multiple clients share one session?

Scope:

- Should `Scope` be a shared value object?
- Should namespace/project consistency be enforced when both IDs are present?
- Can tasks or artifacts intentionally cross namespace/project boundaries?

Tasks:

- Should `blocked` be a state or a condition?
- Is `review` a task state or artifact/output state?
- Do tasks need parent/child decomposition?
- Do task dependencies live in task fields or trace links?

Artifacts:

- Where does artifact body live?
- Should artifacts be versioned?
- Should ADRs/RFCs/specs/SOWs be specialized types or generic artifact kinds?
- What tools should create/update/assign artifacts?

Traceability:

- Which relation names are canonical?
- Should trace links be append-only?
- Should allowed relation/type pairs be validated?

Skills:

- What skill specification must be enforced?
- How is opt-out stored and scoped?
- Should skill versions appear in URIs?
- How should future S3/Git/HTTP skill sources be watched or synced?
- Can skills contribute tools/prompts, or only resources?

Knowledge and librarian:

- Is `librarian` the right name?
- Does `query` live as its own application use case?
- Are ad hoc skill resources temporary, persisted, or session-scoped?
- Does the external KB return raw evidence only or synthesized knowledge too?

Agent:

- Is one `agent.Runner` abstraction enough?
- Should Codex prompts be artifacts?
- What validation boundary is required before agent output mutates state?

Tracker:

- When should `tracker` be introduced?
- Does an external issue mirror a task, an artifact, or either?
- Should issue comments become knowledge items?

## Lessons Learned

- The current project already has the right product shape: an adaptive MCP capability kernel centered on `refresh()`.
- The current package structure is useful for a greenfield slice but mixes too many responsibilities for the next phase.
- Go architecture should be package-driven, not layer-folder-driven.
- DDD vocabulary is useful only where it reflects real invariants and lifecycles.
- Hexagonal architecture is valuable here because Workbench has many edge dependencies: MCP, filesystem, Postgres, external KB, Codex CLI, future S3, and future issue trackers.
- `refresh` is not just a setter. It is the synchronization boundary between session scope and dynamic MCP capabilities.
- `workbench:///scope/overview` should remain read-only and should not duplicate full MCP capability manifests.
- Skills should be logical bundles selected by a registry, not hardcoded MCP resources.
- Embedded skills are a default harness, not mandatory always-on behavior.
- Tasks and artifacts should stay separate because they have different lifecycles.
- External issues should not become a core primitive until tracker integration needs it.
- Trace links are likely the key to full traceability without overloading tasks, artifacts, or issues.
- Agent/Codex execution should remain behind an adapter; deterministic Go code must own validation and mutation.
- Avoid a central `ports` folder. Interfaces should be small and close to the consumer.
- Avoid generic `storage`; repository adapters grouped by backend are more useful for transactions and operations.
- Avoid premature package splitting when same-package split files are enough.

## Short Handoff Summary

Build toward this shape:

```text
internal/server          # MCP inbound adapter
internal/refresh         # transport-neutral refresh application service
internal/project         # namespace/project/role/board management
internal/task            # executable work and state machine
internal/artifact        # specs, RFCs, ADRs, SOWs, plans
internal/trace           # trace links across primitives and external refs
internal/session         # durable session selection
internal/skills          # logical skill registry and spec validation
internal/knowledge       # Workbench-owned knowledge records
internal/librarian       # external KB interaction
internal/agent           # headless agent abstraction
internal/repo/file       # file repository adapters
internal/repo/postgres   # Postgres repository adapters
internal/tracker         # future issue tracker integration
```

Keep MCP at the edge. Keep domain rules inward. Keep Go package design simple until complexity proves the next split.
