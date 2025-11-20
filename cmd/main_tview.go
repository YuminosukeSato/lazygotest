//go:build tview

package main

import (
	"context"
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/YuminosukeSato/lazygotest/internal/application/catalogsync"
	"github.com/YuminosukeSato/lazygotest/internal/application/testrun"
	"github.com/YuminosukeSato/lazygotest/internal/interfaces/process"
	"github.com/YuminosukeSato/lazygotest/internal/ui"
)

func main() {
	app := tview.NewApplication()

	theme := ui.DefaultTheme()
	tree := ui.NewTreeAdapter()
	list := ui.NewListAdapter()
	hist := ui.NewHistoryAdapter()
	logView := ui.NewLogAdapter()
	statusBar := ui.NewStatusBar(theme)

	pages := tview.NewPages()

	// Wire filter/help handlers using tview modals.
	filterUI := func(apply func(string)) {
		inField := tview.NewInputField()
		inField.SetLabel("Filter: ").SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				apply(inField.GetText())
			}
			pages.RemovePage("filter")
			app.SetFocus(list)
		})
		modal := tview.NewModal().
			SetText("Enter text to filter tests").
			AddButtons([]string{"Close"}).
			SetDoneFunc(func(_ int, _ string) {
				pages.RemovePage("filter")
				app.SetFocus(list)
			})
		flex := tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(inField, 3, 0, true).
			AddItem(modal, 0, 1, false)
		pages.AddPage("filter", flex, true, true)
		app.SetFocus(inField)
	}

	helpUI := func() {
		text := "[yellow]Keys[-]\n" +
			"h/j/k/l move, gg/G top/bottom, Enter run, r rerun, / filter, ? help, Tab/Shift-Tab focus, q quit"
		tv := tview.NewTextView().SetDynamicColors(true).SetText(text)
		modal := tview.NewModal().
			AddButtons([]string{"Close"}).
			SetDoneFunc(func(_ int, _ string) {
				pages.RemovePage("help")
				app.SetFocus(list)
			})
		flex := tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(tv, 0, 1, true).
			AddItem(modal, 3, 0, false)
		pages.AddPage("help", flex, true, true)
	}

	goRunner := process.NewOSRunner()
	src := process.NewGoCmdSource(".", goRunner)
	modules, err := src.Modules(context.Background())
	if err != nil {
		log.Fatalf("catalog: %v", err)
	}
	snap, _ := catalogsync.BuildSnapshotFromModules(modules)

	testRunner := process.NewGoTestRunner(goRunner)
	appSvc := testrun.NewService(testRunner)
	uiApp := ui.NewApp(tree, list, hist, logView, appSvc, ui.WithFilterUI(filterUI), ui.WithHelpUI(helpUI))
	uiApp.SetSnapshot(snap, nil)

	// layout
	cols := tview.NewFlex().
		AddItem(tree, 0, 1, true).
		AddItem(list, 0, 1, false).
		AddItem(hist, 0, 1, false)
	root := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(cols, 0, 3, true).
		AddItem(logView, 0, 1, false).
		AddItem(statusBar, 1, 0, false)

	tree.SetFocus(true)

	pages.AddPage("root", root, true, true)

	app.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		key := ""
		if ev.Key() == tcell.KeyRune {
			key = string(ev.Rune())
		} else {
			switch ev.Key() {
			case tcell.KeyEnter:
				key = "Enter"
			case tcell.KeyTab:
				key = "Tab"
			case tcell.KeyBacktab:
				key = "Shift-Tab"
			case tcell.KeyEsc:
				key = "q"
			case tcell.KeyUp:
				key = "k"
			case tcell.KeyDown:
				key = "j"
			case tcell.KeyLeft:
				key = "h"
			case tcell.KeyRight:
				key = "l"
			case tcell.KeyPgUp:
				key = "gg"
			case tcell.KeyPgDn:
				key = "G"
			}
		}
		if key != "" && uiApp.HandleKey(key) {
			return nil
		}
		return ev
	})

	// Prefer real screen; if unavailable (e.g., no /dev/tty), fall back to simulation to avoid hard failure.
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Printf("warning: real screen unavailable (%v), falling back to simulation screen", err)
		screen = tcell.NewSimulationScreen("")
	}
	if err := screen.Init(); err != nil {
		log.Printf("warning: screen init failed (%v), using simulation screen", err)
		screen = tcell.NewSimulationScreen("")
		if err := screen.Init(); err != nil {
			log.Fatalf("initialize screen: %v", err)
		}
	}
	app.SetScreen(screen)

	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		log.Fatal(err)
	}
}
