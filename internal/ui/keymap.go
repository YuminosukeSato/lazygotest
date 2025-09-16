package ui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Help      key.Binding
	Next      key.Binding
	Prev      key.Binding
	Quit      key.Binding
	Run       key.Binding
	RunAll    key.Binding
	RunFailed key.Binding
	Repeat    key.Binding
	Filter    key.Binding
	Select    key.Binding
	FailOnly  key.Binding
	Race      key.Binding
	Cover     key.Binding
	Bench     key.Binding
	Fuzz      key.Binding
	Tags      key.Binding
	Watch     key.Binding
	Open      key.Binding
	Save      key.Binding
	Search    key.Binding
	FocusLogs key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Help:      key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Next:      key.NewBinding(key.WithKeys("tab"), key.WithHelp("Tab", "next panel")),
		Prev:      key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("S-Tab", "prev panel")),
		Quit:      key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		Run:       key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "run selected")),
		RunAll:    key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "run all")),
		RunFailed: key.NewBinding(key.WithKeys("R"), key.WithHelp("R", "run failed")),
		Repeat:    key.NewBinding(key.WithKeys("."), key.WithHelp(".", "repeat last")),
		Filter:    key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
		Select:    key.NewBinding(key.WithKeys(" "), key.WithHelp("Space", "toggle select")),
		FailOnly:  key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "show failed only")),
		Race:      key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "toggle -race")),
		Cover:     key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "toggle cover")),
		Bench:     key.NewBinding(key.WithKeys("b"), key.WithHelp("b", "toggle bench")),
		Fuzz:      key.NewBinding(key.WithKeys("z"), key.WithHelp("z", "toggle fuzz")),
		Tags:      key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "set tags")),
		Watch:     key.NewBinding(key.WithKeys("w"), key.WithHelp("w", "toggle watch")),
		Open:      key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "open editor")),
		Save:      key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "save logs")),
		Search:    key.NewBinding(key.WithKeys("ctrl+f"), key.WithHelp("C-f", "search logs")),
		FocusLogs: key.NewBinding(key.WithKeys("enter"), key.WithHelp("Enter", "focus logs")),
	}
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit, k.Run, k.RunAll}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Help, k.Quit, k.Next, k.Prev},
		{k.Run, k.RunAll, k.RunFailed, k.Repeat},
		{k.Filter, k.Select, k.FailOnly},
		{k.Race, k.Cover, k.Bench, k.Fuzz},
		{k.Tags, k.Watch, k.Open, k.Save},
		{k.Search, k.FocusLogs},
	}
}

func (k KeyMap) Tests() []key.Binding {
	return []key.Binding{k.Run, k.RunAll, k.RunFailed, k.Repeat, k.Select, k.Filter, k.FailOnly, k.Open}
}

func (k KeyMap) Flags() []key.Binding {
	return []key.Binding{k.Race, k.Cover, k.Bench, k.Fuzz, k.Tags, k.Watch}
}

func (k KeyMap) Logs() []key.Binding {
	return []key.Binding{k.Save, k.Search, k.FocusLogs}
}
