package model

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadOrCreateConfigSearchDefaults verifies that global search filters are defaulted.
// This ensures config files can omit optional filter fields.
func TestLoadOrCreateConfigSearchDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "clio.yaml")
	if err := os.WriteFile(path, []byte("search_dirs:\n  - path: \"~/notes\"\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	cfg, err := LoadOrCreateConfig(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(cfg.GlobalSuffixes) == 0 {
		t.Fatalf("expected global_suffixes defaults")
	}
	if len(cfg.GlobalIgnorePaths) == 0 {
		t.Fatalf("expected global_ignore_paths defaults")
	}
}

// TestLoadOrCreateConfigLegacyNotesDir verifies notes_dir migration to search_dirs.
// This ensures old configs continue to work.
func TestLoadOrCreateConfigLegacyNotesDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "clio.yaml")
	if err := os.WriteFile(path, []byte("notes_dir: \"~/legacy\"\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	cfg, err := LoadOrCreateConfig(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if got := filepath.Base(cfg.PrimarySearchDir()); got != "legacy" {
		t.Fatalf("expected migrated dir, got %s", cfg.PrimarySearchDir())
	}
}
