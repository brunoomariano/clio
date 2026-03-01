package ui

import (
	"testing"
	"time"

	"clio/internal/index"
	"clio/internal/model"
	"clio/internal/store"

	tea "github.com/charmbracelet/bubbletea"
)

// TestUpdateSearchResultsSuccess verifies that search results are applied and status cleared.
// This ensures successful searches update the list and UI state.
func TestUpdateSearchResultsSuccess(t *testing.T) {
	m := New(model.Config{MaxResults: 10}, store.NewNoteStore(t.TempDir()), index.NewIndex())
	note := &model.Note{ID: "1", Title: "t", Body: "b", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	m.index.Upsert(note)
	m.status = "err"
	_, _ = m.Update(searchResultsMsg{results: []index.SearchResult{{Note: note, Score: 1}}})
	if m.status != "" {
		t.Fatalf("expected status cleared")
	}
	if len(m.list.Items()) == 0 {
		t.Fatalf("expected items")
	}
}

// TestUpdateEditorFinishedSuccess verifies that editor completion triggers refresh.
// This ensures edits are reloaded into the index.
func TestUpdateEditorFinishedSuccess(t *testing.T) {
	m := New(model.Config{MaxResults: 10}, store.NewNoteStore(t.TempDir()), index.NewIndex())
	note, _ := m.store.CreateNote("Title", "Body", nil, nil, time.Now().UTC())
	m.index.Upsert(note)
	_, cmd := m.Update(editorFinishedMsg{path: m.store.NotePath(note.ID)})
	if cmd == nil {
		t.Fatalf("expected refresh command")
	}
	msg := cmd()
	if _, ok := msg.(searchResultsMsg); !ok {
		t.Fatalf("expected searchResultsMsg")
	}
}

// TestUpdateWatcherSuccess verifies watcher messages trigger refresh.
// This ensures filesystem events update the UI.
func TestUpdateWatcherSuccess(t *testing.T) {
	m := New(model.Config{MaxResults: 10}, store.NewNoteStore(t.TempDir()), index.NewIndex())
	_, cmd := m.Update(WatcherMsg{Path: ""})
	if cmd == nil {
		t.Fatalf("expected refresh command")
	}
}

// TestHandleKeyQuit verifies that exit requires pressing CTRL+C twice.
func TestHandleKeyQuit(t *testing.T) {
	m := New(model.Config{MaxResults: 10}, store.NewNoteStore(t.TempDir()), index.NewIndex())
	_, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd != nil {
		t.Fatalf("expected first ctrl+c to only arm quit")
	}
	_, cmd = m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatalf("expected quit command on second ctrl+c")
	}
}

// TestHandleKeyOpenNoSelection verifies that open on empty list does not panic.
// This ensures safe behavior when no item is selected.
func TestHandleKeyOpenNoSelection(t *testing.T) {
	m := New(model.Config{MaxResults: 10}, store.NewNoteStore(t.TempDir()), index.NewIndex())
	_, _ = m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
}
