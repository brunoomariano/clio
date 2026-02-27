package ui

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"clio/internal/index"
	"clio/internal/model"
	"clio/internal/store"
)

// TestRefreshFromFileNonMarkdown verifies that non-md paths are ignored.
// This ensures refresh does not attempt to parse unrelated files.
func TestRefreshFromFileNonMarkdown(t *testing.T) {
	m := New(model.Config{MaxResults: 10}, store.NewNoteStore(t.TempDir()), index.NewIndex())
	cmd := m.refreshFromFile("/tmp/file.txt")
	if cmd() != nil {
		t.Fatalf("expected nil message")
	}
}

// TestRefreshFromFileParseError verifies that invalid markdown returns an error message.
// This ensures parse failures are surfaced to the UI.
func TestRefreshFromFileParseError(t *testing.T) {
	st := store.NewNoteStore(t.TempDir())
	m := New(model.Config{MaxResults: 10}, st, index.NewIndex())
	path := filepath.Join(st.Dir(), "bad.md")
	if err := os.WriteFile(path, []byte("invalid"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	msg := m.refreshFromFile(path)()
	if _, ok := msg.(editorFinishedMsg); !ok {
		t.Fatalf("expected editorFinishedMsg")
	}
}

// TestDeleteSelectedCmdNoSelection verifies delete does nothing without selection.
// This ensures delete is safe when the list is empty.
func TestDeleteSelectedCmdNoSelection(t *testing.T) {
	m := New(model.Config{MaxResults: 10}, store.NewNoteStore(t.TempDir()), index.NewIndex())
	if cmd := m.deleteSelectedCmd(); cmd != nil {
		t.Fatalf("expected nil command")
	}
}

// TestCreateNoteCmdError verifies errors are returned when note creation fails.
// This ensures create errors are propagated to the UI.
func TestCreateNoteCmdError(t *testing.T) {
	st := store.NewNoteStore(t.TempDir())
	_ = os.Chmod(st.Dir(), 0o500)
	m := New(model.Config{MaxResults: 10}, st, index.NewIndex())
	msg := m.createNoteCmd()()
	if _, ok := msg.(editorFinishedMsg); !ok {
		t.Fatalf("expected editorFinishedMsg")
	}
	_ = os.Chmod(st.Dir(), 0o755)
}

// TestEditSelectedCmdNoSelection verifies edit returns nil when no selection exists.
// This ensures edit is safe on empty lists.
func TestEditSelectedCmdNoSelection(t *testing.T) {
	m := New(model.Config{MaxResults: 10}, store.NewNoteStore(t.TempDir()), index.NewIndex())
	if cmd := m.editSelectedCmd(); cmd != nil {
		t.Fatalf("expected nil command")
	}
}

// TestDeleteSelectedCmdSuccess verifies delete removes notes and updates results.
// This ensures delete triggers re-search.
func TestDeleteSelectedCmdSuccess(t *testing.T) {
	st := store.NewNoteStore(t.TempDir())
	idx := index.NewIndex()
	m := New(model.Config{MaxResults: 10}, st, idx)
	note, _ := st.CreateNote("Title", "Body", nil, nil, time.Now().UTC())
	idx.Upsert(note)
	m.applyResults([]index.SearchResult{{Note: note, Score: 1}})
	msg := m.deleteSelectedCmd()()
	if _, ok := msg.(searchResultsMsg); !ok {
		t.Fatalf("expected searchResultsMsg")
	}
}
