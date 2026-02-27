package model

import (
	"testing"
	"time"
)

// TestParseNoteMissingFrontmatter verifies that missing frontmatter is rejected.
// This ensures invalid notes are not silently accepted.
func TestParseNoteMissingFrontmatter(t *testing.T) {
	_, err := ParseNoteBytes([]byte("hello"))
	if err == nil {
		t.Fatalf("expected error for missing frontmatter")
	}
}

// TestParseNoteBadDates verifies that invalid RFC3339 timestamps are rejected.
// This ensures time metadata is validated.
func TestParseNoteBadDates(t *testing.T) {
	content := "---\nid: a\ntitle: t\ntags: []\ncreated_at: bad\nupdated_at: bad\n---\nbody\n"
	_, err := ParseNoteBytes([]byte(content))
	if err == nil {
		t.Fatalf("expected error for bad timestamps")
	}
}

// TestNewIDLength verifies that generated IDs are non-empty and stable length.
// This ensures filenames and IDs remain compact and consistent.
func TestNewIDLength(t *testing.T) {
	id, err := NewID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(id) != 10 {
		t.Fatalf("expected 10-char id, got %d", len(id))
	}
}

// TestRenderNoteExpiresAt verifies that expires_at is rendered when set.
// This ensures optional expiry metadata is preserved.
func TestRenderNoteExpiresAt(t *testing.T) {
	ex := time.Date(2026, 2, 28, 10, 0, 0, 0, time.UTC)
	note := &Note{
		ID:        "id2",
		Title:     "t",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ExpiresAt: &ex,
		Body:      "b",
	}
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
