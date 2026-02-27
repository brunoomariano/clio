package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"clio/internal/model"
)

// TestNotePath verifies that note paths are resolved in the notes directory.
// This ensures filename mapping remains stable and predictable.
func TestNotePath(t *testing.T) {
	st := NewNoteStore("/tmp/notes")
	path := st.NotePath("abc")
	if path != "/tmp/notes/abc.md" {
		t.Fatalf("unexpected path: %s", path)
	}
}

// TestLoadAllIgnoresNonMarkdown verifies that only .md files are loaded.
// This ensures unrelated files do not pollute the index.
func TestLoadAllIgnoresNonMarkdown(t *testing.T) {
	dir := t.TempDir()
	st := NewNoteStore(dir)
	if err := st.EnsureDir(); err != nil {
		t.Fatalf("ensure dir failed: %v", err)
	}
	note := &model.Note{ID: "n1", Title: "t", CreatedAt: time.Now(), UpdatedAt: time.Now(), Body: "b"}
	if err := model.SaveNoteAtomic(filepath.Join(dir, "n1.md"), note); err != nil {
		t.Fatalf("save note failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ignore.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write txt failed: %v", err)
	}
	loaded, err := st.LoadAll()
	if err != nil {
		t.Fatalf("load all failed: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("expected 1 note, got %#v", loaded)
	}
}

// TestUpdateNote verifies that updates change UpdatedAt and persist to disk.
// This ensures editing notes triggers correct metadata updates.
func TestUpdateNote(t *testing.T) {
	dir := t.TempDir()
	st := NewNoteStore(dir)
	note := &model.Note{ID: "n1", Title: "t", CreatedAt: time.Now().Add(-time.Hour), UpdatedAt: time.Now().Add(-time.Hour), Body: "b"}
	if err := model.SaveNoteAtomic(st.NotePath(note.ID), note); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	old := note.UpdatedAt
	note.Body = "updated"
	if err := st.UpdateNote(note); err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if !note.UpdatedAt.After(old) {
		t.Fatalf("expected UpdatedAt to change")
	}
	data, err := os.ReadFile(st.NotePath(note.ID))
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	parsed, err := model.ParseNoteBytes(data)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if strings.TrimSpace(parsed.Body) != "updated" {
		t.Fatalf("expected updated body, got %s", parsed.Body)
	}
}
