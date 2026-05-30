# Plan 0002: Knowledge Base Integration

Status: Draft

RFC: `docs/rfc/0003-workbench-governed-knowledge-base.md`

Branch: `feature/knowledge-base-integration`

## Goal

Move the knowledge-base integration contract into Workbench and prepare the
first GitHub issue set for implementation.

Workbench is the MCP gateway. KB is the backing service for accepted knowledge,
evidence, and source locators.

## Initial GitHub Structure

Create labels:

- `specification`
- `rfc`
- `knowledge-base`
- `integration`

Create initial issues:

1. Adopt Workbench-owned KB integration RFC.
2. Add Workbench KB `/query` client path.
3. Add Workbench knowledge contribution flow.
4. Add evidence and source-range navigation.
5. Align KB service-local docs with the Workbench master RFC.

## Issue 1: Adopt Workbench-Owned KB Integration RFC

Primary manager: Documentation / ADR only

Scope:

- Add the master Workbench RFC for KB integration.
- Document the placement rule for cross-system docs.
- Keep service-local KB artifacts subordinate to the Workbench RFC.

Acceptance:

- Workbench docs include the master RFC.
- `docs/README.md` states the cross-system documentation ownership rule.
- The issue links to the RFC and this plan.

## Issue 2: Add Workbench KB `/query` Client Path

Primary manager: KnowledgeManager

Scope:

- Add a KB client shape for `POST /query`.
- Return `answer`, `evidence`, and `related_topics` through `query`.
- Preserve compatibility with `/content/search` and `/knowledge/query` while KB
  catches up.
- Decide how Workbench represents KB `204 No Content`.

Acceptance:

- Unit tests cover successful evidence-backed answers.
- Unit tests cover no-grounded-answer behavior.
- Existing `query` fallback behavior remains intact.
- No graph IDs, chunks, or raw retrieval scores leak to MCP results.

## Issue 3: Add Workbench Knowledge Contribution Flow

Primary manager: KnowledgeManager

Scope:

- Introduce a contribution-oriented MCP tool after KB supports
  `POST /contributions`.
- Map Workbench scope into contribution context.
- Keep corrections as contributions.
- Keep `feedback` separate from KB submission unless policy explicitly promotes
  feedback into a contribution candidate.

Acceptance:

- Tool input supports statement, focus, optional source reference, and scope.
- Tool output returns KB decision outcome, reasons, and clarification questions.
- Tests cover accepted, rejected, and needs-clarification responses.

## Issue 4: Add Evidence and Source-Range Navigation

Primary manager: MCP protocol surface

Scope:

- Decide whether Workbench needs temporary evidence resources in the first
  implementation.
- If source range retrieval is available, expose source spans as Workbench views
  without making them canonical evidence.
- Link evidence IDs into tasks/artifacts/traces once those models exist.

Acceptance:

- Evidence metadata appears in `query` results.
- Source-span resources, if added, clearly identify their KB source URI and
  locator.
- Excerpts are documented as convenience views, not canonical evidence.

## Issue 5: Align KB Service-Local Docs

Primary manager: Documentation / ADR only

Scope:

- Update the KB repo's draft RFC or follow-up artifact to point to the
  Workbench master RFC.
- Keep KB-specific HTTP, storage, graph, guardian, evidence locator, and source
  retrieval details in the KB repo.

Acceptance:

- KB docs identify Workbench as the master cross-system RFC owner.
- KB docs retain enough implementation detail for service work.
- No Workbench/MCP resource URIs become part of the KB service contract.

## Verification

Documentation-only changes:

```sh
make test
```

Implementation slices should also run:

```sh
make build
go test -race ./...
```

MCP behavior should be verified with the local Workbench MCP test flow once
available.
