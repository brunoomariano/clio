package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"clio/internal/model"
)

// TestCreateNote verifies that creating a note writes a markdown file
// with a valid title fallback and frontmatter.
// This ensures note creation works even without a user-provided title.
func TestCreateNote(t *testing.T) {
	dir := t.TempDir()
	st := NewNoteStore(dir)
	now := time.Date(2026, 2, 27, 10, 0, 0, 0, time.UTC)
	note, err := st.CreateNote("", "Hello body", []string{"Work"}, nil, now)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if note.Title != "Hello body" {
		t.Fatalf("expected fallback title, got %s", note.Title)
	}
	path := st.NotePath(note.ID)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to exist, got %v", err)
	}
}

// TestPurgeExpired verifies that expired notes are removed from disk
// while non-expired notes remain intact.
// This ensures automatic cleanup respects expiration metadata.
func TestPurgeExpired(t *testing.T) {
	dir := t.TempDir()
	st := NewNoteStore(dir)
	if err := st.EnsureDir(); err != nil {
		t.Fatalf("ensure dir failed: %v", err)
	}
	expired := &model.Note{
		ID:        "expired",
		Title:     "old",
		CreatedAt: time.Now().Add(-2 * time.Hour),
		UpdatedAt: time.Now().Add(-2 * time.Hour),
		ExpiresAt: ptrTime(time.Now().Add(-1 * time.Hour)),
		Body:      "x",
	}
	active := &model.Note{
		ID:        "active",
		Title:     "new",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Body:      "y",
	}
	if err := model.SaveNoteAtomic(filepath.Join(dir, "expired.md"), expired); err != nil {
		t.Fatalf("save expired failed: %v", err)
	}
	if err := model.SaveNoteAtomic(filepath.Join(dir, "active.md"), active); err != nil {
		t.Fatalf("save active failed: %v", err)
	}
	removed, err := st.PurgeExpired(time.Now())
	if err != nil {
		t.Fatalf("purge failed: %v", err)
	}
	if len(removed) != 1 || removed[0] != "expired" {
		t.Fatalf("expected expired removal, got %#v", removed)
	}
	if _, err := os.Stat(filepath.Join(dir, "active.md")); err != nil {
		t.Fatalf("expected active note to remain, got %v", err)
	}
}

// TestIsNoteFile verifies that temporary or non-markdown files are ignored.
// This ensures the watcher does not reindex temporary or unrelated files.
func TestIsNoteFile(t *testing.T) {
	if isNoteFile("/tmp/.clio_tmp_123.md") {
		t.Fatalf("expected temp file to be ignored")
	}
	if isNoteFile("/tmp/note.txt") {
		t.Fatalf("expected non-markdown to be ignored")
	}
	if !isNoteFile("/tmp/note.md") {
		t.Fatalf("expected markdown note to be accepted")
	}
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
