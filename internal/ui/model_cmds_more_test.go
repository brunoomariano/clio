package ui

import (
	"os"
	"testing"
	"time"

	"clio/internal/index"
	"clio/internal/model"
	"clio/internal/store"
)

func withEditor(t *testing.T, value string) {
	t.Helper()
	old := os.Getenv("EDITOR")
	if err := os.Setenv("EDITOR", value); err != nil {
		t.Fatalf("set EDITOR failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Setenv("EDITOR", old) })
}

// TestCreateEditDeleteCommands verifies command execution for create, edit and delete.
// This ensures note lifecycle operations are wired to the UI layer.
func TestCreateEditDeleteCommands(t *testing.T) {
	withEditor(t, "/bin/true")
	st := store.NewNoteStore(t.TempDir())
	idx := index.NewIndex()
	m := New(model.Config{MaxResults: 10}, st, idx)

	createMsg := m.createNoteCmd()()
	if _, ok := createMsg.(editorFinishedMsg); !ok {
		t.Fatalf("expected editorFinishedMsg from create")
	}

	notes := idx.AllNotes()
	if len(notes) == 0 {
		t.Fatalf("expected note in index")
	}

	m.applyResults([]index.SearchResult{{Note: notes[0], Score: 1}})
	editMsg := m.editSelectedCmd()()
	if _, ok := editMsg.(editorFinishedMsg); !ok {
		t.Fatalf("expected editorFinishedMsg from edit")
	}

	deleteMsg := m.deleteSelectedCmd()()
	if _, ok := deleteMsg.(searchResultsMsg); !ok {
		t.Fatalf("expected searchResultsMsg from delete")
	}
}

// TestOpenEditor verifies that the editor command runs successfully.
// This ensures external editor integration is functional.
func TestOpenEditor(t *testing.T) {
	withEditor(t, "/bin/true")
	m := New(model.Config{}, store.NewNoteStore(t.TempDir()), index.NewIndex())
	path := m.store.NotePath("x")
	if err := model.SaveNoteAtomic(path, &model.Note{ID: "x", Title: "t", CreatedAt: time.Now(), UpdatedAt: time.Now()}); err != nil {
		t.Fatalf("save note failed: %v", err)
	}
	msg := m.openEditor(path)
	if res, ok := msg.(editorFinishedMsg); !ok || res.err != nil {
		t.Fatalf("expected successful editor run")
	}
}
