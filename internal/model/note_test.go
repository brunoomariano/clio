package model

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestParseFrontmatter verifies that YAML frontmatter is parsed into a Note
// with the expected fields and timestamps.
// This ensures metadata and body are correctly separated and typed.
func TestParseFrontmatter(t *testing.T) {
	content := strings.Join([]string{
		"---",
		"id: abc123",
		"title: Hello",
		"tags:",
		"  - work",
		"created_at: 2026-02-26T10:00:00Z",
		"updated_at: 2026-02-27T11:00:00Z",
		"expires_at: 2026-03-01T00:00:00Z",
		"---",
		"Body line 1",
		"Body line 2",
	}, "\n")

	note, err := ParseNoteBytes([]byte(content))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if note.ID != "abc123" {
		t.Fatalf("expected id, got %s", note.ID)
	}
	if note.Title != "Hello" {
		t.Fatalf("expected title, got %s", note.Title)
	}
	if len(note.Tags) != 1 || note.Tags[0] != "work" {
		t.Fatalf("expected tags, got %#v", note.Tags)
	}
	if note.Body == "" || !strings.Contains(note.Body, "Body line") {
		t.Fatalf("expected body, got %q", note.Body)
	}
	if note.ExpiresAt == nil {
		t.Fatalf("expected expires_at")
	}
}

// TestSaveNoteAtomic verifies that notes are written to a temp file and
// atomically renamed without corrupting the final file content.
// This ensures persistence remains safe on crashes or interruptions.
func TestSaveNoteAtomic(t *testing.T) {
	dir := t.TempDir()
	note := &Note{
		ID:        "abc",
		Title:     "Title",
		Tags:      []string{"tag"},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Body:      "hello",
	}
	path := filepath.Join(dir, "abc.md")
	if err := SaveNoteAtomic(path, note); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	parsed, err := ParseNoteBytes(data)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if parsed.ID != note.ID || parsed.Title != note.Title {
		t.Fatalf("unexpected saved note: %#v", parsed)
	}
}

// TestTitleFallback verifies that the fallback title uses the first non-empty
// line of the body, or a timestamp if no lines exist.
// This ensures new notes always receive a valid title.
func TestTitleFallback(t *testing.T) {
	now := time.Date(2026, 2, 27, 12, 0, 0, 0, time.UTC)
	if got := TitleFallback("\n\nHello", now); got != "Hello" {
		t.Fatalf("expected first non-empty line, got %s", got)
	}
	if got := TitleFallback("\n\n", now); got != "2026-02-27 12:00" {
		t.Fatalf("expected timestamp fallback, got %s", got)
	}
}
