package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"clio/internal/model"
)

// TestPurgeExpiredSkipsUnparseable verifies that invalid notes are ignored during purge.
// This ensures purge doesn't fail on malformed files.
func TestPurgeExpiredSkipsUnparseable(t *testing.T) {
	dir := t.TempDir()
	st := NewNoteStore(dir)
	if err := st.EnsureDir(); err != nil {
		t.Fatalf("ensure dir failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bad.md"), []byte("no frontmatter"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	note := &model.Note{ID: "n1", Title: "t", CreatedAt: time.Now(), UpdatedAt: time.Now(), Body: "b"}
	if err := model.SaveNoteAtomic(filepath.Join(dir, "n1.md"), note); err != nil {
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
