# Artifact Conventions

Artifacts are Markdown documents. By default Workbench stores them under
`docs/artifacts/`; when `WORKBENCH_STORAGE_URL` is set, Workbench stores them
through the autonomous storage service and treats `storage:///...` as the
artifact location.

Required frontmatter:

```yaml
---
id: "stable-id"
type: "rfc"
title: "Artifact title"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---
```

For file-backed artifacts, the file name is `<id>.md`. For storage-backed
artifacts, the storage service owns the physical object key. Artifact IDs are
stable and are placed in scope through `contextualize(artifact_id=...)`.

The public MCP workflow is:

- `artifact.find` returns summaries without Markdown.
- `artifact.create` creates a typed draft and returns its summary.
- `contextualize(artifact_id=...)` checks out one artifact into scope and
  exposes `workbench:///artifacts/<id>`.
- MCP resource reads return the local scoped Markdown when available, so local
  edits are visible before upload.
- `artifact.upload` persists a full Markdown replacement for the artifact in
  scope. It takes no `artifact_id`; pass optional `markdown` or omit it to use
  the server-managed local scoped file.

Supported contract types include `rfc`, `adr`, `prd`, `requirement`, `spec`,
`research_note`, `risk`, `assumption`, `constraint`, `test_strategy`,
`implementation_plan`, `runbook`, `charter`, `problem_statement`,
`opportunity`, and `decision_record`.

Contract validation is deterministic:

- required frontmatter keys must be present and non-empty.
- the artifact type must exist in the registry.
- each required section must exist and contain a body that is not blank,
  `todo`, `tbd`, or `n/a`.

RFCs are living drill hubs. They may accumulate links, research sessions,
alternatives, append-only notes, decisions, and follow-up threads. When an RFC
drill session produces concrete work, create a typed action artifact and link
it from the RFC. The action artifact must link back to the RFC.

Every RFC must include this subsection under `## Open Questions`:

```markdown
### Human-in-the-loop Index

| ID | Nudge | Type | Why it matters | Blocks | Default if unanswered |
|---|---|---|---|---|---|
```

Every human nudge in a packet must be indexed there. Allowed nudge types are
`question`, `decision`, `challenge`, `approval`, and `tradeoff`.

Advanced lineage, delete, archive, supersession, elicitation, sign-off, and
review workflows are deferred to the Artifact v2 epic.
