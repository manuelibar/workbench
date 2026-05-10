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

// SkillWire is the JSON shape skill tools return.
type SkillWire struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Name      string    `json:"name"`
	BodyMD    string    `json:"body_md"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func skillToWire(s domain.Skill) SkillWire {
	return SkillWire{
		ID:        s.ID.String(),
		ProjectID: s.ProjectID.String(),
		Name:      s.Name,
		BodyMD:    s.BodyMD,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

func skillsToWire(in []domain.Skill) []SkillWire {
	out := make([]SkillWire, len(in))
	for i, s := range in {
		out[i] = skillToWire(s)
	}
	return out
}

// skill.create -----------------------------------------------------------

type SkillCreateArgs struct {
	Name   string `json:"name" jsonschema:"unique-within-project skill name"`
	BodyMD string `json:"body_md" jsonschema:"markdown body served as workbench:///skills/{name}"`
}

type SkillCreateResult struct {
	Skill SkillWire `json:"skill"`
}

func (s *Server) handleSkillCreate(ctx context.Context, _ *mcp.CallToolRequest, args SkillCreateArgs) (*mcp.CallToolResult, SkillCreateResult, error) {
	pID, err := s.resolveProjectID("")
	if err != nil {
		return nil, SkillCreateResult{}, err
	}
	sk := domain.Skill{ProjectID: pID, Name: args.Name, BodyMD: args.BodyMD}
	if a, ok := id.FromContext(ctx); ok {
		sk.IdempotencyKey = a.IdempotencyKey
	}
	created, err := s.store.CreateSkill(ctx, sk)
	if err != nil {
		return nil, SkillCreateResult{}, err
	}
	return nil, SkillCreateResult{Skill: skillToWire(created)}, nil
}

// skill.list / get / update / delete -------------------------------------

type SkillListResult struct {
	Skills []SkillWire `json:"skills"`
	Count  int         `json:"count"`
}

func (s *Server) handleSkillList(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, SkillListResult, error) {
	pID, err := s.resolveProjectID("")
	if err != nil {
		return nil, SkillListResult{}, err
	}
	list, err := s.store.ListSkills(ctx, pID)
	if err != nil {
		return nil, SkillListResult{}, err
	}
	return nil, SkillListResult{Skills: skillsToWire(list), Count: len(list)}, nil
}

type SkillGetArgs struct {
	ID   string `json:"id,omitempty" jsonschema:"skill UUID; alternative to name"`
	Name string `json:"name,omitempty" jsonschema:"skill name within the currently-selected project"`
}

type SkillGetResult struct {
	Skill SkillWire `json:"skill"`
}

func (s *Server) handleSkillGet(ctx context.Context, _ *mcp.CallToolRequest, args SkillGetArgs) (*mcp.CallToolResult, SkillGetResult, error) {
	if args.ID != "" {
		id, err := uuid.Parse(args.ID)
		if err != nil {
			return nil, SkillGetResult{}, fmt.Errorf("skill.get: id: %w", err)
		}
		sk, err := s.store.GetSkill(ctx, id)
		if err != nil {
			return nil, SkillGetResult{}, err
		}
		return nil, SkillGetResult{Skill: skillToWire(sk)}, nil
	}
	if args.Name == "" {
		return nil, SkillGetResult{}, fmt.Errorf("skill.get: id or name required")
	}
	pID, err := s.resolveProjectID("")
	if err != nil {
		return nil, SkillGetResult{}, err
	}
	sk, err := s.store.GetSkillByName(ctx, pID, args.Name)
	if err != nil {
		return nil, SkillGetResult{}, err
	}
	return nil, SkillGetResult{Skill: skillToWire(sk)}, nil
}

type SkillUpdateArgs struct {
	ID     string `json:"id" jsonschema:"skill UUID"`
	Name   string `json:"name,omitempty"`
	BodyMD string `json:"body_md,omitempty"`
}

type SkillUpdateResult struct {
	Skill SkillWire `json:"skill"`
}

func (s *Server) handleSkillUpdate(ctx context.Context, _ *mcp.CallToolRequest, args SkillUpdateArgs) (*mcp.CallToolResult, SkillUpdateResult, error) {
	id, err := uuid.Parse(args.ID)
	if err != nil {
		return nil, SkillUpdateResult{}, fmt.Errorf("skill.update: id: %w", err)
	}
	var f pgstore.UpdateSkillFields
	if args.Name != "" {
		name := args.Name
		f.Name = &name
	}
	if args.BodyMD != "" {
		body := args.BodyMD
		f.BodyMD = &body
	}
	sk, err := s.store.UpdateSkill(ctx, id, f)
	if err != nil {
		return nil, SkillUpdateResult{}, err
	}
	return nil, SkillUpdateResult{Skill: skillToWire(sk)}, nil
}

type SkillDeleteArgs struct {
	ID string `json:"id"`
}

type SkillDeleteResult struct {
	Deleted bool   `json:"deleted"`
	ID      string `json:"id"`
}

func (s *Server) handleSkillDelete(ctx context.Context, _ *mcp.CallToolRequest, args SkillDeleteArgs) (*mcp.CallToolResult, SkillDeleteResult, error) {
	id, err := uuid.Parse(args.ID)
	if err != nil {
		return nil, SkillDeleteResult{}, fmt.Errorf("skill.delete: id: %w", err)
	}
	if err := s.store.DeleteSkill(ctx, id); err != nil {
		return nil, SkillDeleteResult{}, err
	}
	return nil, SkillDeleteResult{Deleted: true, ID: id.String()}, nil
}

// registerSkills wires skill CRUD on srv.
func (s *Server) registerSkills(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "skill.create",
		Description: "Create a markdown skill in the currently-selected project. Body is served as the workbench:///skills/{name} resource.",
	}, s.handleSkillCreate)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "skill.list",
		Description: "List skills in the currently-selected project.",
	}, s.handleSkillList)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "skill.get",
		Description: "Fetch a skill by id or by name (within the currently-selected project).",
	}, s.handleSkillGet)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "skill.update",
		Description: "Patch a skill's name and/or body.",
	}, s.handleSkillUpdate)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "skill.delete",
		Description: "Delete a skill by id.",
	}, s.handleSkillDelete)
}
