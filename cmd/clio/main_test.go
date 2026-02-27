package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDefaultConfigPath verifies that the config path resolves to
// ~/.config/clio.yaml when HOME is set.
// This ensures the app uses the expected default location.
func TestDefaultConfigPath(t *testing.T) {
	oldHome := os.Getenv("HOME")
	t.Cleanup(func() { _ = os.Setenv("HOME", oldHome) })
	if err := os.Setenv("HOME", "/home/tester"); err != nil {
		t.Fatalf("set HOME failed: %v", err)
	}
	got := defaultConfigPath()
	want := filepath.Join("/home/tester", ".config", "clio.yaml")
	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}
