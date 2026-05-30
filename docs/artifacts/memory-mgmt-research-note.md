---
id: "memory-mgmt-research-note"
type: "research_note"
title: "Memory Management Research Note"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Memory Management Research Note

## Question

What should Workbench memory management learn from the current repository
contracts, MCP protocol primitives, and current user-facing memory controls so
that durable memory remains explicit, inspectable, scoped, and subordinate to
artifacts, knowledge, session context, and current user intent?

## Sources

Repository-local sources read in this worktree:

- [README.md](../../README.md): defines Workbench as a local stdio MCP context
  and artifact kernel, and defers memory, knowledge, sessions, and other
  systems to epic branches.
- [docs/how-to/epic-branch-workflow.md](../how-to/epic-branch-workflow.md):
  defines the self-contained epic packet workflow and RFC human nudge index.
- [docs/reference/artifact-conventions.md](../reference/artifact-conventions.md):
  defines typed artifact frontmatter, required sections, and RFC drill-hub
  expectations.
- [docs/reference/context-contract.md](../reference/context-contract.md):
  defines the current live context surface as `focus` and `artifact_id`.

External authoritative sources checked on 2026-05-30:

- [MCP Specification 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25):
  current stable MCP protocol version and security principles.
- [MCP server overview](https://modelcontextprotocol.io/specification/2025-11-25/server/index):
  distinguishes prompts, resources, and tools by control model.
- [MCP tools](https://modelcontextprotocol.io/specification/2025-11-25/server/tools):
  describes model-controlled tools, human confirmation expectations, and tool
  security considerations.
- [MCP resources](https://modelcontextprotocol.io/specification/2025-11-25/server/resources):
  describes resource identity, resource templates, subscriptions, and content
  annotations.
- [MCP roots](https://modelcontextprotocol.io/specification/2025-11-25/client/roots):
  defines client-exposed filesystem boundaries and consent expectations.
- [MCP elicitation](https://modelcontextprotocol.io/specification/2025-11-25/client/elicitation):
  defines accept, decline, and cancel flows for user input, including sensitive
  data restrictions.
- [modelcontextprotocol/modelcontextprotocol](https://github.com/modelcontextprotocol/modelcontextprotocol):
  official GitHub repository for MCP specification, schema, and documentation.
- [OpenAI Memory FAQ](https://help.openai.com/en/articles/8590148-memory-in-chatgpt):
  current user-facing saved-memory and reference-chat-history controls.
- [OpenAI Model Spec](https://github.com/openai/model_spec/blob/main/model_spec.md):
  current public model behavior specification, including instruction authority
  levels and treatment of lower-authority or untrusted content.

## Findings

Workbench already draws a hard boundary between live context and durable
artifacts. `context` mutates only focus and selected artifact state, while
artifacts are typed Markdown documents. Memory should follow the same design
principle: expose deliberate operations and durable records instead of hidden
always-in-context text.

MCP's current server primitive model is useful for memory design. Resources are
application-controlled context, while tools are model-controlled actions. That
suggests memory records should be inspectable through resources or resource
templates, while remember, correct, and forget should be tools with audit and
confirmation behavior. MCP's security guidance emphasizes user consent and
control for data access and operations, which aligns with explicit memory
writes and careful deletion flows.

MCP elicitation is relevant to future memory UX. Its accept, decline, and
cancel result states map cleanly to memory confirmation flows, and its
sensitive-data rules reinforce that secrets and credentials should not be
collected through ordinary in-band forms or durable memory records.

MCP roots provide a useful analogy for memory scope. Roots tell servers where
they can operate in the filesystem; memory scopes should tell agents where a
record can influence behavior. A project-scoped memory should not silently
apply globally, and a session-scoped note should not become durable unless the
user promotes it.

OpenAI's memory documentation separates saved memories from reference chat
history. Saved memories are user-manageable and stored separately from chat
history, while chat-history-derived context can change over time. The same
distinction supports Workbench keeping durable memory separate from session
state, with explicit user controls to inspect and delete saved records.

OpenAI's Model Spec provides a useful precedence lesson. It assigns authority
levels to instructions and treats tool outputs and lower-authority content as
not automatically authoritative. Workbench memory should be designed as
advisory context with provenance, not as instruction authority that can
override current user requests or higher-level policy.

Knowledge and memory need a strict boundary. Knowledge should be sourced,
refreshable, and citation-oriented. Memory should be user-controlled and scoped
to the user, workspace, project, artifact, or session. A research note can
inform memory, but it should not become memory unless the user explicitly asks
Workbench to remember a scoped implication.

## Implications

- The first memory API should require explicit metadata: source, source
  reference, confidence, scope, sensitivity, state, and timestamps.
- `memory.recall` should return structured candidates with provenance, not
  silently mutate `workbench:///context`.
- `memory.remember`, `memory.correct`, and `memory.forget` should be treated as
  auditable tools, with stricter confirmation for inferred, broad, sensitive,
  or destructive operations.
- The RFC should define memory as subordinate to current user instructions and
  selected artifacts. It should also define how to expose conflicts with
  knowledge instead of hiding them.
- The implementation plan should start with local file-backed storage to match
  the current kernel and defer retrieval engine selection.
- The test strategy must include conflict, correction, forgetting, sensitivity,
  and scope-boundary fixtures from the beginning.

## Source References

- [memory-mgmt-rfc.md](memory-mgmt-rfc.md)
- [memory-mgmt-concept-map.md](memory-mgmt-concept-map.md)
- [memory-mgmt-initial-implementation-plan.md](memory-mgmt-initial-implementation-plan.md)
- [memory-mgmt-test-strategy.md](memory-mgmt-test-strategy.md)

## Open Questions

No research-specific nudges are open. Packet-level human nudges are indexed in
[memory-mgmt-rfc.md](memory-mgmt-rfc.md).
