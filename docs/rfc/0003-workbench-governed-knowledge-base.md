# RFC 0003: Workbench-Governed Knowledge Base Integration

Status: Draft

Date: 2026-05-30

## Summary

Workbench is the MCP entry point for agents. Cross-system knowledge features
should therefore be specified first in Workbench, even when implementation work
also lands in the Knowledge Base service.

This RFC reframes the contribution-governed knowledge base design as an
ecosystem contract:

- Workbench owns the agent-facing MCP tools, resources, scope selection,
  traceability, GitHub issue linkage, and orchestration across services.
- KB owns source storage, contribution evaluation, evidence-backed knowledge,
  retrieval, and implementation details such as Neo4j, chunking, indexes, and
  guardian review records.
- Service repos may keep their own RFCs, specs, and implementation notes, but
  cross-system behavior has a master RFC in Workbench.

The desired product surface for agents is simple:

- Read side: ask Workbench a question and receive an evidence-backed answer.
- Write side: submit candidate knowledge through Workbench and let KB decide
  whether it becomes accepted knowledge.
- Navigation side: use related topics, evidence, tasks, and artifacts without
  exposing graph storage internals.

## Context

The original contribution-governed KB RFC was drafted inside the KB repo. That
was reasonable while the question looked service-local. The feature is now
clearly cross-system:

- agents enter the ecosystem through Workbench MCP,
- Workbench selects namespace, project, role, board, and later task/artifact
  context,
- Workbench already has an `ask` tool and KB-backed retrieval path,
- Workbench will likely create or link GitHub issues for implementation slices,
- future services such as backlog and documentation will also be reached
  through Workbench.

The current Workbench implementation still calls two KB retrieval primitives:

- `POST /content/search`
- `POST /knowledge/query`

The KB draft proposes a smaller service API:

- `POST /contributions`
- `POST /query`

This RFC does not discard the KB draft. It promotes the cross-system contract
into Workbench and leaves KB-specific implementation details to the KB repo.

## Documentation Ownership Rule

Use this rule for future docs:

- Workbench `docs/rfc`: master RFCs for cross-system behavior, MCP-facing
  contracts, ecosystem workflow, traceability, and integration boundaries.
- Workbench `docs/plans`: delivery slices and GitHub issue breakdowns for
  Workbench-owned integration work.
- Service repos: service-local RFCs, specs, implementation notes, schema/API
  details, and service acceptance criteria.
- Service-local docs should point back to the Workbench RFC when the behavior
  affects agent-facing workflow or multiple services.

Example:

- This RFC owns the Workbench/KB integration contract.
- The KB repo can keep an implementation RFC for the `/contributions` and
  `/query` HTTP API, graph shape, evidence locators, source storage, and
  guardian internals.

## Goals

- Make Workbench the durable entry point for KB-backed knowledge operations.
- Keep the MCP surface small enough for agents to use without learning KB
  internals.
- Preserve KB as an independent service with a narrow HTTP contract.
- Route project, namespace, role, and future task/artifact scope from Workbench
  into KB queries and contributions.
- Return evidence-backed answers without exposing Neo4j, chunks, statements,
  graph edges, graph IDs, QMD, or Workbench-internal resource IDs from KB.
- Treat corrections as knowledge contributions rather than graph patches.
- Create traceable GitHub issues from this RFC for implementation slices.

## Non-Goals

- Do not make KB an MCP server in the MVP.
- Do not expose Neo4j, Cypher, graph nodes, internal labels, chunks, or raw
  retrieval scores to agents.
- Do not make Workbench store the canonical KB graph.
- Do not require agents to choose vector search, full-text search, graph search,
  or hybrid retrieval.
- Do not finalize every future artifact, task, issue, and trace-link model in
  this RFC.
- Do not implement the full human moderation product in the MVP.

## System Boundary

### Workbench Owns

Workbench owns the agent-facing integration:

- MCP tools such as `ask`, `feedback`, and future knowledge contribution tools.
- MCP resources such as `workbench:///knowledge`, scope overview, tasks,
  artifacts, GitHub config, and future evidence/source views.
- Scope selection and normalization across namespace, project, role, board,
  task, and artifact.
