package index

import (
	"testing"
	"time"

	"clio/internal/model"
)

// TestIndexAccessors verifies NotesCount, Get, and AllNotes return
// consistent data after indexing.
// This ensures basic index accessors behave correctly.
func TestIndexAccessors(t *testing.T) {
	idx := NewIndex()
	n1 := &model.Note{ID: "1", Title: "A", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	n2 := &model.Note{ID: "2", Title: "B", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	idx.Reset([]*model.Note{n1, n2})
	if idx.NotesCount() != 2 {
		t.Fatalf("expected 2 notes, got %d", idx.NotesCount())
	}
	if note, ok := idx.Get("1"); !ok || note.ID != "1" {
		t.Fatalf("expected to get note 1")
	}
	all := idx.AllNotes()
	if len(all) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(all))
	}
}

// TestIsExpired verifies expiry detection for past and future timestamps.
// This ensures expired notes are filtered consistently.
func TestIsExpired(t *testing.T) {
	past := time.Now().Add(-time.Hour)
	future := time.Now().Add(time.Hour)
	if !isExpired(&model.Note{ExpiresAt: &past}, time.Now()) {
		t.Fatalf("expected expired note")
	}
	if isExpired(&model.Note{ExpiresAt: &future}, time.Now()) {
		t.Fatalf("expected non-expired note")
	}
}
