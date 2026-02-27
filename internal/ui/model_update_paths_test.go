package ui

import (
	"errors"
	"testing"
	"time"

	"clio/internal/index"
	"clio/internal/model"
	"clio/internal/store"

	tea "github.com/charmbracelet/bubbletea"
)

// TestUpdateErrorMessages verifies that error messages set status text.
// This ensures UI feedback appears for background failures.
func TestUpdateErrorMessages(t *testing.T) {
	m := New(model.Config{MaxResults: 10}, store.NewNoteStore(t.TempDir()), index.NewIndex())
	_, _ = m.Update(searchResultsMsg{err: errors.New("search")})
	if m.status == "" {
		t.Fatalf("expected status from search error")
	}
	_, _ = m.Update(editorFinishedMsg{err: errors.New("edit")})
	if m.status == "" {
		t.Fatalf("expected status from editor error")
	}
	_, _ = m.Update(WatcherMsg{Err: errors.New("watch")})
	if m.status == "" {
		t.Fatalf("expected status from watcher error")
	}
}

// TestUpdateWindowSize verifies that resize is applied on WindowSizeMsg.
// This ensures layout adapts to terminal size changes.
func TestUpdateWindowSize(t *testing.T) {
	m := New(model.Config{MaxResults: 10}, store.NewNoteStore(t.TempDir()), index.NewIndex())
	_, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	if m.width != 100 || m.height != 30 {
		t.Fatalf("expected size update")
	}
}

// TestUpdatePromptEnter verifies that Enter applies prompt and returns to search mode.
// This ensures prompt interactions are wired into Update.
func TestUpdatePromptEnter(t *testing.T) {
	m := New(model.Config{MaxResults: 10}, store.NewNoteStore(t.TempDir()), index.NewIndex())
	note, _ := m.store.CreateNote("Title", "Body", nil, nil, time.Now().UTC())
	m.index.Upsert(note)
	m.applyResults([]index.SearchResult{{Note: note, Score: 1}})
	m.activatePrompt(modeTags, "Tags")
	m.prompt.SetValue("x")
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.mode != modeSearch {
		t.Fatalf("expected modeSearch")
	}
}

// TestUpdateEscFromPrompt verifies that Esc cancels prompt mode.
// This ensures prompt edits can be aborted.
func TestUpdateEscFromPrompt(t *testing.T) {
	m := New(model.Config{MaxResults: 10}, store.NewNoteStore(t.TempDir()), index.NewIndex())
	m.activatePrompt(modeTags, "Tags")
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.mode != modeSearch {
		t.Fatalf("expected modeSearch")
	}
}
