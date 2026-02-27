package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"clio/internal/model"
)

// TestEnsureDirAndDir verifies that EnsureDir creates the notes directory
// and Dir returns the configured path.
// This ensures filesystem setup is deterministic.
func TestEnsureDirAndDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "notes")
	st := NewNoteStore(dir)
	if err := st.EnsureDir(); err != nil {
		t.Fatalf("ensure dir failed: %v", err)
	}
	if st.Dir() != dir {
		t.Fatalf("unexpected dir: %s", st.Dir())
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("expected dir to exist: %v", err)
	}
}

// TestDeleteNote verifies that deleting an existing note removes the file.
// This ensures delete operations clean up on disk.
func TestDeleteNote(t *testing.T) {
	dir := t.TempDir()
	st := NewNoteStore(dir)
	note := &model.Note{ID: "n1", Title: "t", CreatedAt: time.Now(), UpdatedAt: time.Now(), Body: "b"}
	if err := model.SaveNoteAtomic(st.NotePath(note.ID), note); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	if err := st.DeleteNote(note.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if _, err := os.Stat(st.NotePath(note.ID)); !os.IsNotExist(err) {
		t.Fatalf("expected file removed")
	}
}

// TestPurgeExpiredNone verifies that purge returns empty when no notes are expired.
// This ensures purge does not delete valid notes.
func TestPurgeExpiredNone(t *testing.T) {
	dir := t.TempDir()
	st := NewNoteStore(dir)
	if err := st.EnsureDir(); err != nil {
		t.Fatalf("ensure dir failed: %v", err)
	}
	note := &model.Note{ID: "n1", Title: "t", CreatedAt: time.Now(), UpdatedAt: time.Now(), Body: "b"}
	if err := model.SaveNoteAtomic(st.NotePath(note.ID), note); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	removed, err := st.PurgeExpired(time.Now())
	if err != nil {
		t.Fatalf("purge failed: %v", err)
	}
	if len(removed) != 0 {
		t.Fatalf("expected no removals, got %#v", removed)
	}
}
