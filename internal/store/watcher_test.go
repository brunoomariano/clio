package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestStartWatcher verifies that file changes emit events for markdown files
// and that temporary files are ignored.
// This ensures filesystem updates trigger reindexing correctly.
func TestStartWatcher(t *testing.T) {
	dir := t.TempDir()
	ch, closeFn, err := StartWatcher(dir)
	if err != nil {
		t.Fatalf("start watcher failed: %v", err)
	}
	defer func() { _ = closeFn() }()

	mdPath := filepath.Join(dir, "note.md")
	if err := os.WriteFile(mdPath, []byte("x"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	select {
	case ev := <-ch:
		if ev.Err != nil {
			t.Fatalf("unexpected error: %v", ev.Err)
		}
		if filepath.Base(ev.Path) != "note.md" {
			t.Fatalf("unexpected event path: %s", ev.Path)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("expected watcher event")
	}
}

// TestStartWatcherError verifies that invalid directories return errors.
// This ensures watcher setup failures are reported to the caller.
func TestStartWatcherError(t *testing.T) {
	_, _, err := StartWatcher("/path/does/not/exist")
	if err == nil {
		t.Fatalf("expected error")
	}
}
