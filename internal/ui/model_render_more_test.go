package ui

import (
	"testing"

	"clio/internal/index"
	"clio/internal/model"
	"clio/internal/store"
)

// TestRenderLeftRightFooter verifies rendering of layout sections.
// This ensures the UI can build each column and footer without errors.
func TestRenderLeftRightFooter(t *testing.T) {
	m := New(model.Config{MaxResults: 10}, store.NewNoteStore(t.TempDir()), index.NewIndex())
	m.width = 100
	m.height = 30
	m.resize()
	left := m.renderLeft()
	right := m.renderRight()
	footer := m.renderFooter()
	if left == "" || right == "" || footer == "" {
		t.Fatalf("expected rendered sections")
	}

	m.activatePrompt(modeTags, "Tags")
	leftPrompt := m.renderLeft()
	if leftPrompt == "" {
		t.Fatalf("expected rendered prompt left section")
	}
}
