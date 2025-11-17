//go:build tview

package ui

import (
	"github.com/rivo/tview"
	"github.com/s21066/lazygotest/internal/presentation"
)

// This file provides real tview-backed adapters when built with -tags tview.

type TreeAdapter struct{ *tview.TreeView }

func NewTreeAdapter() *TreeAdapter {
	tv := tview.NewTreeView()
	return &TreeAdapter{TreeView: tv}
}

func (t *TreeAdapter) SetItems(items []presentation.PackageNode) {
	root := tview.NewTreeNode("packages")
	for i, item := range items {
		node := tview.NewTreeNode(item.Label).SetReference(i)
		root.AddChild(node)
	}
	t.SetRoot(root).SetCurrentNode(root)
}

func (t *TreeAdapter) Move(delta int) {
	cur := t.GetCurrentNode()
	if cur == nil {
		return
	}
	parent := cur.GetParent()
	if parent == nil {
		parent = t.GetRoot()
	}
	siblings := parent.GetChildren()
	idx := -1
	for i, n := range siblings {
		if n == cur {
			idx = i
			break
		}
	}
	if idx == -1 {
		return
	}
	idx += delta
	if idx < 0 {
		idx = 0
	}
	if idx >= len(siblings) {
		idx = len(siblings) - 1
	}
	if idx >= 0 && idx < len(siblings) {
		t.SetCurrentNode(siblings[idx])
	}
}

func (t *TreeAdapter) Current() int {
	cur := t.GetCurrentNode()
	if cur == nil {
		return -1
	}
	if ref := cur.GetReference(); ref != nil {
		if idx, ok := ref.(int); ok {
			return idx
		}
	}
	return -1
}

type ListAdapter struct {
	*tview.List
}

func NewListAdapter() *ListAdapter {
	return &ListAdapter{List: tview.NewList()}
}

func (l *ListAdapter) SetItems(items []presentation.TestRow) {
	l.Clear()
	for _, it := range items {
		l.AddItem(it.Name, string(it.Kind), 0, nil)
	}
}

func (l *ListAdapter) Move(delta int) {
	idx := l.GetCurrentItem() + delta
	if idx < 0 {
		idx = 0
	}
	if idx >= l.GetItemCount() {
		idx = l.GetItemCount() - 1
	}
	if idx >= 0 {
		l.SetCurrentItem(idx)
	}
}

func (l *ListAdapter) Current() int { return l.GetCurrentItem() }

type HistoryAdapter struct{ *tview.Table }

func NewHistoryAdapter() *HistoryAdapter {
	return &HistoryAdapter{Table: tview.NewTable().SetFixed(1, 0)}
}

func (h *HistoryAdapter) SetItems(items []presentation.RunRow) {
	h.Clear()
	for row, it := range items {
		h.SetCell(row, 0, tview.NewTableCell(it.Label))
		h.SetCell(row, 1, tview.NewTableCell(string(it.Status)))
		h.SetCell(row, 2, tview.NewTableCell(it.Duration))
	}
}

func (h *HistoryAdapter) ScrollToEnd() {
	// tview Table lacks direct ScrollToEnd; emulate by scrolling to last row.
	if rows := h.GetRowCount(); rows > 0 {
		h.Select(rows-1, 0)
		h.ScrollTo(rows-1, 0)
	}
}

type LogAdapter struct{ *tview.TextView }

func NewLogAdapter() *LogAdapter {
	tv := tview.NewTextView()
	tv.SetDynamicColors(true).SetScrollable(true)
	return &LogAdapter{TextView: tv}
}

func (l *LogAdapter) Append(line string) {
	l.Write([]byte(line))
}

func (l *LogAdapter) Clear() { l.TextView.Clear() }
