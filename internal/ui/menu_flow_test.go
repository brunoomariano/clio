package ui

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"clio/internal/index"
	"clio/internal/model"
	"clio/internal/store"

	tea "github.com/charmbracelet/bubbletea"
)

func TestCtrlCOpensMenuAndQuitsFromMenu(t *testing.T) {
	m := New(model.Config{MaxResults: 10}, store.NewNoteStore(t.TempDir()), index.NewIndex())

	handled, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlC})
	if !handled || cmd != nil {
		t.Fatalf("expected ctrl+c to open menu")
	}
	if m.mode != modeMenu {
		t.Fatalf("expected modeMenu, got %v", m.mode)
	}

	handled, cmd = m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlC})
	if !handled || cmd == nil {
		t.Fatalf("expected second ctrl+c in menu to quit")
	}
}

func TestCtrlCFromFileActionsReturnsToMenu(t *testing.T) {
	m := New(model.Config{MaxResults: 10}, store.NewNoteStore(t.TempDir()), index.NewIndex())
	m.mode = modeFileActions
	handled, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlC})
	if !handled || cmd != nil {
		t.Fatalf("expected ctrl+c to return to menu")
	}
	if m.mode != modeMenu {
		t.Fatalf("expected modeMenu")
	}
}

func TestApplyMenuQuitItem(t *testing.T) {
	m := New(model.Config{MaxResults: 10}, store.NewNoteStore(t.TempDir()), index.NewIndex())
	m.mode = modeMenu
	m.menu.Select(8)
	cmd := m.applyMenu()
	if cmd == nil {
		t.Fatalf("expected quit command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected QuitMsg")
	}
}

func TestApplyMenuSelectedFileOptions(t *testing.T) {
	st := store.NewNoteStore(t.TempDir())
	idx := index.NewIndex()
	m := New(model.Config{MaxResults: 10}, st, idx)
	n, _ := st.CreateNote("Title", "Body", nil, nil, time.Now().UTC())
	idx.Upsert(n)
	m.applyResults([]index.SearchResult{{Note: n, Score: 1}})

	m.mode = modeMenu
	m.menu.Select(1)
	cmd := m.applyMenu()
	if cmd != nil {
		t.Fatalf("expected nil command")
	}
	if m.mode != modeFileActions {
		t.Fatalf("expected modeFileActions, got %v", m.mode)
	}
}

func TestApplyMenuBranchCoverage(t *testing.T) {
	st := store.NewNoteStore(t.TempDir())
	idx := index.NewIndex()
	m := New(model.Config{MaxResults: 10}, st, idx)
	n, _ := st.CreateNote("Title", "Body", nil, nil, time.Now().UTC())
	idx.Upsert(n)
	m.applyResults([]index.SearchResult{{Note: n, Score: 1}})

	m.mode = modeMenu
	m.menu.Select(0)
	if cmd := m.applyMenu(); cmd == nil {
		t.Fatalf("expected new note command")
	}

	m.mode = modeMenu
	m.menu.Select(2)
	if cmd := m.applyMenu(); cmd == nil {
		t.Fatalf("expected delete command")
	}

	m.mode = modeMenu
	m.menu.Select(3)
	if cmd := m.applyMenu(); cmd != nil || m.mode != modeTags {
		t.Fatalf("expected modeTags prompt")
	}

	m.mode = modeMenu
	m.menu.Select(4)
	if cmd := m.applyMenu(); cmd != nil || m.mode != modeExpiry {
		t.Fatalf("expected modeExpiry prompt")
	}

	m.mode = modeMenu
	before := m.regex
	m.menu.Select(5)
	if cmd := m.applyMenu(); cmd == nil {
		t.Fatalf("expected search command after regex toggle")
	}
	if m.regex == before {
		t.Fatalf("expected regex toggled")
	}

	m.mode = modeMenu
	m.menu.Select(6)
	if cmd := m.applyMenu(); cmd != nil || m.mode != modeBoost {
		t.Fatalf("expected modeBoost prompt")
	}

	m.mode = modeMenu
	m.menu.Select(7)
	if cmd := m.applyMenu(); cmd != nil || m.mode != modeExclude {
		t.Fatalf("expected modeExclude prompt")
	}
}

