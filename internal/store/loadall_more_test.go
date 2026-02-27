package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"clio/internal/model"
)

// TestLoadAllMissingDir verifies that LoadAll creates the directory when missing.
// This ensures first-run load works without pre-existing folders.
func TestLoadAllMissingDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "notes")
	st := NewNoteStore(dir)
	_, err := st.LoadAll()
	if err != nil {
		t.Fatalf("load all failed: %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("expected dir to exist: %v", err)
	}
}

// TestLoadAllLoadsMultiple verifies that multiple notes are loaded.
// This ensures batch loading collects all valid notes.
func TestLoadAllLoadsMultiple(t *testing.T) {
	dir := t.TempDir()
	st := NewNoteStore(dir)
	if err := st.EnsureDir(); err != nil {
		t.Fatalf("ensure dir failed: %v", err)
	}
	for i := 0; i < 2; i++ {
		note := &model.Note{ID: string(rune('a' + i)), Title: "t", CreatedAt: time.Now(), UpdatedAt: time.Now(), Body: "b"}
		if err := model.SaveNoteAtomic(filepath.Join(dir, note.ID+".md"), note); err != nil {
			t.Fatalf("save failed: %v", err)
		}
	}
	loaded, err := st.LoadAll()
	if err != nil {
		t.Fatalf("load all failed: %v", err)
	}
	if len(loaded) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(loaded))
	}
}
