package model

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadExistingConfig verifies that existing config values are loaded
// and notes_dir is expanded.
// This ensures user edits are respected.
func TestLoadExistingConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "clio.yaml")
	data := []byte("notes_dir: \"~/my-notes\"\nbm25_k1: 1.5\n")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	cfg, err := LoadOrCreateConfig(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if cfg.BM25K1 != 1.5 {
		t.Fatalf("expected bm25_k1 1.5, got %v", cfg.BM25K1)
	}
	if filepath.Base(cfg.NotesDir) != "my-notes" {
		t.Fatalf("expected expanded notes_dir, got %s", cfg.NotesDir)
	}
}
