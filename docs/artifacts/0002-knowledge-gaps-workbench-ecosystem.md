# 0002 — Knowledge Gaps and Research Questions

Status: Draft
Depends on: 0001

## Domain Model Gaps
1. What is the minimal canonical model? (Namespace, Project, Role, Artifact, Service, Selection)
2. Do we keep `Application` at all, or remove it entirely from MVP?
3. How do we represent context mutation semantics across turns?

## Architecture Gaps
4. What is the clean boundary between Workbench and Ripple after reset?
5. Which capabilities belong in Workbench core vs external services via adapters?
6. What eventing contract is required now (if any) vs later?

## Documentation Service Gaps
7. What artifact types are mandatory in MVP? (Problem Statement, RFC, ADR, Spec, Slice)
8. Should docs live in Git-first markdown with render pipeline as primary source of truth?
9. What traceability schema links artifacts to backlog items?

## Backlog Service Gaps
10. Is backlog-service adopted as-is, or does it require schema/domain changes now?
11. What minimum fields are needed for traceability to artifacts?
12. What workflow states are mandatory vs optional in MVP?

## Delivery & Governance Gaps
13. Do we create a new GitHub org now or keep under existing account until MVP proves itself?
14. What repo topology do we want after merge/reset? (mono-repo vs ecosystem repos)
15. What is the first vertical slice that proves value in <2 weeks?

## Acceptance for this artifact
- Each gap must be answered by either:
  - a decision in an RFC/ADR, or
  - an explicit deferred decision note with rationale.
