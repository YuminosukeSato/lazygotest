//go:build tview

package ui

import (
	"fmt"

	"github.com/rivo/tview"
)

type StatusBar struct {
	*tview.TextView
	theme Theme
}

func NewStatusBar(theme Theme) *StatusBar {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	tv.SetBackgroundColor(theme.StatusBarBG)
	tv.SetTextColor(theme.StatusBarText)

	sb := &StatusBar{
		TextView: tv,
		theme:    theme,
	}
	sb.UpdateHelp("tree")
	return sb
}

func (sb *StatusBar) UpdateHelp(focus string) {
	keys := sb.getKeysForFocus(focus)
	text := sb.formatKeys(keys)
	sb.SetText(text)
}

func (sb *StatusBar) getKeysForFocus(focus string) []KeyHelp {
	common := []KeyHelp{
		{"q", "quit"},
		{"Tab", "next"},
		{"?", "help"},
	}

	switch focus {
	case "tree", "list":
		return append([]KeyHelp{
			{"↑↓/jk", "move"},
			{"gg/G", "top/btm"},
			{"Enter", "run"},
			{"r", "rerun"},
			{"/", "filter"},
		}, common...)
	case "history":
		return append([]KeyHelp{
			{"↑↓/jk", "move"},
			{"Enter", "view"},
		}, common...)
	case "log":
		return append([]KeyHelp{
			{"↑↓/jk", "scroll"},
			{"PgUp/PgDn", "page"},
		}, common...)
	default:
		return common
	}
}

type KeyHelp struct {
	Key  string
	Desc string
}

func (sb *StatusBar) formatKeys(keys []KeyHelp) string {
	var text string
	for i, kh := range keys {
		if i > 0 {
			text += "  "
		}
		text += fmt.Sprintf("[white:darkcyan:b]%s[-:-:-] %s",
			kh.Key,
			kh.Desc)
	}
	return text
}
