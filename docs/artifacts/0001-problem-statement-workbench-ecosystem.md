# 0001 — Problem Statement: Workbench Ecosystem Reset

Status: Draft
Owner: Hermes (PM role) + Manuel
Date: 2026-05-16

## Context
We have multiple half-baked repos and overlapping concepts (Workbench, Ripple, backlog-service, and related utilities). The vision is strong, but implementation is fragmented and traceability is weak.

## Problem
We do not yet have a single coherent, artifact-driven system architecture that:
1. Lets an agent select project context and mutate capabilities safely.
2. Scales from personal to multi-organization use.
3. Preserves rigorous documentation and decision traceability.
4. Avoids vendor lock-in by exposing our own service abstractions.

## In Scope
- Re-challenge existing assumptions and repo boundaries.
- Define Workbench as ecosystem entrypoint (MCP harness/facade).
- Start with project selection and add namespaces.
- Treat role selection as first-class context mutation.
- Build artifact-centric documentation discipline and publishing path.
- Prioritize backlog service + documentation service.

## Out of Scope (for MVP)
- Full AFK autonomous loop system implementation.
- Large-scale multi-tenant hosting hardening.
- Final production-grade auth and billing.
- Rich media asset platform (S3 VFS abstraction) beyond minimal design notes.

## Constraints
- Intellectual honesty: clear separation of implemented vs proposed.
- Configurability over hardcoded personal workflow.
- Artifacts must be durable and reviewable.

## Desired Outcome
A focused MVP path that produces a usable Workbench ecosystem foundation and creates reusable artifacts for all future decisions.
