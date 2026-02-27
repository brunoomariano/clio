package ui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"clio/internal/index"
	"clio/internal/model"
	"clio/internal/store"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type inputMode int

const (
	modeSearch inputMode = iota
	modeTags
	modeExpiry
	modeBoost
	modeExclude
	modeMenu
)

type editorFinishedMsg struct {
	path string
	err  error
}

type searchResultsMsg struct {
	results []index.SearchResult
	err     error
}

type WatcherMsg struct {
	Path string
	Op   string
	Err  error
}

// Model is the Bubble Tea model for the application.
type Model struct {
	cfg         model.Config
	store       *store.NoteStore
	index       *index.Index
	deb         index.DebouncedExecutor[[]index.SearchResult]
	ctx         context.Context
	cancel      context.CancelFunc
	keyMap      keyMap
	help        help.Model
	list        list.Model
	menu        list.Model
	view        viewport.Model
	search      textinput.Model
	prompt      textinput.Model
	mode        inputMode
	regex       bool
	status      string
	boost       []string
	exclude     []string
	width       int
	height      int
	pendingQuit bool
	resultCount int
}

func New(cfg model.Config, store *store.NoteStore, idx *index.Index) *Model {
	search := textinput.New()
	search.Placeholder = "Search notes..."
	search.Focus()
	search.Prompt = "> "

	prompt := textinput.New()
	prompt.Placeholder = ""
	prompt.Prompt = ": "

	delegate := newNoteDelegate()
	lst := list.New([]list.Item{}, delegate, 0, 0)
	lst.SetShowHelp(false)
	lst.SetShowStatusBar(false)
	lst.SetFilteringEnabled(false)
	lst.Styles.Title = lipgloss.NewStyle()

	menu := list.New(menuItems(), list.NewDefaultDelegate(), 0, 0)
	menu.SetShowHelp(false)
	menu.SetShowStatusBar(false)
	menu.SetFilteringEnabled(false)
	menu.Title = "CLIO MENU"

	vp := viewport.New(0, 0)
	vp.Style = lipgloss.NewStyle().Padding(0, 1)

	return &Model{
		cfg:    cfg,
		store:  store,
		index:  idx,
		deb:    index.DebouncedExecutor[[]index.SearchResult]{Delay: time.Duration(cfg.DebounceMS) * time.Millisecond},
		keyMap: newKeyMap(),
		help:   help.New(),
		list:   lst,
		menu:   menu,
		view:   vp,
		search: search,
		prompt: prompt,
		mode:   modeSearch,
	}
}

func (m *Model) Init() tea.Cmd {
	m.ctx, m.cancel = context.WithCancel(context.Background())
	return m.runSearch()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.resize()
		return m, nil
	case searchResultsMsg:
		if msg.err != nil {
			m.status = msg.err.Error()
			return m, nil
		}
		m.status = ""
		m.applyResults(msg.results)
		return m, nil
	case editorFinishedMsg:
		if msg.err != nil {
			m.status = msg.err.Error()
			return m, nil
		}
		return m, m.refreshFromFile(msg.path)
	case WatcherMsg:
		if msg.Err != nil {
			m.status = msg.Err.Error()
			return m, nil
		}
		return m, m.refreshFromFile(msg.Path)
	case tea.KeyMsg:
		if handled, cmd := m.handleKey(msg); handled {
			return m, cmd
		}
	}

	var cmd tea.Cmd
	if m.mode == modeSearch {
		m.search, cmd = m.search.Update(msg)
		if msg, ok := msg.(tea.KeyMsg); ok {
			if msg.Type == tea.KeyRunes || msg.Type == tea.KeyBackspace || msg.Type == tea.KeyDelete {
				return m, tea.Batch(cmd, m.runSearch())
			}
		}
	} else if m.mode == modeMenu {
		m.menu, cmd = m.menu.Update(msg)
	} else {
		m.prompt, cmd = m.prompt.Update(msg)
	}

	var listCmd tea.Cmd
	m.list, listCmd = m.list.Update(msg)
	m.updateViewport()
	return m, tea.Batch(cmd, listCmd)
}

func (m *Model) View() string {
	header := m.renderHeader()
	searchBar := m.renderSearchBar()
	stats := m.renderStats()
	pattern := m.renderPatternLine()
	left := m.renderLeft()
	right := m.renderRight()
	footer := m.renderFooter()

	bodyHeight := m.height - lipgloss.Height(header) - lipgloss.Height(searchBar) - lipgloss.Height(stats) - lipgloss.Height(pattern) - lipgloss.Height(footer) - 1
	if bodyHeight < 3 {
		bodyHeight = 3
	}
	layout := lipgloss.NewStyle().Width(m.width).Height(bodyHeight)
	columns := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	body := layout.Render(columns)
	if m.mode == modeMenu {
		body = m.renderMenuOverlay(bodyHeight)
	}

	wrapped := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1).
		UnsetForeground().
		UnsetBackground().
		Render(lipgloss.JoinVertical(lipgloss.Left, header, searchBar, stats, pattern, body, footer))

	return wrapped
}