- Trace links among RFCs, specs, tasks, GitHub issues, PRs, commits, KB
  evidence, and service-local artifacts.
- GitHub issue and PR integration when tracker support lands.
- Routing between KB, backlog, docs, Ripple, and future services.

### KB Owns

KB owns the knowledge service implementation:

- raw source storage,
- contribution records,
- source parsing and chunking,
- extracted internal statements or knowledge units,
- evidence locators,
- accepted/rejected/clarification decisions,
- correction and supersession history,
- graph, full-text, vector, and hybrid retrieval internals,
- guardian evaluation records.

KB should not know about MCP resources, Workbench resource URIs, GitHub issue
IDs, or Workbench session mechanics.

## Agent-Facing MCP Contract

### Read Path

Workbench keeps `ask` as the agent-facing read tool.

Current shape:

```json
{
  "criteria": "how does the order system work?",
  "scope": {
    "namespace_id": "acme",
    "project_id": "platform",
    "role": "coder"
  },
  "limit": 20
}
```

Target shape:

```json
{
  "criteria": "how does the order system work?",
  "scope": {
    "namespace_id": "acme",
    "project_id": "platform",
    "role": "coder"
  }
}
```

Workbench should translate this into KB's target query contract and return:

```json
{
  "criteria": "how does the order system work?",
  "answer": "Orders are created after checkout completes and remain pending until payment authorization succeeds.",
  "evidence": [
    {
      "id": "kb:evidence:01J...",
      "source_uri": "kb:source:01J...",
      "source_title": "Order System Notes",
      "source_media_type": "text/markdown",
      "locator": {
        "line_start": 12,
        "line_end": 20
      }
    }
  ],
  "related_topics": [
    {
      "title": "Payment Authorization",
      "relationship": "needed_to_explain_order_progression",
      "query": "payment authorization in the order system"
    }
  ]
}
```

Workbench may also publish temporary MCP resources for evidence or synthesized
materials when that improves agent navigation. Those resources are Workbench
views. They are not KB's canonical evidence object.

### Write Path

Workbench should expose a contribution-oriented write path once KB implements
the service contract.

Candidate MCP tool:

```text
knowledge.contribute
```

Candidate input:

```json
{
  "statement": "Orders are created after checkout completes.",
  "focus": "Correct order lifecycle knowledge if supported.",
  "source_uri": "optional Workbench artifact or uploaded source reference",
  "scope": {
    "namespace_id": "acme",
    "project_id": "platform",
    "role": "coder"
  }
}
```

Workbench maps this to KB's contribution endpoint. KB decides whether the
contribution is accepted, rejected, or needs clarification.

Candidate output:

```json
{
  "contribution": {
    "uri": "kb:contribution:01J...",
    "created_at": "2026-05-30T12:00:00Z"
  },
  "decision": {
    "outcome": "needs_clarification",
    "summary": "The correction conflicts with accepted project knowledge but does not cite a stronger source.",
    "questions": [
      "Which source should be treated as authoritative for the checkout order lifecycle?"
    ],
    "reasons": [
      "conflict_without_stronger_evidence"
    ]
  }
}
```

The existing `feedback` tool remains Workbench-owned. Feedback can become a KB
contribution only after a deterministic Workbench policy decides that the
feedback is knowledge-bearing and should be submitted.

## KB Service Contract

The target KB service API remains HTTP-first:

```http
POST /contributions
Content-Type: multipart/form-data
```

```text
file=<optional binary source>
statement=<optional free text>
focus=<optional free text>
```

```http
POST /query
Content-Type: application/json
```

```json
{
  "query": "what is Diataxis?"
}
```

The KB response should include:

- `answer`,
- `evidence`,
- `related_topics`.

KB should return `204 No Content` when it cannot produce a grounded answer.
Workbench may translate that into a structured MCP tool result if agents need a
machine-readable "not grounded" response.

Later KB may add:

```http
GET /sources/{source_uri}/content
```

for source range retrieval.

## Evidence Handling

Evidence remains the trust primitive.

KB evidence should be canonical as:

