package ui

import (
	"strings"
	"testing"
	"time"

	"clio/internal/index"
	"clio/internal/model"
	"clio/internal/store"

	tea "github.com/charmbracelet/bubbletea"
)

func newTestModel(t *testing.T) *Model {
	t.Helper()
	st := store.NewNoteStore(t.TempDir())
	idx := index.NewIndex()
	m := New(model.Config{MaxResults: 50, DebounceMS: 1, BoostWeight: 2, BM25K1: 1.2, BM25B: 0.75}, st, idx)
	return m
}

// TestHandleKeyFocusSearch verifies that ESC returns to search mode from prompts.
func TestHandleKeyFocusSearch(t *testing.T) {
	m := newTestModel(t)
	m.mode = modeTags
	m.prompt.Focus()
	handled, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	if !handled {
		t.Fatalf("expected handled key")
	}
	if m.mode != modeSearch {
		t.Fatalf("expected modeSearch, got %v", m.mode)
	}
	if !m.search.Focused() {
		t.Fatalf("expected search to be focused")
	}
}

// TestHandleKeyRegexToggle verifies that regex mode toggles via menu action.
func TestHandleKeyRegexToggle(t *testing.T) {
	m := newTestModel(t)
	m.regex = false
	m.mode = modeMenu
	m.menu.Select(5)
	m.applyMenu()
	if !m.regex {
		t.Fatalf("expected regex to be enabled")
	}
}

// TestHandleKeyPromptModes verifies that menu entries activate prompt modes.
func TestHandleKeyPromptModes(t *testing.T) {
	m := newTestModel(t)
	m.mode = modeMenu
	m.menu.Select(6)
	m.applyMenu()
	if m.mode != modeBoost {
		t.Fatalf("expected modeBoost, got %v", m.mode)
	}
	m.mode = modeMenu
	m.menu.Select(7)
	m.applyMenu()
	if m.mode != modeExclude {
		t.Fatalf("expected modeExclude, got %v", m.mode)
	}
}

// TestApplyPromptBoostExclude verifies that boost/exclude inputs are stored
// and reset back to search mode.
// This ensures filter chips reflect user input.
func TestApplyPromptBoostExclude(t *testing.T) {
	m := newTestModel(t)
	m.mode = modeBoost
	m.prompt.SetValue("work")
	m.applyPrompt()
	if len(m.boost) != 1 || m.boost[0] != "work" {
		t.Fatalf("expected boost tag stored")
	}
	m.mode = modeExclude
	m.prompt.SetValue("private")
	m.applyPrompt()
	if len(m.exclude) != 1 || m.exclude[0] != "private" {
		t.Fatalf("expected exclude tag stored")
	}
	if m.mode != modeSearch {
		t.Fatalf("expected modeSearch")
	}
}

// TestApplyPromptTags verifies that tag edits are persisted to the note file.
// This ensures metadata editing updates both store and index.
func TestApplyPromptTags(t *testing.T) {
	m := newTestModel(t)
	note, err := m.store.CreateNote("Title", "Body", []string{"old"}, nil, time.Now().UTC())
	if err != nil {
		t.Fatalf("create note failed: %v", err)
	}
	m.index.Upsert(note)
	m.applyResults([]index.SearchResult{{Note: note, Score: 1}})
	m.mode = modeTags
	m.prompt.SetValue("a, b")
	m.applyPrompt()
	if len(note.Tags) != 2 || note.Tags[0] != "a" {
		t.Fatalf("expected updated tags, got %#v", note.Tags)
	}
}

// TestApplyPromptExpiry verifies that expiry input is parsed and stored.
// This ensures expiry edits update the note metadata.
func TestApplyPromptExpiry(t *testing.T) {
	m := newTestModel(t)
	note, err := m.store.CreateNote("Title", "Body", nil, nil, time.Now().UTC())
	if err != nil {
		t.Fatalf("create note failed: %v", err)
	}
	m.index.Upsert(note)
	m.applyResults([]index.SearchResult{{Note: note, Score: 1}})
	ex := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	m.mode = modeExpiry
	m.prompt.SetValue(ex)
	m.applyPrompt()
	if note.ExpiresAt == nil {
		t.Fatalf("expected expires_at to be set")
	}
}

