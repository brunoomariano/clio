package model

import (
	"testing"
	"time"
)

// TestParseNoteNoExpiry verifies that notes without expires_at parse successfully.
// This ensures optional fields are handled correctly.
func TestParseNoteNoExpiry(t *testing.T) {
	content := "---\nid: a\ntitle: t\ntags: []\ncreated_at: 2026-02-26T10:00:00Z\nupdated_at: 2026-02-27T11:00:00Z\nexpires_at: \n---\nbody\n"
	note, err := ParseNoteBytes([]byte(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if note.ExpiresAt != nil {
		t.Fatalf("expected nil expires_at")
	}
}

// TestRenderNoteEmptyBody verifies that empty bodies still render valid notes.
// This ensures new notes can be created with empty content.
func TestRenderNoteEmptyBody(t *testing.T) {
	note := &Note{ID: "id", Title: "t", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	data, err := RenderNote(note)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatalf("expected output")
	}
}
