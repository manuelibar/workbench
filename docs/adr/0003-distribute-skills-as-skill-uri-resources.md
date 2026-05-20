# 3. Distribute skills as skill URI resources

Date: 2026-05-19

## Status

Accepted

## Context

Workbench should not require every agent host to share the same filesystem skill directory. Skills are part of the context/capability surface selected through Workbench, and agents should be able to discover and consume them through MCP after `refresh`.

## Decision

Workbench exposes skill bundles as MCP resources using canonical `skill://` URIs. Examples:

- `skill://workbench-orient/manifest`
- `skill://workbench-orient/SKILL.md`
- `skill://workbench-system-prompt/manifest`
- `skill://go-coding-guidelines/SKILL.md`

The `refresh` result and navigation resources point to these `skill://` resources. Legacy `workbench:///skills/...` parsing may remain temporarily for compatibility, but new capability discovery should use `skill://`. Language-specific guidance such as `go-coding-guidelines` is a better example than Workbench-only demo content because it represents the kind of reusable project policy Workbench should distribute.

## Consequences

Skills become a protocol-level distribution mechanism rather than a local filesystem convention. This enables future backing services and headless agents to generate, update, or specialize skills dynamically. The main risk is URI naming collision; future work should define namespace conventions such as `skill://codex/committer` or an equivalent encoded resource naming scheme.
