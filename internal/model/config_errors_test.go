package model

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadOrCreateConfigBadYAML verifies that invalid YAML returns an error.
// This ensures configuration parsing is strict.
func TestLoadOrCreateConfigBadYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "clio.yaml")
	if err := os.WriteFile(path, []byte("notes_dir: ["), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	if _, err := LoadOrCreateConfig(path); err == nil {
		t.Fatalf("expected error")
	}
}

// TestLoadOrCreateConfigCreateError verifies that create errors are returned.
// This ensures config creation failures are surfaced.
func TestLoadOrCreateConfigCreateError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "clio.yaml")
	if err := os.WriteFile(filepath.Join(dir, "subdir"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	if _, err := LoadOrCreateConfig(path); err == nil {
		t.Fatalf("expected error")
	}
}
