---
id: "artifact-b2-research-note"
type: "research_note"
title: "Artifact B2 Workflow Research Note"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Artifact B2 Workflow Research Note

## Question

What current Workbench and MCP constraints should shape Artifact B2's advanced
artifact lifecycle, relationship, elicitation, sign-off, and governance
contract?

## Sources

Local Workbench sources read for this packet:

- [README.md](../../README.md): current foundation scope, MCP surface, and
  file-backed artifact directory.
- [Epic Branch Workflow](../how-to/epic-branch-workflow.md): packet bootstrap
  rules, RFC hub expectations, action artifact linking, and human nudge index.
- [Artifact Conventions](../reference/artifact-conventions.md): required
  frontmatter, supported artifact types, validation behavior, and B2 deferrals.
- `internal/mcpserver/contracts.go`: current type registry and required
  sections used by runtime validation.

Current authoritative web sources checked on 2026-05-30:

- [MCP Specification, 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25):
  latest MCP overview, capability model, and trust-and-safety principles.
- [MCP Resources, 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25/server/resources):
  resource identity, listing, reading, subscriptions, list-changed behavior,
  URI guidance, and resource security.
- [MCP Tools, 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25/server/tools):
  tool discovery, invocation, list-changed behavior, schema expectations,
  resource links, structured content, and security considerations.
- [MCP Elicitation, 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25/client/elicitation):
  server-initiated user requests, form and URL modes, response actions,
  completion notifications, error handling, and sensitive-data boundaries.

## Findings

Workbench findings:

- The current kernel is intentionally narrow: `context`, artifact CRUD-like
  reads and section updates, validation, and artifact resources.
- Current artifacts are Markdown files with required frontmatter and required
  sections determined by type.
- RFCs are the living drill hub and every packet-level human nudge must be
  indexed under the RFC's `## Open Questions` section.
- The local artifact conventions explicitly defer advanced lineage, delete,
  archive, supersession, elicitation, sign-off, and review workflows to
  Artifact B2.

MCP findings:

- The latest MCP specification is version `2025-11-25`; it frames MCP around
  JSON-RPC, stateful connections, capability negotiation, resources, tools, and
  client features such as elicitation.
- MCP resources are URI-identified context objects. Servers can expose list,
  read, subscribe, and list-changed behavior, but the protocol does not mandate
  one user interaction model for selecting resources.
- MCP tools are model-invoked functions with JSON Schema inputs and optional
  structured outputs. The tools specification recommends human confirmation for
  sensitive operations and treats tool annotations as untrusted unless the
  server is trusted.
- MCP elicitation lets a server request additional information through the
  client. Form mode collects structured non-sensitive data, while URL mode is
  required for sensitive interactions such as credentials. Clients need clear
  server identity, decline/cancel options, and user approval controls.
- The MCP security guidance emphasizes user consent, user control, privacy,
  access controls, auditability, and careful handling of tool execution.

## Implications

Artifact B2 should use Markdown as the durable source and define additive
metadata before runtime mutation tools. Lifecycle transitions that alter
currentness or visibility should be explicit, reasoned, and auditable. Hard
delete should not be the default behavior.

Relationships should be represented with stable artifact IDs and direction, so
they can later map cleanly to resources or tool results. Human nudges should be
represented in the RFC index first, then mapped to MCP elicitation only after
client support and security boundaries are clear. Sign-off should be modeled as
governance metadata tied to a specific transition or artifact version, not as
loose prose.

## Source References

- [Artifact B2 RFC](artifact-b2-rfc.md)
- [Artifact B2 Concept Map](artifact-b2-concept-map.md)
- [Artifact B2 Initial Implementation Plan](artifact-b2-initial-implementation-plan.md)

## Open Questions

The research leaves five packet-level nudges, all indexed in the
[Artifact B2 RFC](artifact-b2-rfc.md#human-in-the-loop-index): `B2-D1`,
`B2-D2`, `B2-Q1`, `B2-T1`, and `B2-C1`.
