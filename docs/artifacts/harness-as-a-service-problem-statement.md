---
id: "harness-as-a-service-problem-statement"
type: "problem_statement"
title: "Harness as a Service Problem Statement"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Harness as a Service Problem Statement

## Context

Current `main` is intentionally small: a local stdio MCP server that manages
context and typed Markdown artifacts. Backlog, notes, memory, knowledge,
sessions, skills, and other higher-level Workbench features are expected to
return through epic branches that first define their contracts under
`docs/artifacts/`.

MCP gives Workbench a standard distribution mechanism for that larger feature
surface. The protocol already distinguishes model-controlled tools,
application-controlled resources, user-controlled prompts, and client-provided
roots. That split lets Workbench expose features to any compatible agent without
forcing those features to live inside one monolithic Workbench process.

Related service repositories point in the same direction: a control-plane
catalog and lifecycle service, a data-plane routing service, a remote MCP proxy,
layered configuration packages, and a local backlog service are all shaped as
small services or libraries with narrow contracts.

## Problem

Workbench needs a public harness architecture that makes features available
through MCP while avoiding a return to the old integrated server that mixed too
many domains into one process.

Without a harness contract, every feature epic can make different assumptions
about how tools are named, how resources are addressed, how backing services are
started, how defaults are disabled, and how a user later replaces a bundled
service. That creates lock-in to Workbench defaults, makes third-party extension
unclear, and prevents compatible agents from treating Workbench features as a
portable MCP surface.

The immediate challenge is to build enough of this architecture for personal
use first without prematurely solving full provider marketplaces, hosted
multi-tenancy, or production-grade remote access.

## Impact

The affected users are local Workbench users and agents that want a dependable
feature harness instead of one-off MCP servers per feature. The affected
maintainers are feature owners who need a clear line between the MCP distribution
contract and the standalone service architecture behind that contract.

If the harness boundary is not explicit, default features will become hard-coded
into the MCP server, provider replacement will require rewrites, and the later
public harness-as-a-service story will inherit local-only shortcuts as
accidental architecture.

## Constraints

- Bootstrap work for this epic is docs-only and lives under `docs/artifacts/`.
- The RFC must be the living drill hub and must index all human nudges.
- Final artifacts must derive from current worktree docs, current authoritative
  MCP docs, and current GitHub research.
- The proof of concept is personal-first: local, explicit, auditable, and useful
  before it is shareable.
- Configurability must be designed into the contracts, but deep provider
  override runtime is deferred.
- Backing services are allowed to have their own architecture, storage,
  deployment, and release model as long as they satisfy their provider contract.

## Source References

- [Harness RFC](./harness-as-a-service-rfc.md)
- `README.md`
- `docs/how-to/epic-branch-workflow.md`
- https://modelcontextprotocol.io/specification/2025-11-25/server/tools
- https://modelcontextprotocol.io/specification/2025-11-25/server/resources
- https://github.com/manuelibar/backlog-service

## Open Questions

Open human nudges are tracked centrally in
[the RFC Human-in-the-loop Index](./harness-as-a-service-rfc.md#human-in-the-loop-index).
