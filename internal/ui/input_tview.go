//go:build tview

package ui

import (
	"github.com/rivo/tview"
)

// WireTviewInputs sets filter/help UI hooks on App using tview modals.
func WireTviewInputs(app *App, pages *tview.Pages, log *LogAdapter) {
	app.filterUI = func(apply func(string)) {
		input := tview.NewInputField().
			SetLabel("Filter: ").
			SetDoneFunc(func(key tview.Key) {
				if key == tview.KeyEnter {
					apply(input.GetText())
				}
				pages.RemovePage("filter")
			})
		modal := tview.NewModal().
			AddButtons([]string{"Close"}).
			SetText("Type to filter tests").
			SetFocus(0)
		flex := tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(input, 3, 0, true).
			AddItem(modal, 0, 1, false)
		pages.AddPage("filter", flex, true, true)
	}

	app.helpUI = func() {
		text := "[yellow]Keys[-]\n" +
			"h/j/k/l move, gg/G top/bottom, Enter run, r rerun, / filter, ? help, q quit"
		tv := tview.NewTextView().SetDynamicColors(true).SetText(text)
		modal := tview.NewModal().
			AddButtons([]string{"Close"}).
			SetText("Help")
		flex := tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(tv, 0, 1, true).
			AddItem(modal, 3, 0, false)
		pages.AddPage("help", flex, true, true)
	}

	// ensure log appends scroll
	logAppend := log.Append
	log.Append = func(line string) {
		logAppend(line)
		log.ScrollToEnd()
	}
}
