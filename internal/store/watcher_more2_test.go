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
