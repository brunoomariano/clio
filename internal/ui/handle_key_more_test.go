package ui

import (
	"os"
	"testing"
	"time"

	"clio/internal/index"
	"clio/internal/model"
	"clio/internal/store"

	tea "github.com/charmbracelet/bubbletea"
)

// TestHandleKeyActions verifies that action keys return commands without errors.
// This ensures key bindings are wired for create/edit/open/delete and tag prompts.
func TestHandleKeyActions(t *testing.T) {
	oldEditor := os.Getenv("EDITOR")
	if err := os.Setenv("EDITOR", "/bin/true"); err != nil {
		t.Fatalf("set EDITOR failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Setenv("EDITOR", oldEditor) })

	st := store.NewNoteStore(t.TempDir())
	idx := index.NewIndex()
	m := New(model.Config{MaxResults: 10}, st, idx)

	note, err := st.CreateNote("Title", "Body", nil, nil, time.Now().UTC())
	if err != nil {
		t.Fatalf("create note failed: %v", err)
	}
	idx.Upsert(note)
	m.applyResults([]index.SearchResult{{Note: note, Score: 1}})

	cases := []rune{'n', 'e', 'd', 't', 'x', 'r'}
	for _, r := range cases {
		_, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		if r == 't' || r == 'x' || r == 'r' {
			// prompt toggles or regex toggle; cmd may be nil
			continue
		}
		if cmd == nil {
			t.Fatalf("expected command for key %q", r)
		}
	}
}

// TestHandleKeyOpen verifies that Enter triggers edit when a note is selected.
// This ensures default open behavior is connected to the selection.
func TestHandleKeyOpen(t *testing.T) {
	oldEditor := os.Getenv("EDITOR")
	if err := os.Setenv("EDITOR", "/bin/true"); err != nil {
		t.Fatalf("set EDITOR failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Setenv("EDITOR", oldEditor) })

	st := store.NewNoteStore(t.TempDir())
	idx := index.NewIndex()
	m := New(model.Config{MaxResults: 10}, st, idx)

	note, _ := st.CreateNote("Title", "Body", nil, nil, time.Now().UTC())
	idx.Upsert(note)
	m.applyResults([]index.SearchResult{{Note: note, Score: 1}})

	_, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatalf("expected command on enter")
	}
}
