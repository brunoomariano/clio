package store

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"clio/internal/model"
)

// TestLoadNoteMissing verifies that loading a missing note returns an error.
// This ensures callers can distinguish absent files from parse errors.
func TestLoadNoteMissing(t *testing.T) {
	st := NewNoteStore(t.TempDir())
	_, err := st.LoadNote(filepath.Join(st.Dir(), "missing.md"))
	if err == nil {
		t.Fatalf("expected error for missing note")
	}
}

// TestDeleteNoteMissing verifies that deleting a missing note returns ErrNoteNotFound.
// This ensures delete operations report expected error semantics.
func TestDeleteNoteMissing(t *testing.T) {
	st := NewNoteStore(t.TempDir())
	err := st.DeleteNote("missing")
	if !errors.Is(err, ErrNoteNotFound) {
		t.Fatalf("expected ErrNoteNotFound, got %v", err)
	}
}

// TestLoadAllSkipsInvalid verifies that invalid markdown is skipped on load.
// This ensures corrupt notes do not break loading other notes.
func TestLoadAllSkipsInvalid(t *testing.T) {
	dir := t.TempDir()
	st := NewNoteStore(dir)
	if err := st.EnsureDir(); err != nil {
		t.Fatalf("ensure dir failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bad.md"), []byte("not frontmatter"), 0o644); err != nil {
		t.Fatalf("write bad note failed: %v", err)
	}
	note := &model.Note{ID: "good", Title: "t", CreatedAt: time.Now(), UpdatedAt: time.Now(), Body: "b"}
	if err := model.SaveNoteAtomic(filepath.Join(dir, "good.md"), note); err != nil {
		t.Fatalf("save good note failed: %v", err)
	}
	loaded, err := st.LoadAll()
	if err != nil {
		t.Fatalf("load all failed: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("expected 1 valid note, got %#v", loaded)
	}
}
