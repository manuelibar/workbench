# Error Handling

Workbench errors are classified where the failure is understood, decorated as
they move outward, and exposed only at control-flow or transport boundaries.

Use `errs.New` when creating a new failure boundary. Keep an `attrs` map near
the start of the control flow and add diagnostic fields as they become known:

```go
attrs := map[string]any{"artifact_id": id}
return errs.New(
	"Artifact not found",
	errs.WithSentinel(errs.ErrNotFound),
	errs.WithCode("workbench.artifact.not_found"),
	errs.WithAttrs(attrs),
)
```

Use `errs.Decorate` for the same failure plus more private context:

```go
attrs["tool"] = "artifact.upload"
return errs.Decorate(err, errs.WithAttrs(attrs))
```

`Decorate` accepts the same options as constructors, but `WithCause` is ignored
so decoration does not rewrite the failure cause. Sentinel, code, severity,
retryability, and attributes are metadata and can be patched as errors bubble.

Sentinels are for programmatic classes such as invalid input, missing data, or
dependency failure. Stable codes identify a specific Workbench failure in logs
and client contracts. Attributes are private diagnostics for logs and debugging;
they are not part of client responses.

Use `errs.NewMulti` when a workflow tolerates multiple failures and returns one
aggregate error at the end. Add children with `Add`; inspect them with
`errors.Is`, `errors.As`, or `Unwrap() []error`.

Handle or expose errors only at boundaries. The MCP boundary is responsible for
sanitizing classified errors: tool failures become `isError=true` results with a
public title, code, and retryability; resource failures become JSON-RPC errors.
Do not return raw filesystem, parser, or planner errors directly once the layer
knows what failed.
