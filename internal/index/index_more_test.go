package index

import (
	"testing"
	"time"

	"clio/internal/model"
)

// TestSortResultsTiebreak verifies that when scores are equal the most
// recently updated note comes first.
// This ensures stable and intuitive ordering for equal relevance.
func TestSortResultsTiebreak(t *testing.T) {
	now := time.Date(2026, 2, 27, 12, 0, 0, 0, time.UTC)
	old := now.Add(-time.Hour)
	n1 := &model.Note{ID: "1", Title: "a", UpdatedAt: old}
	n2 := &model.Note{ID: "2", Title: "b", UpdatedAt: now}
	results := []SearchResult{{Note: n1, Score: 1}, {Note: n2, Score: 1}}
	SortResults(results)
	if results[0].Note.ID != "2" {
		t.Fatalf("expected newest first, got %s", results[0].Note.ID)
	}
}

// TestSearchNoTokens verifies that a query with no tokens returns no results
// and does not panic.
// This ensures empty or symbol-only queries are handled safely.
func TestSearchNoTokens(t *testing.T) {
	idx := NewIndex()
	note := &model.Note{ID: "1", Title: "Alpha", Body: "beta", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	idx.Reset([]*model.Note{note})
	results, err := idx.Search(SearchOptions{Query: "!!!", K1: 1.2, B: 0.75, Now: time.Now()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected no results, got %#v", results)
	}
}

// TestSearchRegexBoost verifies that boost tags affect regex search results.
// This ensures boosted tags are applied consistently across search modes.
func TestSearchRegexBoost(t *testing.T) {
	idx := NewIndex()
	boosted := &model.Note{ID: "1", Title: "Alpha", Body: "beta", Tags: []string{"hot"}, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	plain := &model.Note{ID: "2", Title: "Alpha", Body: "beta", Tags: []string{"cold"}, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	idx.Reset([]*model.Note{boosted, plain})
	results, err := idx.Search(SearchOptions{Query: "Alpha", Regex: true, BoostTags: []string{"hot"}, BoostWeight: 2, Now: time.Now()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 || results[0].Note.ID != "1" {
		t.Fatalf("expected boosted note first, got %#v", results)
	}
}
