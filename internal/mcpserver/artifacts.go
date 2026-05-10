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

// ArtifactWire is the JSON shape artifact tools return.
type ArtifactWire struct {
	ID            string    `json:"id"`
	ProjectID     string    `json:"project_id"`
	Type          string    `json:"type"`
	Status        string    `json:"status"`
	Parents       []string  `json:"parents"`
	LatestVersion int       `json:"latest_version"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func artifactToWire(a domain.Artifact) ArtifactWire {
	parents := make([]string, len(a.Parents))
	for i, p := range a.Parents {
		parents[i] = p.String()
	}
	return ArtifactWire{
		ID:            a.ID.String(),
		ProjectID:     a.ProjectID.String(),
		Type:          a.Type,
		Status:        a.Status,
		Parents:       parents,
		LatestVersion: a.LatestVersion,
		CreatedAt:     a.CreatedAt,
		UpdatedAt:     a.UpdatedAt,
	}
}

// ArtifactVersionWire bundles a single version's content.
type ArtifactVersionWire struct {
	Version     int            `json:"version"`
	Content     map[string]any `json:"content,omitempty"`
	ContentText string         `json:"content_text,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

// artifact.create --------------------------------------------------------

type ArtifactCreateArgs struct {
	Type        string         `json:"type" jsonschema:"free-form artifact type, e.g. 'note', 'prd', 'spec', 'task'"`
	Status      string         `json:"status,omitempty" jsonschema:"draft (default) | reviewing | signed_off | archived"`
	Parents     []string       `json:"parents,omitempty" jsonschema:"optional parent artifact UUIDs (lineage)"`
	Content     map[string]any `json:"content,omitempty" jsonschema:"initial structured body (becomes version 1 if non-empty)"`
	ContentText string         `json:"content_text,omitempty" jsonschema:"plain-text projection of the body"`
}

type ArtifactCreateResult struct {
	Artifact ArtifactWire `json:"artifact"`
}

func (s *Server) handleArtifactCreate(ctx context.Context, _ *mcp.CallToolRequest, args ArtifactCreateArgs) (*mcp.CallToolResult, ArtifactCreateResult, error) {
	pID, err := s.resolveProjectID("")
	if err != nil {
		return nil, ArtifactCreateResult{}, err
	}
	if args.Type == "" {
		return nil, ArtifactCreateResult{}, fmt.Errorf("artifact.create: type required")
	}
	a := domain.Artifact{
		ProjectID: pID,
		Type:      args.Type,
		Status:    args.Status,
	}
	if a.Status == "" {
		a.Status = domain.ArtifactStatusDraft
	}
	for _, raw := range args.Parents {
		pp, err := uuid.Parse(raw)
		if err != nil {
			return nil, ArtifactCreateResult{}, fmt.Errorf("artifact.create: parent: %w", err)
		}
		a.Parents = append(a.Parents, pp)
	}
	if x, ok := id.FromContext(ctx); ok {
		a.IdempotencyKey = x.IdempotencyKey
	}
	out, err := s.store.CreateArtifact(ctx, pgstore.CreateArtifactInput{
		Artifact:    a,
		Content:     args.Content,
		ContentText: args.ContentText,
	})
	if err != nil {
		return nil, ArtifactCreateResult{}, err
	}
	return nil, ArtifactCreateResult{Artifact: artifactToWire(out)}, nil
}

// artifact.list ----------------------------------------------------------

type ArtifactListArgs struct {
	Type   string `json:"type,omitempty"`
	Status string `json:"status,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

type ArtifactListResult struct {
	Artifacts []ArtifactWire `json:"artifacts"`
	Count     int            `json:"count"`
}

func (s *Server) handleArtifactList(ctx context.Context, _ *mcp.CallToolRequest, args ArtifactListArgs) (*mcp.CallToolResult, ArtifactListResult, error) {
	pID, err := s.resolveProjectID("")
	if err != nil {
		return nil, ArtifactListResult{}, err
	}
	list, err := s.store.ListArtifacts(ctx, pID, pgstore.ListArtifactsFilter{
		Type:   args.Type,
		Status: args.Status,
		Limit:  args.Limit,
	})
	if err != nil {
		return nil, ArtifactListResult{}, err
	}
	out := make([]ArtifactWire, len(list))
	for i, a := range list {
		out[i] = artifactToWire(a)
	}
	return nil, ArtifactListResult{Artifacts: out, Count: len(out)}, nil
}

// artifact.get -----------------------------------------------------------

type ArtifactGetArgs struct {
	ID      string `json:"id" jsonschema:"artifact UUID"`
	Version int    `json:"version,omitempty" jsonschema:"specific version; 0 or omitted = latest"`
}

type ArtifactGetResult struct {
	Artifact ArtifactWire        `json:"artifact"`
	Version  ArtifactVersionWire `json:"version"`
}

func (s *Server) handleArtifactGet(ctx context.Context, _ *mcp.CallToolRequest, args ArtifactGetArgs) (*mcp.CallToolResult, ArtifactGetResult, error) {
	id, err := uuid.Parse(args.ID)
	if err != nil {
		return nil, ArtifactGetResult{}, fmt.Errorf("artifact.get: id: %w", err)
	}
	a, err := s.store.GetArtifact(ctx, id)
	if err != nil {
		return nil, ArtifactGetResult{}, err
	}
	v, err := s.store.GetArtifactVersion(ctx, id, args.Version)
	if err != nil {
		return nil, ArtifactGetResult{}, err
	}
	return nil, ArtifactGetResult{
		Artifact: artifactToWire(a),
		Version: ArtifactVersionWire{
			Version:     v.Version,
			Content:     v.Content,
			ContentText: v.ContentText,
			CreatedAt:   v.CreatedAt,
		},
	}, nil
}

// artifact.update --------------------------------------------------------

type ArtifactUpdateArgs struct {
	ID          string         `json:"id" jsonschema:"artifact UUID"`
	Content     map[string]any `json:"content,omitempty" jsonschema:"new structured body; appended as a new version"`
	ContentText string         `json:"content_text,omitempty"`
	Status      string         `json:"status,omitempty" jsonschema:"new status; ignored if empty"`
}

type ArtifactUpdateResult struct {
	Artifact   ArtifactWire `json:"artifact"`
	NewVersion int          `json:"new_version,omitempty"`
}

func (s *Server) handleArtifactUpdate(ctx context.Context, _ *mcp.CallToolRequest, args ArtifactUpdateArgs) (*mcp.CallToolResult, ArtifactUpdateResult, error) {
	id, err := uuid.Parse(args.ID)
	if err != nil {
		return nil, ArtifactUpdateResult{}, fmt.Errorf("artifact.update: id: %w", err)
	}
	updated := domain.Artifact{}
	newVersion := 0
	if len(args.Content) > 0 || args.ContentText != "" {
		a, v, err := s.store.AppendArtifactVersion(ctx, id, args.Content, args.ContentText)
		if err != nil {
			return nil, ArtifactUpdateResult{}, err
		}
		updated, newVersion = a, v
	}
	if args.Status != "" {
		a, err := s.store.SetArtifactStatus(ctx, id, args.Status)
		if err != nil {
			return nil, ArtifactUpdateResult{}, err
		}
		updated = a
	}
	if updated.ID == uuid.Nil {
		// no-op patch — return current state
		a, err := s.store.GetArtifact(ctx, id)
		if err != nil {
			return nil, ArtifactUpdateResult{}, err
		}
		updated = a
	}
	return nil, ArtifactUpdateResult{Artifact: artifactToWire(updated), NewVersion: newVersion}, nil
}

// artifact.delete --------------------------------------------------------

type ArtifactDeleteArgs struct {
	ID string `json:"id"`
}

type ArtifactDeleteResult struct {
	Deleted bool   `json:"deleted"`
	ID      string `json:"id"`
}

func (s *Server) handleArtifactDelete(ctx context.Context, _ *mcp.CallToolRequest, args ArtifactDeleteArgs) (*mcp.CallToolResult, ArtifactDeleteResult, error) {
	id, err := uuid.Parse(args.ID)
	if err != nil {
		return nil, ArtifactDeleteResult{}, fmt.Errorf("artifact.delete: id: %w", err)
	}
	if err := s.store.DeleteArtifact(ctx, id); err != nil {
		return nil, ArtifactDeleteResult{}, err
	}
	return nil, ArtifactDeleteResult{Deleted: true, ID: id.String()}, nil
}

// artifact.attach --------------------------------------------------------

type ArtifactAttachArgs struct {
	ID       string `json:"id"        jsonschema:"artifact UUID to attach a parent to"`
	ParentID string `json:"parent_id" jsonschema:"parent artifact UUID (idempotent: duplicates are ignored)"`
}

type ArtifactAttachResult struct {
	Artifact ArtifactWire `json:"artifact"`
}

func (s *Server) handleArtifactAttach(ctx context.Context, _ *mcp.CallToolRequest, args ArtifactAttachArgs) (*mcp.CallToolResult, ArtifactAttachResult, error) {
	id, err := uuid.Parse(args.ID)
	if err != nil {
		return nil, ArtifactAttachResult{}, fmt.Errorf("artifact.attach: id: %w", err)
	}
	parentID, err := uuid.Parse(args.ParentID)
	if err != nil {
		return nil, ArtifactAttachResult{}, fmt.Errorf("artifact.attach: parent_id: %w", err)
	}
	a, err := s.store.AttachArtifactParent(ctx, id, parentID)
	if err != nil {
		return nil, ArtifactAttachResult{}, err
	}
	return nil, ArtifactAttachResult{Artifact: artifactToWire(a)}, nil
}

// artifact.sign_off / archive --------------------------------------------

type ArtifactStatusArgs struct {
	ID string `json:"id" jsonschema:"artifact UUID"`
}

type ArtifactStatusResult struct {
	Artifact ArtifactWire `json:"artifact"`
}

func (s *Server) handleArtifactSignOff(ctx context.Context, _ *mcp.CallToolRequest, args ArtifactStatusArgs) (*mcp.CallToolResult, ArtifactStatusResult, error) {
	id, err := uuid.Parse(args.ID)
	if err != nil {
		return nil, ArtifactStatusResult{}, fmt.Errorf("artifact.sign_off: id: %w", err)
	}
	a, err := s.store.SetArtifactStatus(ctx, id, domain.ArtifactStatusSignedOff)
	if err != nil {
		return nil, ArtifactStatusResult{}, err
	}
	return nil, ArtifactStatusResult{Artifact: artifactToWire(a)}, nil
}

func (s *Server) handleArtifactArchive(ctx context.Context, _ *mcp.CallToolRequest, args ArtifactStatusArgs) (*mcp.CallToolResult, ArtifactStatusResult, error) {
	id, err := uuid.Parse(args.ID)
	if err != nil {
		return nil, ArtifactStatusResult{}, fmt.Errorf("artifact.archive: id: %w", err)
	}
	a, err := s.store.SetArtifactStatus(ctx, id, domain.ArtifactStatusArchived)
	if err != nil {
		return nil, ArtifactStatusResult{}, err
	}
	return nil, ArtifactStatusResult{Artifact: artifactToWire(a)}, nil
}

// registerArtifacts wires the project-scoped artifact surface, including
// the lifecycle verbs (attach / sign_off / archive).
func (s *Server) registerArtifacts(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "artifact.create",
		Description: "Create a typed, versioned artifact in the currently-selected project. Optional initial content becomes version 1.",
	}, s.handleArtifactCreate)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "artifact.list",
		Description: "List artifacts in the currently-selected project, most-recently-updated first. Filters: type, status, limit.",
	}, s.handleArtifactList)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "artifact.get",
		Description: "Fetch an artifact and one of its versions (default: latest).",
	}, s.handleArtifactGet)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "artifact.update",
		Description: "Update an artifact: append a new version (if content is supplied) and/or change status.",
	}, s.handleArtifactUpdate)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "artifact.delete",
		Description: "Delete an artifact and all its versions.",
	}, s.handleArtifactDelete)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "artifact.attach",
		Description: "Attach a parent artifact to record lineage. Idempotent: duplicates are ignored.",
	}, s.handleArtifactAttach)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "artifact.sign_off",
		Description: "Move an artifact to status='signed_off'.",
	}, s.handleArtifactSignOff)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "artifact.archive",
		Description: "Move an artifact to status='archived'.",
	}, s.handleArtifactArchive)
}
