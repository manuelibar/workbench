# 1. Maintain architecture decisions as part of every specification

Date: 2026-05-19

## Status

Accepted

## Context

Workbench is a greenfield project with fast-moving architecture. The risk is not that we write too many ADRs; the risk is that implementation and documentation drift while abstractions such as session managers, project managers, skill registries, knowledge managers, task managers, and background managers are still being shaped.

GitHub issues are part of the project specification surface. If issues describe architectural work but do not mention ADR impact, agents may implement behavior without updating the decision record.

## Decision

Every architecture-affecting change must update or create ADRs in the same work item.

A change is architecture-affecting when it changes any of these:

- manager boundaries or interfaces
- MCP tool/resource/prompt contracts
- persistence or registry backends
- task/session/project/namespace state machines
- capability discovery behavior
- integration configuration behavior
- background execution or headless-agent routing
- knowledge management semantics

For greenfield Workbench work, ADRs may be rewritten in place when the old text is still provisional and no external users depend on it. Once a decision becomes externally relied upon, prefer a new ADR that supersedes the old one.

GitHub issue templates must include an ADR impact section so every specification explicitly answers whether ADRs must be added or updated.

## Consequences

This keeps docs, specs, and implementation synchronized. It also makes future agents safer: before changing code, they can inspect ADRs to understand the current contract rather than reconstructing intent from scattered conversations.

The cost is a little more documentation per issue, but that is acceptable because Workbench itself is an architecture-heavy project.
