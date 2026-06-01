# Contributing

## Local development

```bash
make test
make vet
make test-race
make smoke
```

Code must be `gofmt`-clean and pass the standard test suite. Use `make smoke`
when changing stdio MCP startup, capability relist behavior, or tool/resource
registration.

## Style

- Runtime dependencies are intentionally minimal:
  `github.com/modelcontextprotocol/go-sdk` and `github.com/google/uuid`. New
  deps require discussion and a CHANGELOG entry.
- Add godoc for exported identifiers that are part of the intended package
  surface. Style: complete sentences, opening with the identifier name, with
  `[SymbolName]` references where helpful.
- Keep `internal/mcp` focused on the current kernel: context state,
  capability planning/sync, artifact contracts, artifact file IO, MCP tools,
  and MCP resources. Defer unrelated product surfaces to epic branches until
  their artifact packets define the contract.
- Tests may live in the source package when they need package internals. Prefer
  black-box `_test` packages for behavior that can be exercised through public
  APIs.
- Concurrency: prefer CSP — channel-owned state with a goroutine actor — over
  `sync.Mutex` around shared maps. A short `sync.Mutex` is acceptable for
  tight, single-variable critical sections.

## Commits

- Small, reviewable commits. Each commit's message describes *why* the change
  is being made, not just *what* it does.
- Public-API or behaviour changes update `CHANGELOG.md` (`[Unreleased]`) in
  the same commit.

## Releases

`v0.x.y` until the surface settles. Breaking changes between minor releases
are allowed; tag a new minor and document the migration.
