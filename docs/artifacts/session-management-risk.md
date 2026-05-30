---
id: "session-management-risk"
type: "risk"
title: "Session History Becomes Hidden Transcript Storage"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Session History Becomes Hidden Transcript Storage

## Description

Session event history could expand from concise operational events into a
shadow transcript store that captures raw prompts, tool arguments, command
output, secrets, or personal data without explicit user intent.

## Impact

The risk would make Workbench harder to trust and harder to clean up. It would
also increase the blast radius of local file disclosure, complicate retention
policy, and blur the boundary between this epic and future memory or knowledge
systems.

## Likelihood

Medium. Event sourcing is attractive for resume and auditability, but without a
strict payload policy it is easy for implementation code to store raw MCP
messages or complete tool results because those are convenient to capture.

## Mitigation

Define normalized event types with bounded payloads. Store artifact IDs,
section keys, status summaries, hashes, and short notes by default. Do not
store raw MCP messages, full tool outputs, or secrets unless a later explicit
artifact defines an opt-in event class with retention and redaction rules.

Add tests that fail if event payloads exceed configured size limits or include
fields reserved for raw transcript content. Make retention defaults visible in
`session.current`, `session.get`, or equivalent inspection surfaces.

## Owner

The session management epic owner owns the payload policy. Future memory and
knowledge epic owners own any separate long-term semantic storage they add.

## Source References

- [RFC: Session Management](session-management-rfc.md)
- [MCP security best practices](https://modelcontextprotocol.io/docs/tutorials/security/security_best_practices)
- [MCP transport specification, version 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25/basic/transports)

## Open Questions

`SM-HIL-002` and `SM-HIL-003` in the RFC decide how much session history is
retained and how event payloads are shaped.
