package model

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadOrCreateConfigWithExistingAndDefaults verifies that defaults are preserved
// when fields are omitted in the config.
// This ensures backward-compatible config parsing.
func TestLoadOrCreateConfigWithExistingAndDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "clio.yaml")
	if err := os.WriteFile(path, []byte("search_dirs:\n  - path: \"~/notes\"\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	cfg, err := LoadOrCreateConfig(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if cfg.BM25K1 == 0 || cfg.MaxResults == 0 {
		t.Fatalf("expected defaults, got %#v", cfg)
	}
}
