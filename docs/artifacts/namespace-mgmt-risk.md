---
id: "namespace-mgmt-risk"
type: "risk"
title: "Namespace Model Becomes a Planning Model"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Namespace Model Becomes a Planning Model

## Description

Namespace management could accidentally absorb boards, prioritization, WIP,
daily planning, or backlog semantics because project and repository scopes are
often used by planning tools. That would make the epic larger than scope
identity and create overlap with future planning-focused epics.

## Impact

If this risk materializes, namespace records may become overloaded with fields
that do not belong to identity, such as priority, status column, sprint, due
date, or daily plan membership. This would make namespace selection harder to
reuse for artifacts, notes, memory, GitHub synchronization, and local working
area discovery.

It would also make the bootstrap packet less durable because changes in
planning behavior would force changes in the core namespace model.

## Likelihood

Likelihood is medium. The terms "project" and "workspace" are commonly used by
planning and collaboration products, and GitHub Projects are adjacent to
repositories and organizations. Confidence is high that an explicit boundary is
needed.

## Mitigation

- Keep namespace records limited to identity, aliases, metadata evidence, and
  relationships.
- Model GitHub Projects, issues, boards, and prioritization as downstream
  dependencies or consumers, not namespace-owned entities.
- Add contract tests in implementation work to reject planning fields in the
  namespace core model.
- Require future planning artifacts to link to namespace IDs rather than adding
  planning state to namespaces.
- Keep the RFC Human-in-the-loop Index focused on identity decisions only.

## Owner

The namespace management epic owner owns this risk until implementation hands a
stable namespace contract to consuming epics.

## Source References

- [Namespace management charter](namespace-mgmt-charter.md)
- [Namespace management concept map](namespace-mgmt-concept-map.md)
- [GitHub GraphQL Projects](https://docs.github.com/en/graphql/reference/projects)

## Open Questions

No additional human nudge is needed beyond the RFC index. The mitigation
defaults keep planning behavior out of this epic.
