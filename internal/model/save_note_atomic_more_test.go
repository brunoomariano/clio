package model

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestSaveNoteAtomicSuccess verifies that the happy path writes and renames
// the note file successfully.
// This ensures atomic saves work under normal conditions.
func TestSaveNoteAtomicSuccess(t *testing.T) {
	dir := t.TempDir()
	note := &Note{ID: "x", Title: "t", CreatedAt: time.Now(), UpdatedAt: time.Now(), Body: "b"}
	path := filepath.Join(dir, "x.md")
	if err := SaveNoteAtomic(path, note); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file, got %v", err)
	}
}
