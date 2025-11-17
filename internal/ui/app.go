package ui

import (
	"context"
	"errors"
	"strings"

	"github.com/s21066/lazygotest/internal/application/testrun"
	"github.com/s21066/lazygotest/internal/domain/catalog"
	"github.com/s21066/lazygotest/internal/domain/execution"
	"github.com/s21066/lazygotest/internal/presentation"
)

// TreeView abstracts the package tree pane.
type TreeView interface {
	SetItems([]presentation.PackageNode)
	Move(delta int)
	Current() int
}

// ListView abstracts the test list pane.
type ListView interface {
	SetItems([]presentation.TestRow)
	Move(delta int)
	Current() int
}

// HistoryView abstracts the run history pane.
type HistoryView interface {
	SetItems([]presentation.RunRow)
	ScrollToEnd()
}

// LogView abstracts the log output pane.
type LogView interface {
	Append(line string)
	Clear()
}

type App struct {
	tree    TreeView
	tests   ListView
	history HistoryView
	log     LogView

	snapshot catalog.Snapshot
	vm       presentation.ViewModel
	runs     []execution.Run
	runner   testrun.Service
	filter   string
}

func NewApp(tree TreeView, tests ListView, history HistoryView, log LogView, runner testrun.Service) *App {
	return &App{
		tree:    tree,
		tests:   tests,
		history: history,
		log:     log,
		runner:  runner,
	}
}

func (a *App) SetSnapshot(snapshot catalog.Snapshot, runs []execution.Run) {
	a.snapshot = snapshot
	a.runs = runs
	a.vm = presentation.Project(snapshot, runs)
	a.tree.SetItems(a.vm.Packages)
	a.tests.SetItems(a.vm.Tests)
	a.history.SetItems(a.vm.Runs)
}

// HandleKey processes vim-like keybindings.
// Returns true if the key was handled.
func (a *App) HandleKey(key string) bool {
	switch key {
	case "j":
		a.tests.Move(1)
		return true
	case "k":
		a.tests.Move(-1)
		return true
	case "h":
		a.tree.Move(-1)
		return true
	case "l":
		a.tree.Move(1)
		return true
	case "G":
		a.end()
		return true
	case "gg":
		a.start()
		return true
	case "Enter":
		_ = a.runSelected()
		return true
	case "r":
		_ = a.rerunLast()
		return true
	case "/":
		a.ApplyFilter("")
		a.log.Append("filter reset")
		return true
	case "q":
		// caller decides to quit
		return true
	default:
		return false
	}
}

func (a *App) start() {
	a.tree.Move(-9999)
	a.tests.Move(-9999)
}

func (a *App) end() {
	a.tree.Move(9999)
	a.tests.Move(9999)
}

func (a *App) runSelected() error {
	pkgIdx := a.tree.Current()
	if pkgIdx < 0 || pkgIdx >= len(a.vm.Packages) {
		return errors.New("no package selected")
	}
	testIdx := a.tests.Current()
	var testName string
	if testIdx >= 0 && testIdx < len(a.vm.Tests) {
		testName = a.vm.Tests[testIdx].Name
	}
	run, err := a.runner.Run(context.Background(), a.vm.Packages[pkgIdx].Path, testName)
	if err != nil {
		a.log.Append(err.Error())
	}
	a.runs = append(a.runs, run)
	a.refreshRuns()
	for _, ln := range run.Logs() {
		a.log.Append(ln)
	}
	return err
}

func (a *App) rerunLast() error {
	if len(a.runs) == 0 {
		return errors.New("no previous run")
	}
	last := a.runs[len(a.runs)-1]
	run, err := a.runner.Run(context.Background(), last.PackagePath(), last.Target())
	if err != nil {
		a.log.Append(err.Error())
	}
	a.runs = append(a.runs, run)
	a.refreshRuns()
	for _, ln := range run.Logs() {
		a.log.Append(ln)
	}
	return err
}

func (a *App) refreshRuns() {
	vm := presentation.Project(a.snapshot, a.runs)
	a.history.SetItems(vm.Runs)
	a.history.ScrollToEnd()
}

func (a *App) ApplyFilter(term string) {
	a.filter = term
	if term == "" {
		a.tests.SetItems(a.vm.Tests)
		return
	}
	filtered := make([]presentation.TestRow, 0, len(a.vm.Tests))
	for _, tr := range a.vm.Tests {
		if strings.Contains(tr.Name, term) {
			filtered = append(filtered, tr)
		}
	}
	a.tests.SetItems(filtered)
}

// Runs returns a copy of run history (primarily for tests).
func (a *App) Runs() []execution.Run {
	cp := make([]execution.Run, len(a.runs))
	copy(cp, a.runs)
	return cp
}
