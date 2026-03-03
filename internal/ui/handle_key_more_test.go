package ui

import (
	"testing"
	"time"

	"clio/internal/index"
	"clio/internal/model"
	"clio/internal/store"

	tea "github.com/charmbracelet/bubbletea"
)

// TestHandleKeyActions verifies hotkeys are ignored in favor of menu actions only.
func TestHandleKeyActions(t *testing.T) {
	st := store.NewNoteStore(t.TempDir())
	idx := index.NewIndex()
	m := New(model.Config{MaxResults: 10}, st, idx)

	note, err := st.CreateNote("Title", "Body", nil, nil, time.Now().UTC())
	if err != nil {
		t.Fatalf("create note failed: %v", err)
	}
	idx.Upsert(note)
	m.applyResults([]index.SearchResult{{Note: note, Score: 1}})

	cases := []rune{'n', 'e', 'd', 't', 'x', 'r', '+', '-', 'q'}
	for _, r := range cases {
		handled, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		if handled || cmd != nil {
			t.Fatalf("expected hotkey %q to be ignored", r)
		}
	}
}

// TestHandleKeyOpen verifies that Enter in search mode opens the file-action menu.
func TestHandleKeyOpen(t *testing.T) {
	st := store.NewNoteStore(t.TempDir())
	idx := index.NewIndex()
	m := New(model.Config{MaxResults: 10}, st, idx)

	note, _ := st.CreateNote("Title", "Body", nil, nil, time.Now().UTC())
	idx.Upsert(note)
	m.applyResults([]index.SearchResult{{Note: note, Score: 1}})

	handled, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if !handled || cmd != nil {
		t.Fatalf("expected handled enter with no immediate command")
	}
	if m.mode != modeFileActions {
		t.Fatalf("expected modeFileActions, got %v", m.mode)
	}
}
