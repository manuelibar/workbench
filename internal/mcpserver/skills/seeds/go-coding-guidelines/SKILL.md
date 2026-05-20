# Go Coding Guidelines

## Trigger

Use this skill whenever you are writing, reviewing, refactoring, or debugging Go code in the selected project.

## Skip

Skip if the selected project has no Go code, or if the task is purely documentation, planning, or infrastructure with no Go changes.

---

You are working in a Go codebase. Prefer small, idiomatic, boring Go that is easy to test and maintain.

## Core principles

- Keep packages cohesive and names short, specific, and domain-oriented.
- Prefer simple control flow over clever abstractions.
- Return errors with useful context; do not swallow errors.
- Keep exported APIs intentionally small and documented.
- Make behavior testable before adding configurability.
- Use the standard library unless a dependency clearly pays for itself.

## Formatting and tooling

Before considering Go changes done:

1. Run `gofmt` on modified Go files.
2. Run the project test command, usually `go test ./...` or the repo's `make test`.
3. If concurrency, I/O, or shared state changed, prefer `go test -race ./...`.
4. Keep generated or vendored files out of manual edits unless the task explicitly requires them.

## Error handling

- Check every returned error unless the API contract makes the ignored error provably irrelevant.
- Wrap lower-level errors at boundaries with `%w` so callers can inspect them.
- Use sentinel errors only when callers need branching behavior.
- Avoid logging and returning the same error from the same layer; choose one owner for reporting.

Example style:

```go
if err := store.Save(ctx, item); err != nil {
    return fmt.Errorf("save item %q: %w", item.ID, err)
}
```

## Concurrency

- Protect shared mutable state with a clear ownership rule: mutex, channel owner, or immutable snapshot.
- Keep locked sections small and avoid calling external code while holding a lock.
- Respect `context.Context` for cancellation on I/O, RPC, subprocesses, and long-running operations.
- Prefer deterministic tests for concurrent behavior; avoid sleeps unless testing timing itself.

## Interfaces and abstractions

- Accept interfaces at boundaries when they make tests or alternate implementations simpler.
- Return concrete types unless callers genuinely need an abstraction.
- Avoid premature generic helpers. Duplicate two small call sites before introducing a vague abstraction.
- Keep interface names behavior-focused: `Store`, `Clock`, `Runner`, `Notifier`.

## Tests

- Use table-driven tests when cases share one behavior.
- Name tests after observable behavior, not implementation details.
- Include failure-path coverage for parsing, I/O, invalid input, and permission checks.
- Prefer real standard-library components over mocks when cheap: `httptest`, `fstest`, temp dirs, in-memory stores.
- Assert on outputs, side effects, and errors that matter to users or callers.

## Project hygiene

- Keep package-level state rare and explicit.
- Do not introduce global configuration that makes tests order-dependent.
- Keep CLI entrypoints thin; put business logic in internal packages.
- Update README, ADRs, or docs when behavior or public contracts change.

## Verification checklist

- [ ] Modified Go files are `gofmt` formatted.
- [ ] Tests covering the changed behavior were added or updated.
- [ ] `go test ./...` or the repo's test command passes.
- [ ] Race-sensitive changes were checked with `go test -race ./...`.
- [ ] Errors include enough context for debugging.
- [ ] New exported identifiers have comments or are intentionally unexported.
