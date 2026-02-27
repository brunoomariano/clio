package ui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	FocusSearch key.Binding
	Open        key.Binding
	New         key.Binding
	Edit        key.Binding
	Delete      key.Binding
	Tags        key.Binding
	Expiry      key.Binding
	Regex       key.Binding
	Boost       key.Binding
	Exclude     key.Binding
	Quit        key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		FocusSearch: key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
		Open:        key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open")),
		New:         key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new")),
		Edit:        key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
		Delete:      key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
		Tags:        key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "tags")),
		Expiry:      key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "expiry")),
		Regex:       key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "regex")),
		Boost:       key.NewBinding(key.WithKeys("+"), key.WithHelp("+", "boost")),
		Exclude:     key.NewBinding(key.WithKeys("-"), key.WithHelp("-", "exclude")),
		Quit:        key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.FocusSearch, k.New, k.Edit, k.Delete, k.Tags, k.Expiry, k.Regex, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.FocusSearch, k.Open, k.New, k.Edit, k.Delete},
		{k.Tags, k.Expiry, k.Regex, k.Boost, k.Exclude, k.Quit},
	}
}
