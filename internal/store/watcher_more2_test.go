package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestStartWatcherIgnoresTemp verifies that temporary files are ignored.
// This ensures temp writes do not trigger reindex.
func TestStartWatcherIgnoresTemp(t *testing.T) {
	dir := t.TempDir()
	ch, closeFn, err := StartWatcher(dir)
	if err != nil {
		t.Fatalf("start watcher failed: %v", err)
	}
	defer func() { _ = closeFn() }()

	tmpPath := filepath.Join(dir, ".clio_tmp_123.md")
	if err := os.WriteFile(tmpPath, []byte("x"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	select {
	case <-ch:
		t.Fatalf("did not expect event for temp file")
	case <-time.After(200 * time.Millisecond):
	}
}

// TestStartWatcherDirs verifies multi-root watcher setup succeeds.
func TestStartWatcherDirs(t *testing.T) {
	root := t.TempDir()
	a := filepath.Join(root, "a")
	b := filepath.Join(root, "b")
	if err := os.MkdirAll(a, 0o755); err != nil {
		t.Fatalf("mkdir a failed: %v", err)
	}
	if err := os.MkdirAll(b, 0o755); err != nil {
		t.Fatalf("mkdir b failed: %v", err)
	}
	ch, closeFn, err := StartWatcherDirs([]string{a, b})
	if err != nil {
		t.Fatalf("start watcher dirs failed: %v", err)
	}
	defer func() { _ = closeFn() }()
	if err := os.WriteFile(filepath.Join(a, "x.md"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatalf("expected event")
	}
}
