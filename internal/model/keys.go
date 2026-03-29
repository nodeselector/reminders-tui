package model

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	Complete key.Binding
	Add      key.Binding
	Edit     key.Binding
	Delete   key.Binding
	Search   key.Binding
	Refresh  key.Binding
	Quit     key.Binding
	Help     key.Binding
	Tab      key.Binding
	View1    key.Binding
	View2    key.Binding
	View3    key.Binding
	View4    key.Binding
	Enter    key.Binding
	Escape   key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("k/↑", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("j/↓", "down"),
	),
	Complete: key.NewBinding(
		key.WithKeys("x", " "),
		key.WithHelp("x/space", "complete"),
	),
	Add: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next view"),
	),
	View1: key.NewBinding(
		key.WithKeys("1"),
		key.WithHelp("1", "today"),
	),
	View2: key.NewBinding(
		key.WithKeys("2"),
		key.WithHelp("2", "upcoming"),
	),
	View3: key.NewBinding(
		key.WithKeys("3"),
		key.WithHelp("3", "lists"),
	),
	View4: key.NewBinding(
		key.WithKeys("4"),
		key.WithHelp("4", "all"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
	),
}
