package index

import (
	"testing"
	"time"

	"clio/internal/model"
)

// TestBM25ScoreZeroDF verifies that unknown terms score zero.
// This ensures missing terms do not affect ranking.
func TestBM25ScoreZeroDF(t *testing.T) {
	idx := NewIndex()
	note := &model.Note{ID: "1", Body: "alpha", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	idx.Reset([]*model.Note{note})
	if score := idx.bm25Score("missing", "1", 1.2, 0.75); score != 0 {
		t.Fatalf("expected zero score, got %f", score)
	}
}

// TestBM25ScoreZeroTF verifies that zero term frequency scores zero.
// This ensures documents without the term are not ranked.
func TestBM25ScoreZeroTF(t *testing.T) {
	idx := NewIndex()
	note := &model.Note{ID: "1", Body: "alpha", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	idx.Reset([]*model.Note{note})
	if score := idx.bm25Score("alpha", "missing", 1.2, 0.75); score != 0 {
		t.Fatalf("expected zero score, got %f", score)
	}
}