func (m *Model) handleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	if msg.Type != tea.KeyCtrlC && m.pendingQuit {
		m.pendingQuit = false
		if m.status == "Press CTRL+C again to quit" {
			m.status = ""
		}
	}

	if key.Matches(msg, m.keyMap.Menu) {
		if m.mode == modeMenu {
			m.mode = modeSearch
			m.search.Focus()
			return true, nil
		}
		m.mode = modeMenu
		return true, nil
	}

	if msg.Type == tea.KeyCtrlC {
		if m.pendingQuit {
			if m.cancel != nil {
				m.cancel()
			}
			return true, tea.Quit
		}
		m.pendingQuit = true
		m.status = "Press CTRL+C again to quit"
		return true, nil
	}

	if m.mode == modeMenu && msg.Type == tea.KeyEnter {
		return true, m.applyMenu()
	}
	if m.mode != modeSearch && msg.Type == tea.KeyEnter {
		return true, m.applyPrompt()
	}

	return false, nil
}

func (m *Model) activatePrompt(mode inputMode, placeholder string) {
	m.mode = mode
	m.prompt.SetValue("")
	m.prompt.Placeholder = placeholder
	m.prompt.Focus()
	m.search.Blur()
}

func (m *Model) applyPrompt() tea.Cmd {
	value := strings.TrimSpace(m.prompt.Value())
	m.prompt.Blur()
	m.prompt.Reset()
	m.search.Focus()

	switch m.mode {
	case modeBoost:
		if value != "" {
			m.boost = append(m.boost, value)
		}
	case modeExclude:
		if value != "" {
			m.exclude = append(m.exclude, value)
		}
	case modeTags:
		note := m.selectedNote()
		if note != nil {
			if value == "" {
				note.Tags = nil
			} else {
				note.Tags = splitTags(value)
			}
			_ = m.store.UpdateNote(note)
			m.index.Upsert(note)
		}
	case modeExpiry:
		note := m.selectedNote()
		if note != nil {
			if value == "" {
				note.ExpiresAt = nil
			} else if t, err := time.Parse(time.RFC3339, value); err == nil {
				note.ExpiresAt = &t
			}
			_ = m.store.UpdateNote(note)
			m.index.Upsert(note)
		}
	}
	m.mode = modeSearch
	return m.runSearch()
}

type menuItem string

func (m menuItem) Title() string       { return string(m) }
func (m menuItem) Description() string { return "" }
func (m menuItem) FilterValue() string { return string(m) }

func menuItems() []list.Item {
	return []list.Item{
		menuItem("New note"),
		menuItem("Open/Edit"),
		menuItem("Delete"),
		menuItem("Edit tags"),
		menuItem("Set/Clear expiry"),
		menuItem("Toggle regex"),
		menuItem("Add boost tag"),
		menuItem("Add exclude tag"),
		menuItem("Quit"),
	}
}

func (m *Model) applyMenu() tea.Cmd {
	item := m.menu.SelectedItem()
	if item == nil {
		return nil
	}
	switch item.(menuItem) {
	case "New note":
		m.mode = modeSearch
		return m.createNoteCmd()
	case "Open/Edit":
		m.mode = modeSearch
		return m.editSelectedCmd()
	case "Delete":
		m.mode = modeSearch
		return m.deleteSelectedCmd()
	case "Edit tags":
		m.activatePrompt(modeTags, "Tags (comma separated)")
		return nil
	case "Set/Clear expiry":
		m.activatePrompt(modeExpiry, "Expiry RFC3339 (empty clears)")
		return nil
	case "Toggle regex":
		m.regex = !m.regex
		m.mode = modeSearch
		return m.runSearch()
	case "Add boost tag":
		m.activatePrompt(modeBoost, "Boost tag")
		return nil
	case "Add exclude tag":
		m.activatePrompt(modeExclude, "Exclude tag")
		return nil
	case "Quit":
		m.pendingQuit = true
		m.status = "Press CTRL+C again to quit"
		m.mode = modeSearch
		return nil
	default:
		return nil
	}
}

