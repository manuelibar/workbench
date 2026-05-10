package pgstore_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/manuelibar/workbench/internal/domain"
	"github.com/manuelibar/workbench/internal/pgstore"
)

func TestStore_ProjectLifecycle(t *testing.T) {
	s, ctx, _ := openForIntegration(t)

	ns, err := s.CreateNamespace(ctx, domain.Namespace{Name: "proj-test-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("CreateNamespace: %v", err)
	}
	t.Cleanup(func() { _ = s.DeleteNamespace(ctx, ns.ID) })

	p, err := s.CreateProject(ctx, domain.Project{
		NamespaceID: &ns.ID,
		Name:        "alpha",
		Description: "first",
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	got, err := s.GetProject(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetProject: %v", err)
	}
	if got.Name != "alpha" {
		t.Errorf("name roundtrip differs")
	}

	list, err := s.ListProjects(ctx, &ns.ID)
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("ListProjects returned %d, want 1", len(list))
	}

	_, err = s.CreateProject(ctx, domain.Project{NamespaceID: &ns.ID, Name: "alpha"})
	if !errors.Is(err, pgstore.ErrConflict) {
		t.Errorf("expected ErrConflict on duplicate name; got: %v", err)
	}

	newName := "alpha-renamed"
	updated, err := s.UpdateProject(ctx, p.ID, pgstore.UpdateProjectFields{Name: &newName})
	if err != nil {
		t.Fatalf("UpdateProject: %v", err)
	}
	if updated.Name != newName {
		t.Errorf("name not updated")
	}

	if err := s.DeleteProject(ctx, p.ID); err != nil {
		t.Fatalf("DeleteProject: %v", err)
	}
	if _, err := s.GetProject(ctx, p.ID); !errors.Is(err, pgstore.ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete; got: %v", err)
	}
}

func TestStore_ProjectStandalone(t *testing.T) {
	s, ctx, _ := openForIntegration(t)
	p, err := s.CreateProject(ctx, domain.Project{Name: "standalone-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("CreateProject standalone: %v", err)
	}
	t.Cleanup(func() { _ = s.DeleteProject(ctx, p.ID) })
	if p.NamespaceID != nil {
		t.Errorf("standalone project namespace_id = %v, want nil", p.NamespaceID)
	}
}
