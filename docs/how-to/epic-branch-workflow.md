# Epic Branch Workflow

Epic branches start from the current foundation `main`. The current repository
state, current `main` documentation, targeted web research, and explicitly
named external repositories are the authoritative research base for new epic
packets.

Archive branches are stale orientation material. Agents may inspect them to
understand historical direction, dead-code cleanup candidates, or missing
primitives, but final artifacts must not cite archive branches, link to them,
or treat them as authoritative evidence.

Each epic branch starts with a self-contained kickoff artifact packet under
`docs/artifacts/`:

- charter
- problem statement
- taxonomy or concept map when useful
- assumptions and risks
- research note
- RFC draft

Stable packet artifacts should read as durable references. The RFC is the
living drill hub: it can collect links, research sessions, alternatives,
append-only notes, decisions, and follow-up threads. When drilling creates
actionable work, add action artifacts such as specs, implementation plans,
requirements, constraints, test strategies, decision records, ADRs, rollout
plans, or integration plans.

Every RFC must include this subsection under `## Open Questions`:

```markdown
### Human-in-the-loop Index

| ID | Nudge | Type | Why it matters | Blocks | Default if unanswered |
|---|---|---|---|---|---|
```

Every human nudge in the packet must appear in that index. Allowed nudge types
are `question`, `decision`, `challenge`, `approval`, and `tradeoff`.

The expected branch loop is:

1. Check out the epic branch created from current `main`.
2. Read current `main` docs and the branch artifacts.
3. Run targeted research with authoritative sources where possible.
4. Update only `docs/artifacts/` during the bootstrap pass.
5. Keep the kickoff packet self-contained.
6. Use the RFC as the living drill hub and link any action artifacts from it.
7. Make action artifacts link back to the RFC that produced them.
8. Commit one docs-only bootstrap commit and push the branch.

Runtime behavior, code ports, and tests belong to later implementation passes
after the epic packet establishes the contract.