func TestApplyFileActionMenuOpenAndBack(t *testing.T) {
	st := store.NewNoteStore(t.TempDir())
	idx := index.NewIndex()
	m := New(model.Config{MaxResults: 10}, st, idx)
	n, _ := st.CreateNote("Title", "Body", nil, nil, time.Now().UTC())
	idx.Upsert(n)
	m.applyResults([]index.SearchResult{{Note: n, Score: 1}})
	m.mode = modeFileActions

	m.fileMenu.Select(1)
	cmd := m.applyFileActionMenu()
	if cmd == nil {
		t.Fatalf("expected open command")
	}
	if _, ok := cmd().(requestOpenEditorMsg); !ok {
		t.Fatalf("expected requestOpenEditorMsg")
	}

	m.mode = modeFileActions
	m.fileMenu.Select(2)
	if cmd := m.applyFileActionMenu(); cmd != nil {
		t.Fatalf("expected nil command on back")
	}
	if m.mode != modeSearch {
		t.Fatalf("expected modeSearch")
	}
}

func TestApplyFileActionMenuCopyNoSelection(t *testing.T) {
	m := New(model.Config{MaxResults: 10}, store.NewNoteStore(t.TempDir()), index.NewIndex())
	m.mode = modeFileActions
	m.fileMenu.Select(0)
	if cmd := m.applyFileActionMenu(); cmd != nil {
		t.Fatalf("expected nil command")
	}
	if m.status == "" {
		t.Fatalf("expected status message")
	}
}

func TestApplyFileActionMenuCopySelected(t *testing.T) {
	dir := t.TempDir()
	pbcopy := filepath.Join(dir, "pbcopy")
	if err := os.WriteFile(pbcopy, []byte("#!/bin/sh\n/bin/cat >/dev/null\n"), 0o755); err != nil {
		t.Fatalf("write pbcopy failed: %v", err)
	}
	oldPath := os.Getenv("PATH")
	if err := os.Setenv("PATH", dir); err != nil {
		t.Fatalf("set PATH failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Setenv("PATH", oldPath) })

	st := store.NewNoteStore(t.TempDir())
	idx := index.NewIndex()
	m := New(model.Config{MaxResults: 10}, st, idx)
	n, _ := st.CreateNote("Title", "Body", nil, nil, time.Now().UTC())
	idx.Upsert(n)
	m.applyResults([]index.SearchResult{{Note: n, Score: 1}})

	m.mode = modeFileActions
	m.fileMenu.Select(0)
	if cmd := m.applyFileActionMenu(); cmd != nil {
		t.Fatalf("expected nil command")
	}
	if m.status != "path copied" {
		t.Fatalf("expected copied status, got %s", m.status)
	}
}

func TestRenderSearchBarAndMax(t *testing.T) {
	m := New(model.Config{MaxResults: 10}, store.NewNoteStore(t.TempDir()), index.NewIndex())
	m.width = 80
	m.regex = true
	if out := m.renderSearchBar(); out == "" {
		t.Fatalf("expected search bar output")
	}
	m.mode = modeTags
	m.prompt.SetValue("abc")
	if out := m.renderSearchBar(); out == "" {
		t.Fatalf("expected prompt search bar output")
	}
	if max(1, 2) != 2 || max(3, 2) != 3 {
		t.Fatalf("unexpected max")
	}
}

func TestApplyFileActionMenuEmptyItem(t *testing.T) {
	m := New(model.Config{MaxResults: 10}, store.NewNoteStore(t.TempDir()), index.NewIndex())
	m.mode = modeFileActions
	m.fileMenu.SetItems(nil)
	if cmd := m.applyFileActionMenu(); cmd != nil {
		t.Fatalf("expected nil command")
	}
}

func TestMenuItemTypeMethods(t *testing.T) {
	mi := menuItem("X")
	if mi.Title() != "X" || mi.Description() != "" || mi.FilterValue() != "X" {
		t.Fatalf("unexpected menu item methods")
	}
	fa := fileActionItem("Y")
	if fa.Title() != "Y" || fa.Description() != "" || fa.FilterValue() != "Y" {
		t.Fatalf("unexpected file action item methods")
	}
}

func TestPendingEditorPathAndOverlay(t *testing.T) {
	m := New(model.Config{MaxResults: 10}, store.NewNoteStore(t.TempDir()), index.NewIndex())
	_, cmd := m.Update(requestOpenEditorMsg{path: "/tmp/a.md"})
	if cmd == nil {
		t.Fatalf("expected quit command")
	}
	if got := m.PendingEditorPath(); got != "/tmp/a.md" {
		t.Fatalf("unexpected pending path: %s", got)
	}

	m.mode = modeFileActions
	out := m.renderMenuOverlay(10)
	if out == "" {
		t.Fatalf("expected rendered overlay")
	}
}

func TestCopyToClipboardEmpty(t *testing.T) {
	if err := copyToClipboard(" "); err == nil {
		t.Fatalf("expected error for empty input")
	}
}
