package ui

import (
	"testing"
	"time"

	"clio/internal/index"
	"clio/internal/model"
)

// TestAllNotesExclude verifies that excluded tags are filtered when
// listing all notes with an empty query.
// This ensures exclude filters work outside BM25 searches.
func TestAllNotesExclude(t *testing.T) {
	idx := index.NewIndex()
	note1 := &model.Note{ID: "1", Title: "a", Tags: []string{"skip"}, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	note2 := &model.Note{ID: "2", Title: "b", Tags: []string{"keep"}, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	idx.Reset([]*model.Note{note1, note2})
	m := &Model{index: idx, cfg: model.Config{MaxResults: 10}}
	results := m.allNotes(index.SearchOptions{ExcludeTags: []string{"skip"}, MaxResults: 10, Now: time.Now()})
	if len(results) != 1 || results[0].Note.ID != "2" {
		t.Fatalf("expected excluded note to be filtered, got %#v", results)
	}
}

// TestApplyResultsSelectsFirst verifies that applying search results
// sets the list items and selects the first entry.
// This ensures the preview remains in sync with search results.
func TestApplyResultsSelectsFirst(t *testing.T) {
	idx := index.NewIndex()
	note := &model.Note{ID: "1", Title: "a", Body: "b", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	idx.Reset([]*model.Note{note})
	m := New(model.Config{MaxResults: 10}, nil, idx)
	m.applyResults([]index.SearchResult{{Note: note, Score: 1}})
	if m.list.SelectedItem() == nil {
		t.Fatalf("expected selected item")
	}
}
