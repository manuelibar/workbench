---
id: "role-mgmt-research-note"
type: "research_note"
title: "Role Management Research Note"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Role Management Research Note

## Question

How should Workbench define role management using current repo contracts and
current MCP/Go SDK guidance, while keeping role definitions, role selection,
instruction shaping, capability mapping, custom role storage, and precedence
boundaries deterministic?

## Sources

- `README.md`: Defines Workbench as a local stdio MCP context and artifact
  kernel, lists current tools/resources, and defers roles, memory, sessions,
  and skills to epic branches.
- `docs/how-to/epic-branch-workflow.md`: Defines the docs-only packet workflow,
  RFC drill hub, current-doc research base, and human-in-the-loop index.
- `docs/reference/artifact-conventions.md`: Defines frontmatter, artifact
  types, required sections, and RFC/action-artifact linkage.
- `docs/reference/context-contract.md`: Defines `context` tri-state patch
  behavior, context results, capability indexes, and MCP list sync categories.
- `docs/explanation/context-window-thesis.md`: States that Workbench treats the
  model context window as scarce and defers roles until stronger contracts
  exist.
- `docs/explanation/progressive-disclosure.md`: Defines capability visibility
  as layered and synchronized through MCP list-changed behavior.
- [MCP specification overview, version 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25):
  Current authoritative protocol overview, including hosts, clients, servers,
  capability negotiation, and trust principles.
- [MCP architecture](https://modelcontextprotocol.io/specification/2025-11-25/architecture):
  Describes host/client/server responsibilities and security boundaries.
- [MCP server feature overview](https://modelcontextprotocol.io/specification/2025-11-25/server/index):
  Defines prompts as user-controlled, resources as application-controlled, and
  tools as model-controlled.
- [MCP tools](https://modelcontextprotocol.io/specification/2025-11-25/server/tools):
  Defines tool discovery, list-changed notifications, schemas, output schemas,
  untrusted annotations, and security considerations.
- [MCP resources](https://modelcontextprotocol.io/specification/2025-11-25/server/resources):
  Defines resource URIs, content, annotations, subscriptions, and resource
  security considerations.
- [MCP prompts](https://modelcontextprotocol.io/specification/2025-11-25/server/prompts):
  Defines prompt templates, prompt messages, embedded resources, validation,
  and prompt security.
- [MCP roots](https://modelcontextprotocol.io/specification/2025-11-25/client/roots):
  Defines filesystem root boundaries that clients expose to servers.
- [modelcontextprotocol/modelcontextprotocol](https://github.com/modelcontextprotocol/modelcontextprotocol):
  Official MCP specification, schema, and documentation repository.
- [modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk):
  Official Go SDK used by this repo; its docs state that features such as
  tools, prompts, and resources infer capabilities and list-changed behavior.

## Findings

MCP does not supply a role primitive. Its current server primitives are
resources, prompts, and tools, with a clear control hierarchy: prompts are
selected by users, resources are managed by applications, and tools are invoked
by models. Workbench roles should therefore be local policy that compiles into
context rendering, prompt/resource/tool exposure, and capability diagnostics.

Workbench already has the right foundation for deterministic role selection:
tri-state context patching, an exact `workbench:///context` resource, a
capability catalog, deterministic capability planning, list-changed sync, and a
fallback capability index. Adding `role_id` later should follow the same
contract shape as `focus` and `artifact_id`.

Security guidance points toward conservative semantics. MCP requires careful
handling of powerful data and code execution paths, treats tool behavior
annotations as untrusted unless from trusted servers, requires prompt input and
output validation, and defines roots as filesystem boundaries. A Workbench role
must not convert untrusted text into higher authority or expand capabilities
beyond the base context and harness.

The official Go SDK supports the current MCP feature set and automatically
infers capabilities when server features are added. That makes role-driven
capability mapping feasible, but it also means Workbench must test list-changed
notifications whenever a role changes visible tools, resources, templates, or
prompts.

## Implications

The role-management contract should:

- Represent selected role state explicitly in Workbench context.
- Keep role definitions inspectable and file-backed.
- Treat custom role content as untrusted until validated.
- Make role capability policy restrictive by default.
- Define precedence before implementation, especially around skills, memory,
  sessions, and harness configuration.
- Include diagnostics so users can explain why a role changed context or
  capability visibility.

## Open Questions

The packet-level human nudges are indexed in
[Role Management RFC](role-mgmt-rfc.md#human-in-the-loop-index).
