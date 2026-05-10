package pgstore_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/manuelibar/workbench/internal/domain"
	"github.com/manuelibar/workbench/internal/pgstore"
)

func TestStore_NoteLifecycle(t *testing.T) {
	s, ctx, _ := openForIntegration(t)
	u, err := s.EnsureSingletonUser(ctx, "")
	if err != nil {
		t.Fatalf("EnsureSingletonUser: %v", err)
	}

	body := "first note " + uuid.NewString()
	added, err := s.AddNote(ctx, domain.Note{
		UserID: u.ID,
		BodyMD: body,
		Tags:   []string{"intent:test", "phase:4"},
	})
	if err != nil {
		t.Fatalf("AddNote: %v", err)
	}
	if added.ID == uuid.Nil {
		t.Fatal("note id is zero")
	}

	got, err := s.GetNote(ctx, u.ID, added.ID)
	if err != nil {
		t.Fatalf("GetNote: %v", err)
	}
	if got.BodyMD != body {
		t.Errorf("body roundtrip differs")
	}

	tags := []string{"phase:4"}
	notes, err := s.ListNotes(ctx, u.ID, pgstore.ListNotesFilter{Tag: "phase:4"})
	if err != nil {
		t.Fatalf("ListNotes: %v", err)
	}
	if len(notes) == 0 {
		t.Fatalf("ListNotes returned 0 with filter %v", tags)
	}

	hits, err := s.SearchNotes(ctx, u.ID, "first note", 0)
	if err != nil {
		t.Fatalf("SearchNotes: %v", err)
	}
	found := false
	for _, n := range hits {
		if n.ID == added.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("SearchNotes did not return the inserted note")
	}

	newBody := body + " (updated)"
	updated, err := s.UpdateNote(ctx, u.ID, added.ID, pgstore.UpdateNoteFields{
		BodyMD: &newBody,
		Tags:   &[]string{"phase:4", "updated"},
	})
	if err != nil {
		t.Fatalf("UpdateNote: %v", err)
	}
	if updated.BodyMD != newBody {
		t.Errorf("UpdateNote did not change body")
	}
	if !strings.Contains(strings.Join(updated.Tags, ","), "updated") {
		t.Errorf("UpdateNote tags missing 'updated': %v", updated.Tags)
	}

	if err := s.DeleteNote(ctx, u.ID, added.ID); err != nil {
		t.Fatalf("DeleteNote: %v", err)
	}
	if _, err := s.GetNote(ctx, u.ID, added.ID); err == nil {
		t.Error("expected ErrNotFound after delete")
	}
}

func TestStore_NoteIdempotency(t *testing.T) {
	s, ctx, _ := openForIntegration(t)
	u, err := s.EnsureSingletonUser(ctx, "")
	if err != nil {
		t.Fatalf("EnsureSingletonUser: %v", err)
	}

	key := "phase4-idemp-" + uuid.NewString()
	first, err := s.AddNote(ctx, domain.Note{
		UserID:         u.ID,
		BodyMD:         "first attempt",
		IdempotencyKey: key,
	})
	if err != nil {
		t.Fatalf("AddNote first: %v", err)
	}
	second, err := s.AddNote(ctx, domain.Note{
		UserID:         u.ID,
		BodyMD:         "second attempt with same key",
		IdempotencyKey: key,
	})
	if err != nil {
		t.Fatalf("AddNote second: %v", err)
	}
	if first.ID != second.ID {
		t.Errorf("idempotency replay returned different id: %s vs %s", first.ID, second.ID)
	}
	if second.BodyMD != "first attempt" {
		t.Errorf("idempotency replay returned mutated body: %q", second.BodyMD)
	}
}
