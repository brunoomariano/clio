package model

import (
	"strings"
	"testing"
	"time"
)

// TestNormalizeTags verifies that tags are lowercased, trimmed, de-duplicated
// and empty entries are removed.
// This ensures tag storage remains consistent.
func TestNormalizeTags(t *testing.T) {
	got := NormalizeTags([]string{" Work ", "work", "", "Personal"})
	if len(got) != 2 {
		t.Fatalf("expected 2 tags, got %#v", got)
	}
	if got[0] != "work" || got[1] != "personal" {
		t.Fatalf("unexpected tags: %#v", got)
	}
}

// TestRenderNote verifies that rendering produces valid frontmatter
// and that parsing the output round-trips correctly.
// This ensures serialization is stable and safe.
func TestRenderNote(t *testing.T) {
	note := &Note{
		ID:        "id1",
		Title:     "Title",
		Tags:      []string{"a", "b"},
		CreatedAt: time.Date(2026, 2, 27, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 2, 27, 12, 0, 0, 0, time.UTC),
		Body:      "Body\nLine",
	}
	data, err := RenderNote(note)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if !strings.HasPrefix(string(data), FrontmatterDelimiter) {
		t.Fatalf("expected frontmatter delimiter")
	}
	parsed, err := ParseNoteBytes(data)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if parsed.Title != note.Title || parsed.ID != note.ID {
		t.Fatalf("unexpected parsed note: %#v", parsed)
	}
}

// TestExpandPath verifies that tilde paths are expanded to the user home.
// This ensures config paths resolve correctly on Linux.
func TestExpandPath(t *testing.T) {
	home := "/home/tester"
	if got := ExpandPath("~/notes", home); got != "/home/tester/notes" {
		t.Fatalf("unexpected expand: %s", got)
	}
	if got := ExpandPath("~", home); got != "/home/tester" {
		t.Fatalf("unexpected expand: %s", got)
	}
}
