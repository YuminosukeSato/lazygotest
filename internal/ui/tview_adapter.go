package ui

import (
	"github.com/rivo/tview"
	"github.com/s21066/lazygotest/internal/presentation"
)

// TreeAdapter wraps tview.TreeView and keeps index-based selection.
type TreeAdapter struct {
	*tview.TreeView
	nodes   []*tview.TreeNode
	current int
}

func NewTreeAdapter() *TreeAdapter {
	tv := tview.NewTreeView()
	return &TreeAdapter{TreeView: tv, current: 0}
}

func (t *TreeAdapter) SetItems(items []presentation.PackageNode) {
	root := tview.NewTreeNode("packages")
	t.nodes = make([]*tview.TreeNode, len(items))
	for i, item := range items {
		node := tview.NewTreeNode(item.Label).SetReference(i)
		t.nodes[i] = node
		root.AddChild(node)
	}
	t.current = 0
	t.SetRoot(root)
	if len(t.nodes) > 0 {
		t.SetCurrentNode(t.nodes[0])
	}
}

func (t *TreeAdapter) Move(delta int) {
	if len(t.nodes) == 0 {
		return
	}
	t.current += delta
	if t.current < 0 {
		t.current = 0
	}
	if t.current >= len(t.nodes) {
		t.current = len(t.nodes) - 1
	}
	t.SetCurrentNode(t.nodes[t.current])
}

func (t *TreeAdapter) Current() int { return t.current }

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
	count := l.GetItemCount()
	if count == 0 {
		return
	}
	idx := l.GetCurrentItem() + delta
	if idx < 0 {
		idx = 0
	}
	if idx >= count {
		idx = count - 1
	}
	l.SetCurrentItem(idx)
}

func (l *ListAdapter) Current() int { return l.GetCurrentItem() }

type HistoryAdapter struct {
	*tview.Table
}

func NewHistoryAdapter() *HistoryAdapter {
	tbl := tview.NewTable().SetFixed(1, 0)
	return &HistoryAdapter{Table: tbl}
}

func (h *HistoryAdapter) SetItems(items []presentation.RunRow) {
	h.Clear()
	for row, it := range items {
		h.SetCell(row, 0, tview.NewTableCell(it.Label))
		h.SetCell(row, 1, tview.NewTableCell(string(it.Status)))
		h.SetCell(row, 2, tview.NewTableCell(it.Duration))
	}
	if h.GetRowCount() > 0 {
		h.Select(h.GetRowCount()-1, 0)
	}
}

func (h *HistoryAdapter) ScrollToEnd() {
	if rows := h.GetRowCount(); rows > 0 {
		h.Select(rows-1, 0)
	}
}

type LogAdapter struct {
	*tview.TextView
}

func NewLogAdapter() *LogAdapter {
	tv := tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	return &LogAdapter{TextView: tv}
}

func (l *LogAdapter) Append(line string) {
	l.Write([]byte(line))
	l.ScrollToEnd()
}

func (l *LogAdapter) Clear()       { l.TextView.Clear() }
func (l *LogAdapter) ScrollToEnd() { l.TextView.ScrollToEnd() }