- stable evidence identifier,
- source identifier,
- source display metadata,
- locator into the raw source,
- optional robust selectors internally.

Workbench can present evidence in three forms:

- inline evidence metadata in the `ask` result,
- temporary MCP resources for source spans when useful,
- trace links from tasks/artifacts/issues/PRs back to evidence IDs.

Workbench should not treat excerpts as canonical evidence. Excerpts are display
or navigation conveniences. The durable reference is the KB evidence ID plus
source locator.

## Scope Mapping

Workbench owns agent scope. KB owns graph scope.

For MVP:

- Workbench passes selected namespace/project/role when available.
- KB may map that to one implicit default scope until it has first-class scope
  nodes.
- Workbench should not expose KB graph scope IDs unless a later RFC justifies
  it.

Future task and artifact selection should become contribution/query context, but
not necessarily public KB scope.

## Traceability and GitHub

Workbench should be the place where cross-system knowledge work becomes
traceable:

```text
Workbench RFC -> GitHub issue -> implementation task -> PR -> KB evidence
```

For this feature branch, GitHub issues should be created from the Workbench
delivery plan. Initial labels:

- `specification`
- `rfc`
- `knowledge-base`
- `integration`

GitHub issues are tracker records. They should mirror Workbench tasks or
artifacts once tracker support exists, but they should not become the core
Workbench domain primitive.

## Compatibility Path

The current Workbench `ask` path uses:

- `/content/search`,
- `/knowledge/query`,
- one synthesis call.

The migration path should be:

1. Keep the current path as compatibility while KB implements `/query`.
2. Add a new KB client shape for `/query`.
3. Teach `ask` to prefer `/query` when the configured KB supports it.
4. Keep old retrieval primitives behind a compatibility adapter only if needed.
5. Add `knowledge.contribute` after KB implements `/contributions`.
6. Remove old endpoint assumptions once KB and Workbench tests cover the new
   contract.

## Open Questions

1. Should Workbench translate KB `204 No Content` into an empty answer result or
   a structured `not_grounded` status?
2. Should source range retrieval be part of the first Workbench/KB integration
   slice?
3. Should the contribution tool be named `knowledge.contribute`,
   `knowledge.submit`, or something else?
4. Should statement-only contributions be accepted through Workbench without a
   source artifact?
5. Should `feedback` automatically create contribution candidates, or should
   that require an explicit later policy slice?
6. Should GitHub issues mirror artifacts, tasks, or both before Workbench has a
   durable artifact model?
7. Is `internal/librarian` still the preferred package name for external KB
   interaction?

## Recommended MVP Decisions

1. Treat this RFC as the master cross-system artifact.
2. Keep the KB repo RFC as service-local implementation input or replace it with
   a pointer to this RFC plus KB-specific details.
3. Implement Workbench `ask` against KB `/query`.
4. Preserve evidence and related topic fields in the MCP tool result.
5. Add a contribution MCP tool only after KB supports `/contributions`.
6. Keep source range retrieval out of the first slice unless evidence audit is
   blocked without it.
7. Create GitHub issues from the delivery plan and link them back to this RFC.
8. Do not expose KB graph internals in Workbench.

## Rejected Alternatives

### Keep the Master RFC in KB

Rejected because the feature is not service-local. Agents enter through
Workbench, and Workbench coordinates scope, tasks, artifacts, GitHub, and future
services.

### Make KB Speak MCP Directly

Rejected for MVP because Workbench is the ecosystem gateway. Adding MCP to KB
would duplicate entry points and weaken the integration boundary.

### Keep Separate KB Retrieval Choices in Workbench

Rejected as the target contract because agents should ask for knowledge, not
choose content search versus graph query. Compatibility may remain temporarily.

### Make GitHub Issues Core Workbench Domain Objects

Rejected because GitHub is one tracker adapter. Workbench should model tasks,
artifacts, trace links, and external references, then mirror or sync to GitHub.

## Final Position

Workbench is the agent-facing knowledge gateway.

KB is the evidence-backed knowledge service behind that gateway.

Cross-system behavior belongs in Workbench RFCs. Service-local implementation
details belong in the service repos and should link back to the Workbench master
artifact.
