# Contributing

## Local development

```bash
make compose-up
make test       # unit tests, -short
make vet
gofmt -l .
```

CI runs the same. Code must be `gofmt`-clean and pass under the race detector.

## Style

- Three runtime dependencies only: `github.com/modelcontextprotocol/go-sdk`,
  `github.com/jackc/pgx/v5`, `github.com/google/uuid`. New deps require
  discussion and a CHANGELOG entry.
- Per-symbol godoc on every exported identifier. Style: complete sentences,
  opening with the identifier name, with `[SymbolName]` references where
  helpful.
- Marker-interface options per sub-package where options apply.
- Tests in the `_test` external package (black-box). White-box tests go in
  `<file>_internal_test.go` and live in the source package.
- One concept per sub-package. If a layer accumulates concerns from several
  different domains, split it.
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
