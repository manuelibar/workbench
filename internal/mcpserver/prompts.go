package mcpserver

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/manuelibar/workbench/internal/domain"
	"github.com/manuelibar/workbench/internal/id"
	"github.com/manuelibar/workbench/internal/pgstore"
)

// PromptWire is the JSON shape prompt tools return.
type PromptWire struct {
	ID          string             `json:"id"`
	ProjectID   string             `json:"project_id"`
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Body        string             `json:"body"`
	Args        []domain.PromptArg `json:"args"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

func promptToWire(p domain.Prompt) PromptWire {
	args := p.Args
	if args == nil {
		args = []domain.PromptArg{}
	}
	return PromptWire{
		ID:          p.ID.String(),
		ProjectID:   p.ProjectID.String(),
		Name:        p.Name,
		Description: p.Description,
		Body:        p.Body,
		Args:        args,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func promptsToWire(in []domain.Prompt) []PromptWire {
	out := make([]PromptWire, len(in))
	for i, p := range in {
		out[i] = promptToWire(p)
	}
	return out
}

// prompt.create ----------------------------------------------------------

type PromptCreateArgs struct {
	Name        string             `json:"name" jsonschema:"unique-within-project prompt name"`
	Description string             `json:"description,omitempty"`
	Body        string             `json:"body" jsonschema:"prompt template body; may contain {{name}} placeholders"`
	Args        []domain.PromptArg `json:"args,omitempty" jsonschema:"declared template arguments"`
}

type PromptCreateResult struct {
	Prompt PromptWire `json:"prompt"`
}

func (s *Server) handlePromptCreate(ctx context.Context, _ *mcp.CallToolRequest, args PromptCreateArgs) (*mcp.CallToolResult, PromptCreateResult, error) {
	pID, err := s.resolveProjectID("")
	if err != nil {
		return nil, PromptCreateResult{}, err
	}
	p := domain.Prompt{
		ProjectID:   pID,
		Name:        args.Name,
		Description: args.Description,
		Body:        args.Body,
		Args:        args.Args,
	}
	if a, ok := id.FromContext(ctx); ok {
		p.IdempotencyKey = a.IdempotencyKey
	}
	created, err := s.store.CreatePrompt(ctx, p)
	if err != nil {
		return nil, PromptCreateResult{}, err
	}
	return nil, PromptCreateResult{Prompt: promptToWire(created)}, nil
}

// prompt.list / get / update / delete ------------------------------------

type PromptListResult struct {
	Prompts []PromptWire `json:"prompts"`
	Count   int          `json:"count"`
}

func (s *Server) handlePromptList(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, PromptListResult, error) {
	pID, err := s.resolveProjectID("")
	if err != nil {
		return nil, PromptListResult{}, err
	}
	list, err := s.store.ListPrompts(ctx, pID)
	if err != nil {
		return nil, PromptListResult{}, err
	}
	return nil, PromptListResult{Prompts: promptsToWire(list), Count: len(list)}, nil
}

type PromptGetArgs struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type PromptGetResult struct {
	Prompt PromptWire `json:"prompt"`
}

func (s *Server) handlePromptGet(ctx context.Context, _ *mcp.CallToolRequest, args PromptGetArgs) (*mcp.CallToolResult, PromptGetResult, error) {
	if args.ID != "" {
		id, err := uuid.Parse(args.ID)
		if err != nil {
			return nil, PromptGetResult{}, fmt.Errorf("prompt.get: id: %w", err)
		}
		p, err := s.store.GetPrompt(ctx, id)
		if err != nil {
			return nil, PromptGetResult{}, err
		}
		return nil, PromptGetResult{Prompt: promptToWire(p)}, nil
	}
	if args.Name == "" {
		return nil, PromptGetResult{}, fmt.Errorf("prompt.get: id or name required")
	}
	pID, err := s.resolveProjectID("")
	if err != nil {
		return nil, PromptGetResult{}, err
	}
	p, err := s.store.GetPromptByName(ctx, pID, args.Name)
	if err != nil {
		return nil, PromptGetResult{}, err
	}
	return nil, PromptGetResult{Prompt: promptToWire(p)}, nil
}

type PromptUpdateArgs struct {
	ID          string             `json:"id" jsonschema:"prompt UUID"`
	Name        string             `json:"name,omitempty"`
	Description string             `json:"description,omitempty"`
	Body        string             `json:"body,omitempty"`
	Args        []domain.PromptArg `json:"args,omitempty" jsonschema:"replace declared arguments; pass [] to clear"`
}

type PromptUpdateResult struct {
	Prompt PromptWire `json:"prompt"`
}

func (s *Server) handlePromptUpdate(ctx context.Context, _ *mcp.CallToolRequest, args PromptUpdateArgs) (*mcp.CallToolResult, PromptUpdateResult, error) {
	id, err := uuid.Parse(args.ID)
	if err != nil {
		return nil, PromptUpdateResult{}, fmt.Errorf("prompt.update: id: %w", err)
	}
	var f pgstore.UpdatePromptFields
	if args.Name != "" {
		name := args.Name
		f.Name = &name
	}
	if args.Description != "" {
		desc := args.Description
		f.Description = &desc
	}
	if args.Body != "" {
		body := args.Body
		f.Body = &body
	}
	if args.Args != nil {
		a := args.Args
		f.Args = &a
	}
	p, err := s.store.UpdatePrompt(ctx, id, f)
	if err != nil {
		return nil, PromptUpdateResult{}, err
	}
	return nil, PromptUpdateResult{Prompt: promptToWire(p)}, nil
}

type PromptDeleteArgs struct {
	ID string `json:"id"`
}

type PromptDeleteResult struct {
	Deleted bool   `json:"deleted"`
	ID      string `json:"id"`
}

func (s *Server) handlePromptDelete(ctx context.Context, _ *mcp.CallToolRequest, args PromptDeleteArgs) (*mcp.CallToolResult, PromptDeleteResult, error) {
	id, err := uuid.Parse(args.ID)
	if err != nil {
		return nil, PromptDeleteResult{}, fmt.Errorf("prompt.delete: id: %w", err)
	}
	if err := s.store.DeletePrompt(ctx, id); err != nil {
		return nil, PromptDeleteResult{}, err
	}
	return nil, PromptDeleteResult{Deleted: true, ID: id.String()}, nil
}

// registerPrompts wires prompt CRUD on srv.
func (s *Server) registerPrompts(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "prompt.create",
		Description: "Create a prompt template in the currently-selected project. Body may contain {{name}} placeholders matching declared args.",
	}, s.handlePromptCreate)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "prompt.list",
		Description: "List prompts in the currently-selected project.",
	}, s.handlePromptList)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "prompt.get",
		Description: "Fetch a prompt by id or name.",
	}, s.handlePromptGet)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "prompt.update",
		Description: "Patch a prompt's name / description / body / args.",
	}, s.handlePromptUpdate)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "prompt.delete",
		Description: "Delete a prompt by id.",
	}, s.handlePromptDelete)
}
