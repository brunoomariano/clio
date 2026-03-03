package ui

import (
	"strings"
	"testing"
	"time"

	"clio/internal/model"
)

// TestNoteItemFields verifies title, description, and filter value outputs.
// This ensures list rendering data is consistent.
func TestNoteItemFields(t *testing.T) {
	ex := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	note := &model.Note{ID: "1", Title: "Title", Tags: []string{"a", "b"}, ExpiresAt: &ex, Body: "Body line", Path: "/tmp/example.md"}
	item := noteItem{note: note}
	if item.Title() != "example.md" {
		t.Fatalf("unexpected title: %s", item.Title())
	}
	desc := item.Description()
	if !strings.Contains(desc, "Body") || !strings.Contains(desc, "/tmp/example.md") {
		t.Fatalf("unexpected description: %s", desc)
	}
	if item.FilterValue() != "Title" {
		t.Fatalf("unexpected filter value: %s", item.FilterValue())
	}
}
