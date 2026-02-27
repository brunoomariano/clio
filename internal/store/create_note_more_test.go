package store

import (
	"testing"
	"time"
)

// TestCreateNoteWithTitle verifies that provided titles are preserved.
// This ensures manual titles are not overwritten.
func TestCreateNoteWithTitle(t *testing.T) {
	st := NewNoteStore(t.TempDir())
	now := time.Now().UTC()
	note, err := st.CreateNote("Custom", "Body", nil, nil, now)
	if err != nil {
		t.Fatalf("create note failed: %v", err)
	}
	if note.Title != "Custom" {
		t.Fatalf("expected custom title, got %s", note.Title)
	}
}
