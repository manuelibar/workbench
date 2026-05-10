package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/manuelibar/workbench/internal/id"
)

// Slog returns a receiving middleware that logs each inbound request via
// log at debug-on-success / warn-on-error, including method, session id,
// request id, and duration.
//
// If log is nil, [slog.Default] is used.
func Slog(log *slog.Logger) mcp.Middleware {
	if log == nil {
		log = slog.Default()
	}
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			start := time.Now()
			res, err := next(ctx, method, req)
			attrs := []any{
				"mcp_method", method,
				"duration_ms", time.Since(start).Milliseconds(),
			}
			if a, ok := id.FromContext(ctx); ok {
				attrs = append(attrs, "request_id", a.RequestID.String())
				if a.CorrelationID != [16]byte{} {
					attrs = append(attrs, "correlation_id", a.CorrelationID.String())
				}
			}
			if sess := req.GetSession(); sess != nil {
				if sid := sess.ID(); sid != "" {
					attrs = append(attrs, "mcp_session_id", sid)
				}
			}
			if err != nil {
				attrs = append(attrs, "err", err.Error())
				log.Warn("mcp request failed", attrs...)
			} else {
				log.Debug("mcp request ok", attrs...)
			}
			return res, err
		}
	}
}
