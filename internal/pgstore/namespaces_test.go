package pgstore_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/manuelibar/workbench/internal/domain"
	"github.com/manuelibar/workbench/internal/pgstore"
)

func TestStore_NamespaceTree(t *testing.T) {
	s, ctx, _ := openForIntegration(t)

	rootName := "test-root-" + uuid.NewString()[:8]
	root, err := s.CreateNamespace(ctx, domain.Namespace{Name: rootName, Description: "root"})
	if err != nil {
		t.Fatalf("CreateNamespace root: %v", err)
	}
	t.Cleanup(func() { _ = s.DeleteNamespace(ctx, root.ID) })

	child, err := s.CreateNamespace(ctx, domain.Namespace{
		ParentID: &root.ID,
		Name:     "child",
	})
	if err != nil {
		t.Fatalf("CreateNamespace child: %v", err)
	}

	got, err := s.GetNamespace(ctx, child.ID)
	if err != nil {
		t.Fatalf("GetNamespace: %v", err)
	}
	if got.ParentID == nil || *got.ParentID != root.ID {
		t.Errorf("child parent_id mismatch")
	}

	siblings, err := s.ListNamespaces(ctx, &root.ID)
	if err != nil {
		t.Fatalf("ListNamespaces: %v", err)
	}
	if len(siblings) != 1 || siblings[0].ID != child.ID {
		t.Errorf("ListNamespaces returned %d entries; want [child]", len(siblings))
	}

	// Unique name within parent
	_, err = s.CreateNamespace(ctx, domain.Namespace{ParentID: &root.ID, Name: "child"})
	if !errors.Is(err, pgstore.ErrConflict) {
		t.Errorf("expected ErrConflict on duplicate name; got: %v", err)
	}

	// Update
	newDesc := "updated description"
	updated, err := s.UpdateNamespace(ctx, child.ID, pgstore.UpdateNamespaceFields{Description: &newDesc})
	if err != nil {
		t.Fatalf("UpdateNamespace: %v", err)
	}
	if updated.Description != newDesc {
		t.Errorf("description not updated")
	}

	// Cascade delete on root removes child too
	if err := s.DeleteNamespace(ctx, root.ID); err != nil {
		t.Fatalf("DeleteNamespace root: %v", err)
	}
	if _, err := s.GetNamespace(ctx, child.ID); !errors.Is(err, pgstore.ErrNotFound) {
		t.Errorf("expected ErrNotFound after cascade; got: %v", err)
	}
}

func TestStore_NamespaceIdempotency(t *testing.T) {
	s, ctx, _ := openForIntegration(t)
	key := "ns-idemp-" + uuid.NewString()[:8]
	first, err := s.CreateNamespace(ctx, domain.Namespace{
		Name:           "ns-" + key,
		IdempotencyKey: key,
	})
	if err != nil {
		t.Fatalf("CreateNamespace first: %v", err)
	}
	t.Cleanup(func() { _ = s.DeleteNamespace(ctx, first.ID) })

	second, err := s.CreateNamespace(ctx, domain.Namespace{
		Name:           "different-name",
		IdempotencyKey: key,
	})
	if err != nil {
		t.Fatalf("CreateNamespace replay: %v", err)
	}
	if first.ID != second.ID {
		t.Errorf("idempotency replay returned different id")
	}
	if second.Name != first.Name {
		t.Errorf("idempotency replay returned different name (mutated)")
	}
}
