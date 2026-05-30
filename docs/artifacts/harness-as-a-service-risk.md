---
id: "harness-as-a-service-risk"
type: "risk"
title: "Harness Defaults Become Accidental Lock-in"
status: "draft"
created: "2026-05-30T00:00:00Z"
updated: "2026-05-30T00:00:00Z"
---

# Harness Defaults Become Accidental Lock-in

## Description

The personal proof of concept could make Workbench-provided services so implicit
that the MCP feature contract, provider implementation, and local development
defaults become indistinguishable. That would make later provider replacement,
public extension, and opt-out behavior expensive.

The risk is not that Workbench ships defaults. Defaults are necessary for a
useful personal harness. The risk is that defaults leak into tool names,
resource URI semantics, manifest shape, error contracts, or feature ownership in
a way that prevents another provider from satisfying the same feature contract.

## Impact

If this risk materializes, the public harness-as-a-service architecture becomes
a repackaged local Workbench instance rather than a feature distribution layer.
Compatible agents would still see MCP tools, but users could not confidently
disable bundled services, swap providers, or install public extensions without
forking the harness.

The highest-impact failure mode is a hard-coded service path where MCP tools
call one bundled HTTP service directly with no provider binding. That would
force later override work to break public tool/resource contracts.

## Likelihood

Likelihood is medium. The project intentionally starts personal-first, and
personal-first implementations naturally prefer simple hard-coded defaults. The
confidence is high because the current related service repositories are narrow
and local-first, which is useful for speed but can hide provider boundaries if
the harness does not name them.

## Mitigation

- Require each feature package to name its public MCP projection separately from
  its default provider.
- Require provider bindings to carry `enabled`, `provider_id`, `transport`, and
  endpoint/config metadata even when the first provider is bundled.
- Implement opt-out behavior early enough that disabled defaults disappear from
  capability discovery rather than only failing at call time.
- Keep standalone service internals out of the harness contract; adapters map
  MCP calls to service APIs.
- Treat deep provider override runtime as follow-up work, but add tests that
  default services can be disabled and that their tool/resource names do not
  encode local deployment details.

## Owner

Branch owner for `epic/harness-as-a-service`, with review input from owners of
each default Workbench-provided service.

## Source References

- [Harness RFC](./harness-as-a-service-rfc.md)
- [Provider assumption](./harness-as-a-service-assumption.md)
- https://github.com/manuelibar/backlog-service
- https://github.com/manuelibar/go-config
- https://github.com/manuelibar/go-service-config-provider

## Open Questions

Open human nudges are tracked centrally in
[the RFC Human-in-the-loop Index](./harness-as-a-service-rfc.md#human-in-the-loop-index).
