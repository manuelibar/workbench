package middleware

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/manuelibar/workbench/internal/domain"
	"github.com/manuelibar/workbench/internal/id"
)

const methodCallTool = "tools/call"

// EventRecorder is implemented by a [pgstore.Store]; it is the minimal
// surface this middleware needs from the store. Defined here as an
// interface so middleware tests can pass a fake.
type EventRecorder interface {
	RecordEvent(ctx context.Context, e domain.Event) (domain.Event, error)
}

// Events returns a receiving middleware that records every successful and
// failed `tools/call` as an [domain.Event] row, attributed to
// workSessionID. Other RPC methods are passed through unchanged.
//
// If recorder is nil the middleware is a no-op; if log is nil
// [slog.Default] is used.
func Events(recorder EventRecorder, workSessionID uuid.UUID, log *slog.Logger) mcp.Middleware {
	if log == nil {
		log = slog.Default()
	}
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			res, err := next(ctx, method, req)
			if recorder == nil || method != methodCallTool {
				return res, err
			}

			toolName := ""
			if ctr, ok := req.(*mcp.CallToolRequest); ok && ctr != nil {
				toolName = ctr.Params.Name
			}

			ev := domain.Event{
				WorkSessionID: workSessionID,
				Type:          eventType(err, res),
				SubjectKind:   "tool",
				SubjectID:     toolName,
			}
			if sess := req.GetSession(); sess != nil {
				ev.MCPSessionID = sess.ID()
			}
			if a, ok := id.FromContext(ctx); ok {
				ev.RequestID = a.RequestID
				ev.CorrelationID = a.CorrelationID
				ev.CausationID = a.CausationID
				ev.IdempotencyKey = a.IdempotencyKey
			}
			if err != nil {
				ev.Payload = map[string]any{"error": err.Error()}
			}

			// Use a fresh context so a cancelled request still gets recorded.
			recCtx, cancel := context.WithTimeout(context.Background(), 5*1e9)
			defer cancel()
			if _, recErr := recorder.RecordEvent(recCtx, ev); recErr != nil {
				log.Warn("events: record failed", "err", recErr, "tool", toolName)
			}
			return res, err
		}
	}
}

// eventType returns "tool.failed" if the call errored, "tool.call" otherwise.
// A non-error result with [mcp.CallToolResult.IsError]=true is also recorded
// as "tool.failed" so the agent's self-corrective error path is observable.
func eventType(err error, res mcp.Result) string {
	if err != nil {
		return "tool.failed"
	}
	if r, ok := res.(*mcp.CallToolResult); ok && r != nil && r.IsError {
		return "tool.failed"
	}
	return "tool.call"
}
