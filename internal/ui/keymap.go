package ui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Menu key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		Menu: key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "menu")),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Menu}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Menu},
	}
}
