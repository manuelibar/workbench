---
id: "agent-skill-distribution-initial-implementation-plan"
type: "implementation_plan"
title: "Agent Skill Distribution Initial Implementation Plan"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Agent Skill Distribution Initial Implementation Plan

## Objective

Implement the first Workbench slice for agent skill distribution by exposing a
read-only MCP resource catalog for local skill packages, then adding explicit
install planning and approved activation flows after the resource contract is
stable.

This plan is produced by
[agent-skill-distribution-rfc.md](agent-skill-distribution-rfc.md) and should be
kept aligned with that RFC as drilling continues.

## Steps

1. Define the skill package manifest schema in docs before code, including
   identity, version, digest, package files, compatibility, requested
   capabilities, trust metadata, and install targets.
2. Add local fixtures for an instruction-only skill and a skill with
   references, assets, and scripts. Fixtures should be inert and should not
   require network access.
3. Implement read-only catalog discovery through MCP resources:
   `workbench-skill:///catalog`, per-skill manifests, versioned instructions,
   and versioned package file resources.
4. Add resource templates only after the plain resource list works, so hosts
   without template support can still inspect the catalog.
5. Emit resource list-changed notifications when the catalog changes if the
   server advertises the resources `listChanged` capability.
6. Add manifest validation that rejects missing identity, missing instructions,
   ambiguous versions, digest mismatches, unsafe paths, and undeclared
   executable payloads.
7. Add an install-plan tool that returns a dry-run plan with no filesystem
   writes. The plan must include target agent profile, target paths,
   version/digest, requested capabilities, approvals needed, and rollback hint.
8. Add an approved install-apply tool only after the plan format is covered by
   tests. Apply must verify the digest and package version from the plan before
   writing anything.
9. Add activation and deactivation flows after install records exist. Activation
   must separate implicit invocation policy from explicit invocation support.
10. Add a Codex adapter that can map a compatible package into Codex skill or
    plugin conventions without making Codex-specific fields required for
    generic MCP resource discovery.
11. Add update flow support after installation and activation are stable.
    Updates must compare installed digest/version with available versions and
    produce a reviewable plan before mutation.
12. Document host compatibility behavior for generic MCP clients, Codex CLI,
    Codex IDE/app surfaces, and hosts with resources-only support.

## Verification

- Unit tests validate manifest parsing, normalized IDs, safe relative paths,
  digest calculation, version selection, compatibility summaries, and rejected
  malformed packages.
- MCP integration tests verify resource listing, resource reading, MIME types,
  sizes, annotations, templates, and list-changed notifications.
- Mutation tests verify install planning is dry-run only, apply rejects digest
  drift, script files are never executed during discovery, and activation
  requires explicit approval metadata.
- Fixture tests verify Codex-compatible package metadata can be preserved while
  generic MCP clients can still discover and inspect the same package.
- Manual verification confirms an agent can discover a skill, read only the
  manifest, inspect instructions on demand, request an install plan, and stop
  before mutation.

## Rollback

The first runtime change should be split so read-only resource exposure can be
reverted independently from install and activation tools.

If resource exposure behaves incorrectly, disable the skill resource capability
and fall back to the existing Workbench artifact-only MCP surface.

If install or activation tools behave incorrectly, keep catalog resources
enabled but remove the mutation tools from the capability plan. Installed test
fixtures should be cleaned by deleting only the target paths recorded in the
install plan or install record.

## Source References

- [agent-skill-distribution-rfc.md](agent-skill-distribution-rfc.md)
- [agent-skill-distribution-concept-map.md](agent-skill-distribution-concept-map.md)
- [agent-skill-distribution-risk.md](agent-skill-distribution-risk.md)
- [MCP resources](https://modelcontextprotocol.io/specification/2025-11-25/server/resources)
- [MCP tools](https://modelcontextprotocol.io/specification/2025-11-25/server/tools)
- [Codex Agent Skills](https://developers.openai.com/codex/skills)
- [Codex plugin build documentation](https://developers.openai.com/codex/plugins/build)

## Open Questions

Open human decisions are tracked in the RFC's Human-in-the-loop Index.