func splitTags(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func (m *Model) runSearch() tea.Cmd {
	if m.cancel != nil {
		m.cancel()
	}
	m.ctx, m.cancel = context.WithCancel(context.Background())
	query := m.search.Value()
	opts := index.SearchOptions{
		Query:       query,
		MaxResults:  m.cfg.MaxResults,
		BoostTags:   m.boost,
		ExcludeTags: m.exclude,
		BoostWeight: m.cfg.BoostWeight,
		K1:          m.cfg.BM25K1,
		B:           m.cfg.BM25B,
		Regex:       m.regex,
		Now:         time.Now().UTC(),
	}
	return func() tea.Msg {
		ch := m.deb.Run(m.ctx, func(ctx context.Context) ([]index.SearchResult, error) {
			if strings.TrimSpace(query) == "" {
				return m.allNotes(opts), nil
			}
			return m.index.Search(opts)
		})
		res, ok := <-ch
		if !ok {
			return nil
		}
		return searchResultsMsg{results: res.Value, err: res.Err}
	}
}

func (m *Model) allNotes(opts index.SearchOptions) []index.SearchResult {
	notes := m.index.AllNotes()
	results := make([]index.SearchResult, 0, len(notes))
	exclude := make(map[string]struct{})
	for _, tag := range opts.ExcludeTags {
		tag = strings.TrimSpace(strings.ToLower(tag))
		if tag != "" {
			exclude[tag] = struct{}{}
		}
	}
	for _, note := range notes {
		if note.ExpiresAt != nil && note.ExpiresAt.Before(opts.Now) {
			continue
		}
		if len(exclude) > 0 {
			skip := false
			for _, tag := range note.Tags {
				if _, ok := exclude[strings.ToLower(tag)]; ok {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
		}
		results = append(results, index.SearchResult{Note: note, Score: 0})
	}
	index.SortResults(results)
	if len(results) > opts.MaxResults {
		return results[:opts.MaxResults]
	}
	return results
}

func (m *Model) applyResults(results []index.SearchResult) {
	items := make([]list.Item, 0, len(results))
	for _, res := range results {
		items = append(items, noteItem{note: res.Note})
	}
	m.resultCount = len(results)
	m.list.SetItems(items)
	if len(items) > 0 {
		m.list.Select(0)
	}
	m.updateViewport()
}

func (m *Model) selectedNote() *model.Note {
	item := m.list.SelectedItem()
	if item == nil {
		return nil
	}
	return item.(noteItem).note
}

func (m *Model) updateViewport() {
	note := m.selectedNote()
	if note == nil {
		m.view.SetContent("")
		return
	}
	m.view.SetContent(note.Body)
}

func (m *Model) renderHeader() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("213")).
		Render("CLIO")
	count := lipgloss.NewStyle().
		Foreground(lipgloss.Color("109")).
		Render(fmt.Sprintf("%d notes", m.index.NotesCount()))
	mode := "SEARCH"
	if m.mode == modeMenu {
		mode = "MENU"
	} else if m.mode != modeSearch {
		mode = "PROMPT"
	}
	regex := "REGEX OFF"
	if m.regex {
		regex = "REGEX ON"
	}
	status := ""
	if m.status != "" {
		status = " | " + m.status
	}
	return lipgloss.JoinHorizontal(lipgloss.Left,
		title,
		"  ",
		count,
		"  ",
		lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Render(mode),
		"  ",
		lipgloss.NewStyle().Foreground(lipgloss.Color("204")).Render(regex),
		status,
	)
}

func (m *Model) renderLeft() string {
	chips := renderChips(m.boost, m.exclude)
	left := lipgloss.NewStyle().Width(m.width/2-1).Padding(0, 1)
	content := lipgloss.JoinVertical(lipgloss.Left, chips, m.list.View())
	return left.Render(content)
}

func (m *Model) renderRight() string {
	right := lipgloss.NewStyle().Width(m.width/2-1).Padding(0, 1)
	return right.Render(m.view.View())
}

func (m *Model) renderFooter() string {
	return m.help.View(m.keyMap)
}

func (m *Model) resize() {
	leftWidth := m.width/2 - 2
	if leftWidth < 20 {
		leftWidth = 20
	}
	rightWidth := m.width - leftWidth - 4
	m.list.SetWidth(leftWidth)
	m.list.SetHeight(m.height - 6)
	m.menu.SetWidth(56)
	m.menu.SetHeight(16)
	m.view.Width = rightWidth
	m.view.Height = m.height - 6
}

func (m *Model) renderSearchBar() string {
	display := m.search.Value()
	if m.mode != modeSearch && m.mode != modeMenu {
		display = m.prompt.Value()
	}
	if strings.TrimSpace(display) == "" {
		display = "Search notes..."
	}

	switchText := "REGEX OFF"
	if m.regex {
		switchText = "REGEX ON"
	}

	bar := fmt.Sprintf(" SEARCH: %s  [ %s ] ", display, switchText)

	barStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("235")).
		Padding(0, 2).
		Bold(true)

	if m.regex {
		barStyle = barStyle.Background(lipgloss.Color("24"))
	}

	return lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(barStyle.Render(bar))
}

