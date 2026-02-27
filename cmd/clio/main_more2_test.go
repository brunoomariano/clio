package main

import (
	"errors"
	"testing"
	"time"
)

// TestDefaultConfigPathError verifies that defaultConfigPath falls back when
// home resolution fails.
// This ensures a safe default path is returned on error.
func TestDefaultConfigPathError(t *testing.T) {
	oldHome := userHomeDir
	defer func() { userHomeDir = oldHome }()
	userHomeDir = func() (string, error) { return "", errors.New("boom") }

	if got := defaultConfigPath(); got != ".clio.yaml" {
		t.Fatalf("expected fallback path, got %s", got)
	}
}

// TestRealTickerMethods verifies that Chan and Stop can be called without panic.
// This ensures the ticker wrapper satisfies the interface.
func TestRealTickerMethods(t *testing.T) {
	tk := newTicker(1 * time.Millisecond)
	select {
	case <-tk.Chan():
	case <-time.After(50 * time.Millisecond):
		t.Fatalf("expected ticker to tick")
	}
	tk.Stop()
}
