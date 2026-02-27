package index

import (
	"context"
	"testing"
	"time"

	"clio/internal/model"
)

// TestTokenize verifies unicode-aware tokenization lowercases words
// and splits on non-letter/digit characters.
// This ensures consistent indexing across languages and symbols.
func TestTokenize(t *testing.T) {
	tokens := Tokenize("Hello, Go-123!")
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %#v", tokens)
	}
	if tokens[0] != "hello" || tokens[1] != "go" || tokens[2] != "123" {
		t.Fatalf("unexpected tokens: %#v", tokens)
	}
}

// TestBM25Ranking verifies that documents with higher term frequency
// receive a higher BM25 score compared to documents with lower frequency.
// This ensures ranking behaves according to the BM25 specification.
func TestBM25Ranking(t *testing.T) {
	idx := NewIndex()
	note1 := &model.Note{ID: "1", Title: "", Body: "alpha alpha alpha", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	note2 := &model.Note{ID: "2", Title: "", Body: "alpha", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	idx.Reset([]*model.Note{note1, note2})

	results, err := idx.Search(SearchOptions{Query: "alpha", K1: 1.2, B: 0.75, BoostWeight: 0, MaxResults: 10, Now: time.Now()})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(results) < 2 {
		t.Fatalf("expected results, got %#v", results)
	}
	if results[0].Note.ID != "1" {
		t.Fatalf("expected higher TF note first, got %s", results[0].Note.ID)
	}
}

// TestExcludeFilter verifies that excluded tags remove notes from candidates
// before scoring.
// This ensures hard excludes behave as specified.
func TestExcludeFilter(t *testing.T) {
	idx := NewIndex()
	note1 := &model.Note{ID: "1", Title: "", Body: "alpha", Tags: []string{"private"}, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	note2 := &model.Note{ID: "2", Title: "", Body: "alpha", Tags: []string{"public"}, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	idx.Reset([]*model.Note{note1, note2})

	results, err := idx.Search(SearchOptions{Query: "alpha", ExcludeTags: []string{"private"}, K1: 1.2, B: 0.75, BoostWeight: 0, Now: time.Now()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].Note.ID != "2" {
		t.Fatalf("expected only public note, got %#v", results)
	}
}

// TestBoostTags verifies that boost tags increase the score for matching notes.
// This ensures boosted tags influence ranking as required.
func TestBoostTags(t *testing.T) {
	idx := NewIndex()
	note1 := &model.Note{ID: "1", Title: "", Body: "alpha", Tags: []string{"boost"}, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	note2 := &model.Note{ID: "2", Title: "", Body: "alpha", Tags: []string{"other"}, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	idx.Reset([]*model.Note{note1, note2})

	results, err := idx.Search(SearchOptions{Query: "alpha", BoostTags: []string{"boost"}, BoostWeight: 2.0, K1: 1.2, B: 0.75, Now: time.Now()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].Note.ID != "1" {
		t.Fatalf("expected boosted note first, got %s", results[0].Note.ID)
	}
}

// TestRegexMode verifies that regex mode compiles patterns and matches content
// without crashing on invalid regex input.
// This ensures regex mode is safe and reports errors.
func TestRegexMode(t *testing.T) {
	idx := NewIndex()
	note := &model.Note{ID: "1", Title: "Alpha", Body: "beta", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	idx.Reset([]*model.Note{note})

	_, err := idx.Search(SearchOptions{Query: "[invalid", Regex: true, Now: time.Now()})
	if err == nil {
		t.Fatalf("expected regex error")
	}

	results, err := idx.Search(SearchOptions{Query: "Alpha", Regex: true, Now: time.Now()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected regex match, got %#v", results)
	}
}

// TestIncrementalReindex verifies that Upsert replaces previous term stats
// and Remove clears a document from the index.
// This ensures incremental updates keep the index consistent.
func TestIncrementalReindex(t *testing.T) {
	idx := NewIndex()
	note := &model.Note{ID: "1", Title: "", Body: "alpha", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	idx.Upsert(note)
	if len(idx.tf["alpha"]) != 1 {
		t.Fatalf("expected term frequency for alpha")
	}
	note.Body = "beta"
	idx.Upsert(note)
	if _, ok := idx.tf["alpha"]; ok {
		t.Fatalf("expected alpha to be removed after update")
	}
	if _, ok := idx.tf["beta"]; !ok {
		t.Fatalf("expected beta to be indexed")
	}
	idx.Remove(note.ID)
	if _, ok := idx.notesByID[note.ID]; ok {
		t.Fatalf("expected note removed")
	}
}

// TestDebounceCancel verifies that canceled contexts prevent delayed execution
// and that completed searches return results after the debounce window.
// This ensures rapid typing does not trigger obsolete searches.
func TestDebounceCancel(t *testing.T) {
	exec := DebouncedExecutor[int]{Delay: 50 * time.Millisecond}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ch := exec.Run(ctx, func(ctx context.Context) (int, error) {
		return 1, nil
	})
	if res, ok := <-ch; ok && res.Value != 0 {
		t.Fatalf("expected canceled result")
	}

	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()
	ch2 := exec.Run(ctx2, func(ctx context.Context) (int, error) {
		return 42, nil
	})
	res := <-ch2
	if res.Value != 42 {
		t.Fatalf("expected debounced result, got %d", res.Value)
	}
}
