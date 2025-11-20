//go:build tview

package main

import (
	"context"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/s21066/lazygotest/internal/application/catalogsync"
	"github.com/s21066/lazygotest/internal/application/testrun"
	"github.com/s21066/lazygotest/internal/interfaces/process"
	"github.com/s21066/lazygotest/internal/ui"
)

func TestTViewAppStartup(t *testing.T) {
	app := tview.NewApplication()

	tree := ui.NewTreeAdapter()
	list := ui.NewListAdapter()
	hist := ui.NewHistoryAdapter()
	logView := ui.NewLogAdapter()

	pages := tview.NewPages()

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
		t.Fatalf("catalog: %v", err)
	}
	snap, _ := catalogsync.BuildSnapshotFromModules(modules)

	testRunner := process.NewGoTestRunner(goRunner)
	appSvc := testrun.NewService(testRunner)
	uiApp := ui.NewApp(tree, list, hist, logView, appSvc, ui.WithFilterUI(filterUI), ui.WithHelpUI(helpUI))
	uiApp.SetSnapshot(snap, nil)

	cols := tview.NewFlex().
		AddItem(tree, 0, 1, true).
		AddItem(list, 0, 1, false).
		AddItem(hist, 0, 1, false)
	root := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(cols, 0, 3, true).
		AddItem(logView, 0, 1, false)

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
			}
		}
		if key != "" && uiApp.HandleKey(key) {
			return nil
		}
		return ev
	})

	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatalf("initialize simulation screen: %v", err)
	}

	app.SetScreen(screen)

	done := make(chan error, 1)
	go func() {
		done <- app.SetRoot(pages, true).EnableMouse(true).Run()
	}()

	time.Sleep(100 * time.Millisecond)

	app.Stop()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("app.Run failed: %v", err)
		}
	case <-time.After(5 * time.Second):
		screen.Fini()
		t.Fatal("app startup timed out")
	}
}
