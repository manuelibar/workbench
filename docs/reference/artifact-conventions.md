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

Advanced archive, lineage, delete, elicitation, and review workflows are
deferred to the artifact advanced workflows epic.