// TestRunSearchRegexError verifies that invalid regex returns an error message.
// This ensures regex failures are surfaced without crashing.
func TestRunSearchRegexError(t *testing.T) {
	m := newTestModel(t)
	m.regex = true
	m.search.SetValue("[")
	cmd := m.runSearch()
	msg := cmd()
	res, ok := msg.(searchResultsMsg)
	if !ok {
		t.Fatalf("expected searchResultsMsg, got %T", msg)
	}
	if res.err == nil {
		t.Fatalf("expected regex error")
	}
}

// TestRunSearchEmptyQuery verifies that empty query returns all notes
// and respects exclude tags.
// This ensures the default listing matches filters.
func TestRunSearchEmptyQuery(t *testing.T) {
	m := newTestModel(t)
	m.exclude = []string{"private"}
	pub, _ := m.store.CreateNote("Public", "Body", []string{"public"}, nil, time.Now().UTC())
	priv, _ := m.store.CreateNote("Private", "Body", []string{"private"}, nil, time.Now().UTC())
	m.index.Reset([]*model.Note{pub, priv})
	m.search.SetValue("")
	cmd := m.runSearch()
	msg := cmd()
	res := msg.(searchResultsMsg)
	if len(res.results) != 1 || res.results[0].Note.ID != pub.ID {
		t.Fatalf("expected only public note, got %#v", res.results)
	}
}

// TestRefreshFromFile verifies that reindexing a modified file updates the index.
// This ensures the watcher keeps the index in sync with disk.
func TestRefreshFromFile(t *testing.T) {
	m := newTestModel(t)
	note, err := m.store.CreateNote("Title", "Body", nil, nil, time.Now().UTC())
	if err != nil {
		t.Fatalf("create note failed: %v", err)
	}
	m.index.Upsert(note)
	path := m.store.NotePath(note.ID)
	cmd := m.refreshFromFile(path)
	msg := cmd()
	res, ok := msg.(searchResultsMsg)
	if !ok || len(res.results) == 0 {
		t.Fatalf("expected search results")
	}
}

// TestRefreshFromFileMissing verifies that missing files remove notes from the index.
// This ensures deleted notes are reflected in search results.
func TestRefreshFromFileMissing(t *testing.T) {
	m := newTestModel(t)
	note, _ := m.store.CreateNote("Title", "Body", nil, nil, time.Now().UTC())
	m.index.Upsert(note)
	path := m.store.NotePath(note.ID)
	_ = m.store.DeleteNote(note.ID)
	cmd := m.refreshFromFile(path)
	msg := cmd()
	_, ok := msg.(searchResultsMsg)
	if !ok {
		t.Fatalf("expected searchResultsMsg")
	}
	if _, found := m.index.Get(note.ID); found {
		t.Fatalf("expected note removed from index")
	}
}

// TestResizeUpdatesViewport verifies that resize updates list and viewport dimensions.
// This ensures layout adapts to terminal size changes.
func TestResizeUpdatesViewport(t *testing.T) {
	m := newTestModel(t)
	m.width = 120
	m.height = 40
	m.resize()
	if m.view.Width == 0 || m.view.Height == 0 {
		t.Fatalf("expected viewport dimensions to be set")
	}
	if m.list.Width() == 0 || m.list.Height() == 0 {
		t.Fatalf("expected list dimensions to be set")
	}
}

// TestRenderHeader verifies that header includes mode and regex indicators.
// This ensures status is visible to users.
func TestRenderHeader(t *testing.T) {
	m := newTestModel(t)
	m.regex = true
	m.mode = modeSearch
	m.status = "ok"
	h := m.renderHeader()
	if !strings.Contains(h, "REGEX ON") || !strings.Contains(h, "SEARCH") {
		t.Fatalf("unexpected header: %s", h)
	}
}
