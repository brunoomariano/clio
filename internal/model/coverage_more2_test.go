package model

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("read error") }

// TestParseNoteReaderError verifies that reader errors are propagated.
// This ensures ParseNote reports IO failures.
func TestParseNoteReaderError(t *testing.T) {
	_, err := ParseNote(errReader{})
	if err == nil {
		t.Fatalf("expected error")
	}
}

// TestExpandPathTildeUser verifies that tilde-prefixed paths are expanded.
// This ensures non-standard tilde paths are handled.
func TestExpandPathTildeUser(t *testing.T) {
	home := "/home/tester"
	if got := ExpandPath("~other", home); got != "/home/tester/other" {
		t.Fatalf("unexpected expand: %s", got)
	}
}

// TestLoadOrCreateConfigReadError verifies that read errors are surfaced.
// This ensures invalid config paths do not silently succeed.
func TestLoadOrCreateConfigReadError(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "clio.yaml"), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	_, err := LoadOrCreateConfig(filepath.Join(dir, "clio.yaml"))
	if err == nil {
		t.Fatalf("expected error")
	}
}

// TestSaveNoteAtomicCreateTempError verifies that failures to create temp files
// are reported to the caller.
// This ensures atomic saves fail safely when directories are not writable.
func TestSaveNoteAtomicCreateTempError(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatalf("chmod failed: %v", err)
	}
	note := &Note{ID: "x", Title: "t", CreatedAt: time.Now(), UpdatedAt: time.Now(), Body: "b"}
	err := SaveNoteAtomic(filepath.Join(dir, "x.md"), note)
	if err == nil {
		t.Fatalf("expected error")
	}
}

// TestSaveNoteAtomicRenameError verifies that rename failures are surfaced.
// This ensures atomic save errors are not swallowed.
func TestSaveNoteAtomicRenameError(t *testing.T) {
	dir := t.TempDir()
	note := &Note{ID: "x", Title: "t", CreatedAt: time.Now(), UpdatedAt: time.Now(), Body: "b"}
	path := filepath.Join(dir, "targetdir")
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := SaveNoteAtomic(path, note); err == nil {
		t.Fatalf("expected error")
	}
}

// TestParseNoteUsesReader verifies that ParseNote consumes the provided reader.
// This ensures stream parsing works correctly for different readers.
func TestParseNoteUsesReader(t *testing.T) {
	content := "---\nid: a\ntitle: t\ntags: []\ncreated_at: 2026-02-26T10:00:00Z\nupdated_at: 2026-02-27T11:00:00Z\n---\nbody\n"
	_, err := ParseNote(strings.NewReader(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
