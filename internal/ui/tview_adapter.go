//go:build tview

package ui

import (
	"fmt"

	"github.com/rivo/tview"
	"github.com/s21066/lazygotest/internal/domain/execution"
	"github.com/s21066/lazygotest/internal/presentation"
)

// TreeAdapter wraps tview.TreeView and keeps index-based selection.
type TreeAdapter struct {
	*tview.TreeView
	nodes   []*tview.TreeNode
	current int
	theme   Theme
}

func NewTreeAdapter() *TreeAdapter {
	theme := DefaultTheme()
	tv := tview.NewTreeView()
	tv.SetBorder(true).
		SetTitle(" Packages ").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(theme.BorderNormal)
	return &TreeAdapter{TreeView: tv, current: 0, theme: theme}
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

func (t *TreeAdapter) SetFocus(focused bool) {
	if focused {
		t.SetBorderColor(t.theme.BorderFocus)
	} else {
		t.SetBorderColor(t.theme.BorderNormal)
	}
}

type ListAdapter struct {
	*tview.List
	theme Theme
}

func NewListAdapter() *ListAdapter {
	theme := DefaultTheme()
	list := tview.NewList()
	list.SetBorder(true).
		SetTitle(" Tests ").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(theme.BorderNormal)
	return &ListAdapter{List: list, theme: theme}
}

func (l *ListAdapter) SetItems(items []presentation.TestRow) {
	l.Clear()
	for _, it := range items {
		icon := string(IconPending)
		l.AddItem(fmt.Sprintf("%s %s", icon, it.Name), string(it.Kind), 0, nil)
	}
}

func (l *ListAdapter) SetFocus(focused bool) {
	if focused {
		l.SetBorderColor(l.theme.BorderFocus)
	} else {
		l.SetBorderColor(l.theme.BorderNormal)
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
	theme Theme
}

func NewHistoryAdapter() *HistoryAdapter {
	theme := DefaultTheme()
	tbl := tview.NewTable().SetFixed(1, 0)
	tbl.SetBorder(true).
		SetTitle(" History ").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(theme.BorderNormal)
	return &HistoryAdapter{Table: tbl, theme: theme}
}

func (h *HistoryAdapter) SetItems(items []presentation.RunRow) {
	h.Clear()
	for row, it := range items {
		icon := h.getStatusIcon(it.Status)
		label := fmt.Sprintf("%s %s", icon, it.Label)
		h.SetCell(row, 0, tview.NewTableCell(label))
		h.SetCell(row, 1, tview.NewTableCell(string(it.Status)))
		h.SetCell(row, 2, tview.NewTableCell(it.Duration))
	}
	if h.GetRowCount() > 0 {
		h.Select(h.GetRowCount()-1, 0)
	}
}

func (h *HistoryAdapter) getStatusIcon(status execution.Status) string {
	switch status {
	case execution.StatusSuccess:
		return string(IconPass)
	case execution.StatusFailed:
		return string(IconFail)
	case execution.StatusRunning:
		return string(IconRunning)
	default:
		return string(IconPending)
	}
}

func (h *HistoryAdapter) SetFocus(focused bool) {
	if focused {
		h.SetBorderColor(h.theme.BorderFocus)
	} else {
		h.SetBorderColor(h.theme.BorderNormal)
	}
}

func (h *HistoryAdapter) ScrollToEnd() {
	if rows := h.GetRowCount(); rows > 0 {
		h.Select(rows-1, 0)
	}
}

type LogAdapter struct {
	*tview.TextView
	theme Theme
}

func NewLogAdapter() *LogAdapter {
	theme := DefaultTheme()
	tv := tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	tv.SetBorder(true).
		SetTitle(" Log ").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(theme.BorderNormal)
	return &LogAdapter{TextView: tv, theme: theme}
}

func (l *LogAdapter) Append(line string) {
	_, _ = l.Write([]byte(line))
	l.ScrollToEnd()
}

func (l *LogAdapter) SetFocus(focused bool) {
	if focused {
		l.SetBorderColor(l.theme.BorderFocus)
	} else {
		l.SetBorderColor(l.theme.BorderNormal)
	}
}

func (l *LogAdapter) Clear()       { l.TextView.Clear() }
func (l *LogAdapter) ScrollToEnd() { l.TextView.ScrollToEnd() }
