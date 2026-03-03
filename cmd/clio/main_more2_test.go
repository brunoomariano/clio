package main

import (
	"errors"
	"testing"
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
