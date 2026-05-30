---
id: "knowledge-mgmt-research-note"
type: "research_note"
title: "Knowledge Management Research Note"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Knowledge Management Research Note

## Question

How should Workbench define a knowledge management epic that supports
source-governed retrieval, indexing, freshness, ranking, citations, feedback,
and boundaries with explicit memory and durable artifacts?

## Sources

Local current project sources:

- [Workbench README](../../README.md): current runtime scope, MCP surface, and
  local stdio assumptions.
- [Epic Branch Workflow](../how-to/epic-branch-workflow.md): packet shape, RFC
  drill-hub role, and human nudge index requirement.
- [Artifact Conventions](../reference/artifact-conventions.md): frontmatter,
  supported artifact types, required sections, and validation rules.

Current web and GitHub sources:

- [MCP Resources](https://modelcontextprotocol.io/specification/2025-11-25/server/resources):
  resource user model, list/read/template operations, list-changed
  notifications, and subscriptions.
- [MCP Tools](https://modelcontextprotocol.io/specification/2025-11-25/server/tools):
  tool user model, model-controlled invocation, and human-in-the-loop security
  guidance.
- [MCP Elicitation](https://modelcontextprotocol.io/specification/2025-11-25/client/elicitation):
  structured human input model and constraints around sensitive information.
- [MCP Security Best Practices](https://modelcontextprotocol.io/docs/tutorials/security/security_best_practices):
  consent, trust boundary, and proxy-risk guidance.
- [GitHub Docs API](https://docs.github.com/get-started/using-github-docs/github-docs-api):
  authoritative docs article, metadata, pagelist, search, and `llms.txt`
  access patterns.
- [manuelibar/ripple](https://github.com/manuelibar/ripple): public design
  context for Workbench as part of a larger source-of-truth, artifact, and
  memory architecture.
- [arabold/docs-mcp-server](https://github.com/arabold/docs-mcp-server):
  ecosystem example of local, current, multi-source documentation indexing for
  MCP clients.
- [brave/brave-search-mcp-server](https://github.com/brave/brave-search-mcp-server):
  ecosystem example of search tools exposing freshness filters, source
  metadata, and LLM-oriented context.

## Findings

Workbench should treat knowledge as evidence, not as authored truth. The
current kernel already owns context and durable artifacts, while this epic owns
retrieval over source material. That split supports a narrow contract: register
sources, snapshot and index them, search them, return cited evidence, and record
retrieval feedback.

MCP resources and tools imply two complementary surfaces. Source lists,
retrieval packets, and indexed result resources can be application-driven
resources. Search, refresh, and feedback recording are better modeled as tools
because they perform bounded actions. Tool invocations should remain visible to
the user for trust and safety.

MCP elicitation is relevant for human nudges and approvals, but it should not be
used to request secrets in form mode. For knowledge management, elicitation is a
fit for source-scope decisions, ranking tradeoffs, artifact indexing approval,
and feedback prompts.

Freshness needs first-class data. GitHub Docs exposes APIs for page lists,
article bodies, metadata, search, and `llms.txt`, showing that good connectors
can pull structured source inventories and exact Markdown. Ecosystem docs/search
MCP servers also emphasize current official sources, local privacy, freshness
filters, source metadata, and LLM-oriented context. Workbench should capture
those properties without copying any one implementation.

The central boundary is with memory. Feedback that says "this result was
useful" is a retrieval quality signal. It is not the same thing as "remember
this preference about me." Knowledge can later consume explicit memory as a
query hint only through a memory-defined interface.

Durable artifacts have a similar boundary. The knowledge layer may index
artifact contents for discovery, but citations must retain artifact identity and
section. Editing, validation, lineage, and review remain artifact-system work.

## Implications

- The source registry is the first implementation primitive. Without it,
  ranking, citations, freshness, and feedback have no trust anchor.
- Retrieval packets should be structured and inspectable. A plain synthesized
  answer is not enough for this epic.
- The MVP can start local with lexical retrieval and metadata ranking, then add
  embeddings or reranking behind the same source and citation contract.
- Feedback storage must be typed separately from memory storage.
- The RFC should remain the drill hub for defaults and human nudges because
  source scope, ranking, freshness, artifact indexing, and conflict authority
  are real product decisions.

## Source References

- [Knowledge Management RFC](knowledge-mgmt-rfc.md)
- [Knowledge Management Concept Map](knowledge-mgmt-concept-map.md)

## Open Questions

See the RFC human-in-the-loop index for the open decisions and defaults derived
from this research.
