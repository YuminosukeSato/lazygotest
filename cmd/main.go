//go:build !tview

package main

import (
	"log"

	"github.com/s21066/lazygotest/internal/application/testrun"
	"github.com/s21066/lazygotest/internal/domain/catalog"
	"github.com/s21066/lazygotest/internal/interfaces/process"
	"github.com/s21066/lazygotest/internal/ui"
)

// entry point placeholder: in a fully featured app this would set up tview primitives.
// For now we wire adapters and print a brief message to confirm wiring works.
func main() {
	tree := ui.NewTreeAdapter()
	list := ui.NewListAdapter()
	history := ui.NewHistoryAdapter()
	logview := ui.NewLogAdapter()

	noopRunner := process.NewGoTestRunner(process.NewNoopCmdRunner())
	app := ui.NewApp(tree, list, history, logview, testrun.NewService(noopRunner))
	app.SetSnapshot(catalog.NewSnapshot(nil), nil)

	log.Println("lazy gotest UI wiring complete (headless stub)")
}
