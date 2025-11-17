package ui

import "github.com/s21066/lazygotest/internal/presentation"

// NOTE: This is a thin, tview-like adapter implemented locally to keep tests and builds
// working in offline environments. It satisfies the pane interfaces used by App.

type TreeAdapter struct {
	items []presentation.PackageNode
	cur   int
}

func NewTreeAdapter() *TreeAdapter { return &TreeAdapter{cur: 0} }

func (t *TreeAdapter) SetItems(items []presentation.PackageNode) {
	t.items = items
	if t.cur >= len(t.items) {
		t.cur = len(t.items) - 1
	}
	if t.cur < 0 {
		t.cur = 0
	}
}

func (t *TreeAdapter) Move(delta int) {
	t.cur += delta
	if t.cur < 0 {
		t.cur = 0
	}
	if t.cur >= len(t.items) {
		t.cur = len(t.items) - 1
	}
}

func (t *TreeAdapter) Current() int { return t.cur }

type ListAdapter struct {
	items []presentation.TestRow
	cur   int
}

func NewListAdapter() *ListAdapter { return &ListAdapter{cur: 0} }

func (l *ListAdapter) SetItems(items []presentation.TestRow) {
	l.items = items
	if l.cur >= len(l.items) {
		l.cur = len(l.items) - 1
	}
	if l.cur < 0 {
		l.cur = 0
	}
}

func (l *ListAdapter) Move(delta int) {
	l.cur += delta
	if l.cur < 0 {
		l.cur = 0
	}
	if l.cur >= len(l.items) {
		l.cur = len(l.items) - 1
	}
}

func (l *ListAdapter) Current() int { return l.cur }

type HistoryAdapter struct {
	items    []presentation.RunRow
	scrolled bool
}

func NewHistoryAdapter() *HistoryAdapter { return &HistoryAdapter{} }

func (h *HistoryAdapter) SetItems(items []presentation.RunRow) { h.items = items }
func (h *HistoryAdapter) ScrollToEnd()                         { h.scrolled = true }

type LogAdapter struct {
	lines []string
}

func NewLogAdapter() *LogAdapter { return &LogAdapter{} }

func (l *LogAdapter) Append(line string) { l.lines = append(l.lines, line) }
func (l *LogAdapter) Clear()             { l.lines = nil }
func (l *LogAdapter) ScrollToEnd()       {}
