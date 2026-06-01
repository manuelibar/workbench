# Artifact Conventions

Artifacts are Markdown files under `docs/artifacts/`.

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

The file name is `<id>.md`. Artifact IDs are stable and are selected through
`context(artifact_id=...)`.

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