func (m *Model) renderStats() string {
	stats := fmt.Sprintf("%d results  •  %d total", m.resultCount, m.index.NotesCount())
	statsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("110"))
	return lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(statsStyle.Render(stats))
}

func (m *Model) renderPatternLine() string {
	if m.width <= 0 {
		return ""
	}
	pattern := strings.Repeat(".:", (m.width/2)+2)
	pattern = pattern[:max(1, m.width)]
	return lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Foreground(lipgloss.Color("236")).Render(pattern)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m *Model) renderMenuOverlay(bodyHeight int) string {
	menuBox := lipgloss.NewStyle().
		Width(60).
		Padding(1, 3).
		Border(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("63")).
		Render(m.menu.View())
	return lipgloss.Place(m.width, bodyHeight, lipgloss.Center, lipgloss.Center, menuBox)
}

func resolveTerminal(preferred string) string {
	candidates := []string{}
	if preferred != "" {
		candidates = append(candidates, preferred)
	}
	if env := os.Getenv("TERMINAL"); env != "" {
		candidates = append(candidates, env)
	}
	candidates = append(candidates,
		"x-terminal-emulator",
		"gnome-terminal",
		"kitty",
		"alacritty",
		"wezterm",
		"foot",
		"konsole",
		"xfce4-terminal",
		"lxterminal",
		"tilix",
	)
	for _, term := range candidates {
		if _, err := exec.LookPath(term); err == nil {
			return term
		}
	}
	return ""
}

func terminalArgs(term, editor, path string) []string {
	switch term {
	case "kitty", "wezterm", "alacritty", "foot", "konsole", "xfce4-terminal", "lxterminal", "tilix":
		return []string{"-e", editor, path}
	case "gnome-terminal":
		return []string{"--", editor, path}
	case "x-terminal-emulator":
		return []string{"-e", editor, path}
	default:
		return []string{"-e", editor, path}
	}
}

func (m *Model) createNoteCmd() tea.Cmd {
	return func() tea.Msg {
		note, err := m.store.CreateNote("", "", nil, nil, time.Now().UTC())
		if err != nil {
			return editorFinishedMsg{err: err}
		}
		m.index.Upsert(note)
		path := m.store.NotePath(note.ID)
		return m.openEditor(path)
	}
}

func (m *Model) editSelectedCmd() tea.Cmd {
	note := m.selectedNote()
	if note == nil {
		return nil
	}
	path := m.store.NotePath(note.ID)
	return func() tea.Msg { return m.openEditor(path) }
}

func (m *Model) deleteSelectedCmd() tea.Cmd {
	note := m.selectedNote()
	if note == nil {
		return nil
	}
	return func() tea.Msg {
		if err := m.store.DeleteNote(note.ID); err != nil {
			return editorFinishedMsg{err: err}
		}
		m.index.Remove(note.ID)
		return searchResultsMsg{results: m.allNotes(index.SearchOptions{MaxResults: m.cfg.MaxResults, Now: time.Now().UTC()})}
	}
}

func (m *Model) openEditor(path string) tea.Msg {
	editor := m.cfg.Editor
	if editor == "" {
		editor = "nvim"
	}
	if _, err := exec.LookPath(editor); err != nil {
		editor = "nano"
	}

	terminal := resolveTerminal(m.cfg.Terminal)
	if terminal == "" {
		// Fallback to direct editor in current terminal.
		cmd := exec.Command(editor, path)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return editorFinishedMsg{path: path, err: cmd.Run()}
	}

	args := terminalArgs(terminal, editor, path)
	cmd := exec.Command(terminal, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return editorFinishedMsg{path: path, err: cmd.Run()}
}

func (m *Model) refreshFromFile(path string) tea.Cmd {
	return func() tea.Msg {
		if path == "" {
			return searchResultsMsg{results: m.allNotes(index.SearchOptions{MaxResults: m.cfg.MaxResults, Now: time.Now().UTC()})}
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}
		data, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			if os.IsNotExist(err) {
				id := strings.TrimSuffix(filepath.Base(path), ".md")
				m.index.Remove(id)
				return searchResultsMsg{results: m.allNotes(index.SearchOptions{MaxResults: m.cfg.MaxResults, Now: time.Now().UTC()})}
			}
			return editorFinishedMsg{err: err}
		}
		note, err := model.ParseNoteBytes(data)
		if err != nil {
			return editorFinishedMsg{err: err}
		}
		m.index.Upsert(note)
		return searchResultsMsg{results: m.allNotes(index.SearchOptions{MaxResults: m.cfg.MaxResults, Now: time.Now().UTC()})}
	}
}
