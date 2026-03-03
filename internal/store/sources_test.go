package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"clio/internal/model"
)

// TestLoadAllWithSourceOverrides verifies per-source filters override global filters.
// This ensures each source can tune suffix and ignore behavior independently.
func TestLoadAllWithSourceOverrides(t *testing.T) {
	root := t.TempDir()
	writeDir := filepath.Join(root, "write")
	sourceA := filepath.Join(root, "docs")
	sourceB := filepath.Join(root, "logs")
	if err := os.MkdirAll(filepath.Join(sourceA, "ignore"), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.MkdirAll(sourceB, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	md := &model.Note{ID: "a1", Title: "A", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(), Body: "body"}
	if err := model.SaveNoteAtomic(filepath.Join(sourceA, "a1.md"), md); err != nil {
		t.Fatalf("save note failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceA, "ignore", "skip.txt"), []byte("skip"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceB, "picked.txt"), []byte("picked"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceB, "drop.md"), []byte("---\nid: drop\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	cfg := model.Config{
		SearchDirs: []model.SearchDir{
			{Path: sourceA},
			{Path: sourceB, Suffixes: []string{"*.txt"}, IgnorePaths: []string{"skip/*"}},
		},
		GlobalSuffixes:    []string{"*.md"},
		GlobalIgnorePaths: []string{"ignore/*"},
	}
	st := NewNoteStoreWithSources(writeDir, SourcesFromConfig(cfg))
	notes, err := st.LoadAll()
	if err != nil {
		t.Fatalf("load all failed: %v", err)
	}
	if len(notes) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(notes))
	}
}

// TestShouldIndexPath verifies path filtering by configured source patterns.
// This ensures refresh logic ignores irrelevant file events.
func TestShouldIndexPath(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "docs")
	if err := os.MkdirAll(filepath.Join(source, "tests"), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	st := NewNoteStoreWithSources(source, []Source{
		{
			Path:           source,
			SuffixPatterns: []string{"*.md", "*.txt"},
			IgnorePatterns: []string{"tests/*"},
		},
	})
	if !st.ShouldIndexPath(filepath.Join(source, "ok.txt")) {
		t.Fatalf("expected ok.txt to be indexed")
	}
	if st.ShouldIndexPath(filepath.Join(source, "tests", "skip.txt")) {
		t.Fatalf("expected tests path to be ignored")
	}
}
