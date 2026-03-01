package model

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadOrCreateConfig verifies that a missing config file is created
// with default values and that the first search dir is expanded.
// This ensures first-run behavior is predictable and user-friendly.
func TestLoadOrCreateConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "clio.yaml")
	cfg, err := LoadOrCreateConfig(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.PrimarySearchDir() == "" {
		t.Fatalf("expected search_dirs")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected config to be created, got %v", err)
	}
}
