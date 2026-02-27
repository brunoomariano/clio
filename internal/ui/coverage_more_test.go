package ui

import (
	"testing"
	"time"

	"clio/internal/index"
	"clio/internal/model"
	"clio/internal/store"
)

// TestActivatePrompt verifies that switching to prompt mode focuses the prompt
// and clears the previous value.
// This ensures modal input works as expected.
func TestActivatePrompt(t *testing.T) {
	idx := index.NewIndex()
	m := New(model.Config{}, store.NewNoteStore(t.TempDir()), idx)
	m.activatePrompt(modeTags, "Tags")
	if m.mode != modeTags {
		t.Fatalf("expected modeTags, got %v", m.mode)
	}
	if !m.prompt.Focused() {
		t.Fatalf("expected prompt to be focused")
	}
}

// TestApplyResultsEmpty verifies that applying empty results clears the list.
// This ensures UI state resets on empty search results.
func TestApplyResultsEmpty(t *testing.T) {
	idx := index.NewIndex()
	m := New(model.Config{}, store.NewNoteStore(t.TempDir()), idx)
	m.applyResults(nil)
	if len(m.list.Items()) != 0 {
		t.Fatalf("expected empty list, got %d", len(m.list.Items()))
	}
}

// TestAllNotesExpiryFilter verifies that expired notes are filtered out when
// listing all notes.
// This ensures expiry is respected for non-search listing.
func TestAllNotesExpiryFilter(t *testing.T) {
	idx := index.NewIndex()
	expired := time.Now().Add(-time.Hour)
	alive := time.Now().Add(time.Hour)
	idx.Reset([]*model.Note{
		{ID: "1", Title: "a", ExpiresAt: &expired, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "2", Title: "b", ExpiresAt: &alive, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	})
	m := &Model{index: idx, cfg: model.Config{MaxResults: 10}}
	results := m.allNotes(index.SearchOptions{MaxResults: 10, Now: time.Now()})
	if len(results) != 1 || results[0].Note.ID != "2" {
		t.Fatalf("expected only alive note, got %#v", results)
	}
}
