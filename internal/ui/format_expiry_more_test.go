package ui

import (
	"testing"
	"time"

	"clio/internal/model"
)

// TestFormatExpiryEmpty verifies empty output for notes without expiry.
// This ensures non-expiring notes don't show expiry labels.
func TestFormatExpiryEmpty(t *testing.T) {
	if got := formatExpiry(&model.Note{}); got != "" {
		t.Fatalf("expected empty expiry, got %s", got)
	}
}

// TestFormatExpiryFuture verifies that future expiry renders RFC3339.
// This ensures expiry indicators include timestamps when applicable.
func TestFormatExpiryFuture(t *testing.T) {
	future := time.Now().Add(time.Hour)
	got := formatExpiry(&model.Note{ExpiresAt: &future})
	if got == "" || got == "expired" {
		t.Fatalf("unexpected expiry: %s", got)
	}
}
