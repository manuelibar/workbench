package middleware

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/manuelibar/workbench/internal/id"
)

// IDs returns a receiving middleware that ensures every inbound request has
// a four-part audit identifier set in its [context.Context]. v0 simply
// generates a fresh request id; later phases will extract incoming
// correlation/causation hints from request metadata.
func IDs() mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			ctx, _ = id.EnsureRequest(ctx)
			return next(ctx, method, req)
		}
	}
}
