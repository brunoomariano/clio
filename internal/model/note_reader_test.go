package model

import (
	"strings"
	"testing"
	"time"
)

// TestParseNoteReader verifies ParseNote reads from an io.Reader and
// produces a valid Note.
// This ensures reader-based parsing works for streams.
func TestParseNoteReader(t *testing.T) {
	content := "---\nid: abc\ntitle: T\ntags: []\ncreated_at: 2026-02-26T10:00:00Z\nupdated_at: 2026-02-27T11:00:00Z\n---\nbody\n"
	note, err := ParseNote(strings.NewReader(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if note.ID != "abc" || note.Title != "T" {
		t.Fatalf("unexpected note: %#v", note)
	}
}

// TestRenderNoteWithExpiryRoundTrip verifies expires_at survives a render/parse cycle.
// This ensures optional expiry metadata is stable.
func TestRenderNoteWithExpiryRoundTrip(t *testing.T) {
	ex := time.Now().Add(time.Hour).UTC()
	note := &Note{ID: "x", Title: "t", CreatedAt: time.Now(), UpdatedAt: time.Now(), ExpiresAt: &ex, Body: "b"}
	data, err := RenderNote(note)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	parsed, err := ParseNoteBytes(data)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if parsed.ExpiresAt == nil {
		t.Fatalf("expected expires_at")
	}
}
