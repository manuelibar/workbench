package middleware

import (
	"context"
	"errors"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/manuelibar/workbench/internal/errs"
)

type PublicError struct {
	Title     string
	Code      errs.Code
	Retryable bool
	Sentinel  error
}

type ErrorClassifier func(error) (PublicError, bool)

type ErrorLogger func(method string, err error, pub PublicError)

func ObserveCapabilityLists(server *mcpsdk.Server, markObserved func(string)) {
	server.AddReceivingMiddleware(func(next mcpsdk.MethodHandler) mcpsdk.MethodHandler {
		return func(ctx context.Context, method string, req mcpsdk.Request) (mcpsdk.Result, error) {
			res, err := next(ctx, method, req)
			if err == nil {
				markObserved(method)
			}
			return res, err
		}
	})
}

func SanitizeClassifiedErrors(server *mcpsdk.Server, classify ErrorClassifier, log ErrorLogger) {
	server.AddReceivingMiddleware(func(next mcpsdk.MethodHandler) mcpsdk.MethodHandler {
		return func(ctx context.Context, method string, req mcpsdk.Request) (mcpsdk.Result, error) {
			res, err := next(ctx, method, req)
			if err != nil {
				pub, ok := classify(err)
				if !ok {
					return res, err
				}
				log(method, err, pub)
				return nil, jsonrpcErrorFor(pub)
			}
			if method != "tools/call" {
				return res, nil
			}
			toolResult, ok := res.(*mcpsdk.CallToolResult)
			if !ok || !toolResult.IsError {
				return res, nil
			}
			pub, ok := classify(toolResult.GetError())
			if !ok {
				return res, nil
			}
			log(method, toolResult.GetError(), pub)
			return sanitizedToolResult(pub), nil
		}
	})
}

func sanitizedToolResult(pub PublicError) *mcpsdk.CallToolResult {
	return &mcpsdk.CallToolResult{
		Content: []mcpsdk.Content{&mcpsdk.TextContent{Text: pub.Title}},
		StructuredContent: map[string]any{
			"error": map[string]any{
				"title":     pub.Title,
				"code":      pub.Code.String(),
				"retryable": pub.Retryable,
			},
		},
		IsError: true,
	}
}

func jsonrpcErrorFor(pub PublicError) error {
	code := int64(jsonrpc.CodeInternalError)
	switch {
	case errors.Is(pub.Sentinel, errs.ErrNotFound):
		code = mcpsdk.CodeResourceNotFound
	case errors.Is(pub.Sentinel, errs.ErrInvalid):
		code = jsonrpc.CodeInvalidParams
	}
	return &jsonrpc.Error{
		Code:    code,
		Message: pub.Title,
	}
}
