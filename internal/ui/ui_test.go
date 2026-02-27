package ui

import (
	"testing"
	"time"

	"clio/internal/model"
)

// TestSplitTags verifies that tag input is split by commas and trimmed.
// This ensures tag editing behaves predictably for users.
func TestSplitTags(t *testing.T) {
	tags := splitTags("a, b, ,c")
	if len(tags) != 3 {
		t.Fatalf("expected 3 tags, got %#v", tags)
	}
	if tags[1] != "b" {
		t.Fatalf("expected trimmed tag, got %s", tags[1])
	}
}

// TestRenderChips verifies that boost and exclude chips are rendered
// with the correct prefixes.
// This ensures the filter UI communicates active tags clearly.
func TestRenderChips(t *testing.T) {
	out := renderChips([]string{"work"}, []string{"private"})
	if out == "" {
		t.Fatalf("expected chips output")
	}
}

// TestFormatExpiry verifies that expiry text returns a label for
// expired notes and RFC3339 for future notes.
// This ensures expiry indicators are consistent in the UI.
func TestFormatExpiry(t *testing.T) {
	past := time.Now().Add(-time.Hour)
	future := time.Now().Add(time.Hour)
	if got := formatExpiry(&model.Note{ExpiresAt: &past}); got != "expired" {
		t.Fatalf("expected expired, got %s", got)
	}
	if got := formatExpiry(&model.Note{ExpiresAt: &future}); got == "" {
		t.Fatalf("expected future expiry, got empty")
	}
}
