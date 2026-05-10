package id_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/manuelibar/workbench/internal/id"
)

func TestNew_Unique(t *testing.T) {
	t.Parallel()
	a := id.New()
	b := id.New()
	if a == uuid.Nil || b == uuid.Nil {
		t.Fatal("expected non-nil UUIDs")
	}
	if a == b {
		t.Errorf("two consecutive ids collided: %s", a)
	}
	if v := a.Version(); v != 7 {
		t.Errorf("expected v7, got v%d", v)
	}
}

func TestEnsureRequest_GeneratesWhenAbsent(t *testing.T) {
	t.Parallel()
	ctx, a := id.EnsureRequest(context.Background())
	if a.RequestID == uuid.Nil {
		t.Fatal("RequestID is zero")
	}
	got, ok := id.FromContext(ctx)
	if !ok {
		t.Fatal("audit not attached to ctx")
	}
	if got.RequestID != a.RequestID {
		t.Errorf("ctx audit mismatch: %s vs %s", got.RequestID, a.RequestID)
	}
}

func TestEnsureRequest_PreservesExisting(t *testing.T) {
	t.Parallel()
	pre := id.Audit{RequestID: id.New(), IdempotencyKey: "abc"}
	ctx := id.WithAudit(context.Background(), pre)
	_, post := id.EnsureRequest(ctx)
	if post.RequestID != pre.RequestID {
		t.Errorf("EnsureRequest replaced existing RequestID: %s -> %s", pre.RequestID, post.RequestID)
	}
	if post.IdempotencyKey != pre.IdempotencyKey {
		t.Errorf("EnsureRequest dropped IdempotencyKey: %q -> %q", pre.IdempotencyKey, post.IdempotencyKey)
	}
}
