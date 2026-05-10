package pgstore_test

import (
	"testing"
	"time"

	"github.com/manuelibar/workbench/internal/domain"
)

func TestStore_RecordEventAndRecent(t *testing.T) {
	s, ctx, _ := openForIntegration(t)
	u, err := s.EnsureSingletonUser(ctx, "")
	if err != nil {
		t.Fatalf("EnsureSingletonUser: %v", err)
	}
	w, err := s.EnsureOpenWorkSession(ctx, u.ID, "")
	if err != nil {
		t.Fatalf("EnsureOpenWorkSession: %v", err)
	}

	before := time.Now().Add(-time.Second)
	e1, err := s.RecordEvent(ctx, domain.Event{
		WorkSessionID: w.ID,
		Type:          "tool.call",
		SubjectKind:   "tool",
		SubjectID:     "refresh",
		Payload:       map[string]any{"ok": true},
	})
	if err != nil {
		t.Fatalf("RecordEvent: %v", err)
	}
	if e1.OccurredAt.Before(before) {
		t.Errorf("OccurredAt = %s, expected >= %s", e1.OccurredAt, before)
	}

	if _, err := s.RecordEvent(ctx, domain.Event{
		WorkSessionID: w.ID,
		Type:          "tool.call",
		SubjectKind:   "tool",
		SubjectID:     "ask",
	}); err != nil {
		t.Fatalf("RecordEvent second: %v", err)
	}

	got, err := s.RecentEvents(ctx, w.ID, 5)
	if err != nil {
		t.Fatalf("RecentEvents: %v", err)
	}
	if len(got) < 2 {
		t.Fatalf("expected at least 2 events, got %d", len(got))
	}
	if got[0].OccurredAt.Before(got[1].OccurredAt) {
		t.Errorf("expected most-recent-first ordering: %s vs %s", got[0].OccurredAt, got[1].OccurredAt)
	}
}
