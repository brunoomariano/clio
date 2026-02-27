package ui

import (
	"testing"

	"clio/internal/index"
	"clio/internal/model"
	"clio/internal/store"

	tea "github.com/charmbracelet/bubbletea"
)

// TestModelInitUpdateView verifies Init, Update and View paths execute without panic.
// This ensures basic Bubble Tea lifecycle hooks are wired correctly.
func TestModelInitUpdateView(t *testing.T) {
	m := New(model.Config{MaxResults: 10}, store.NewNoteStore(t.TempDir()), index.NewIndex())
	if cmd := m.Init(); cmd == nil {
		t.Fatalf("expected init command")
	}
	_, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	view := m.View()
	if view == "" {
		t.Fatalf("expected view output")
	}
}
