package mcpserver

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// AskArgs are the inputs to the `ask` tool: a question and an optional
// JSON-Schema-shaped form definition that constrains the user's response.
type AskArgs struct {
	Question string         `json:"question" jsonschema:"the question to ask the user"`
	Schema   map[string]any `json:"schema,omitempty" jsonschema:"optional JSON Schema (top-level properties only) constraining the user's response; omit for free-form text"`
}

// AskResult mirrors the SDK's [mcp.ElicitResult]: the user's action plus
// (when accepted) the structured content they submitted.
type AskResult struct {
	Action  string         `json:"action"`            // "accept" / "decline" / "cancel"
	Content map[string]any `json:"content,omitempty"` // only populated when Action == "accept"
}

func (s *Server) registerAsk(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "ask",
		Description: "Ask the connected user a question via MCP elicitation. Returns the user's action (accept/decline/cancel) and the submitted content when accepted. Requires the client to advertise the `elicitation` capability.",
	}, s.handleAsk)
}

func (s *Server) handleAsk(ctx context.Context, req *mcp.CallToolRequest, args AskArgs) (*mcp.CallToolResult, AskResult, error) {
	if args.Question == "" {
		return nil, AskResult{}, fmt.Errorf("ask: question must not be empty")
	}
	params := &mcp.ElicitParams{
		Message: args.Question,
	}
	if len(args.Schema) > 0 {
		params.RequestedSchema = args.Schema
	}
	res, err := req.Session.Elicit(ctx, params)
	if err != nil {
		return nil, AskResult{}, fmt.Errorf("ask: elicit: %w", err)
	}
	return nil, AskResult{Action: res.Action, Content: res.Content}, nil
}
