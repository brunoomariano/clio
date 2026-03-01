package model

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadExistingConfig verifies that existing config values are loaded
// and search_dirs are expanded.
// This ensures user edits are respected.
func TestLoadExistingConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "clio.yaml")
	data := []byte("search_dirs:\n  - path: \"~/my-notes\"\nbm25_k1: 1.5\n")
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
	if filepath.Base(cfg.PrimarySearchDir()) != "my-notes" {
		t.Fatalf("expected expanded search dir, got %s", cfg.PrimarySearchDir())
	}
}
