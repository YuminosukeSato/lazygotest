package ui_test

import (
	"context"
	"testing"

	"github.com/s21066/lazygotest/internal/application/testrun"
	"github.com/s21066/lazygotest/internal/domain/catalog"
	"github.com/s21066/lazygotest/internal/domain/execution"
	"github.com/s21066/lazygotest/internal/interfaces/process"
	"github.com/s21066/lazygotest/internal/presentation"
	"github.com/s21066/lazygotest/internal/ui"
)

type fakeTree struct {
	items []presentation.PackageNode
	cur   int
}

func (f *fakeTree) SetItems(items []presentation.PackageNode) { f.items = items }
func (f *fakeTree) Move(delta int) {
	f.cur += delta
	if f.cur < 0 {
		f.cur = 0
	}
	if f.cur >= len(f.items) {
		f.cur = len(f.items) - 1
	}
}
func (f *fakeTree) Current() int { return f.cur }

type fakeList struct {
	items []presentation.TestRow
	cur   int
}

func (f *fakeList) SetItems(items []presentation.TestRow) { f.items = items }
func (f *fakeList) Move(delta int) {
	f.cur += delta
	if f.cur < 0 {
		f.cur = 0
	}
	if f.cur >= len(f.items) {
		f.cur = len(f.items) - 1
	}
}
func (f *fakeList) Current() int { return f.cur }

type fakeHistory struct {
	items    []presentation.RunRow
	scrolled bool
}

func (f *fakeHistory) SetItems(items []presentation.RunRow) { f.items = items }
func (f *fakeHistory) ScrollToEnd()                         { f.scrolled = true }

type fakeLog struct {
	lines []string
}

func (f *fakeLog) Append(line string) { f.lines = append(f.lines, line) }
func (f *fakeLog) Clear()             { f.lines = nil }

type fakeRunner struct {
	res process.TestResult
	err error
}

func (f *fakeRunner) Run(ctx context.Context, pkg, test string) (process.TestResult, error) {
	return f.res, f.err
}

func buildSnapshot(t *testing.T, pkg string, tests []string) catalog.Snapshot {
	t.Helper()
	tcs := make([]catalog.TestCase, 0, len(tests))
	for _, name := range tests {
		tc, err := catalog.NewTestCase(name, "x_test.go", 1, catalog.TestKindUnit)
		if err != nil {
			t.Fatalf("new test case: %v", err)
		}
		tcs = append(tcs, tc)
	}
	p, err := catalog.NewPackage(pkg, "/repo", tcs)
	if err != nil {
		t.Fatalf("new package: %v", err)
	}
	m, err := catalog.NewModule("mod", "/repo", []catalog.Package{p})
	if err != nil {
		t.Fatalf("new module: %v", err)
	}
	return catalog.NewSnapshot([]catalog.Module{m})
}

func TestAppHandlesEnterRunAndUpdatesHistory(t *testing.T) {
	t.Parallel()
	tree := &fakeTree{}
	list := &fakeList{}
	hist := &fakeHistory{}
	log := &fakeLog{}
	run := execution.NewRun("github.com/me/a", "TestFoo")
	run.AddDuration(0.2)
	run.AppendLog("ok\n")
	run.Complete(nil)

	runner := testrun.NewService(&fakeRunner{res: process.TestResult{
		Status:   process.RunStatusPass,
		Duration: run.Duration(),
		Events: []process.TestEvent{
			{Output: "ok\n"},
		},
	}})
	app := ui.NewApp(tree, list, hist, log, runner)
	snap := buildSnapshot(t, "github.com/me/a", []string{"TestFoo"})
	app.SetSnapshot(snap, nil)

	if handled := app.HandleKey("Enter"); !handled {
		t.Fatalf("expected Enter handled")
	}
	if len(hist.items) != 1 {
		t.Fatalf("expected history updated")
	}
	if !hist.scrolled {
		t.Fatalf("expected history scrolled")
	}
	if len(log.lines) == 0 {
		t.Fatalf("expected log appended")
	}
}

func TestAppRerunLast(t *testing.T) {
	t.Parallel()
	tree := &fakeTree{}
	list := &fakeList{}
	hist := &fakeHistory{}
	log := &fakeLog{}
	first := execution.NewRun("pkg", "TestFoo")
	first.Complete(nil)
	rerun := execution.NewRun("pkg", "TestFoo")
	rerun.AddDuration(0.1)
	rerun.Complete(nil)
	runner := &fakeRunner{res: process.TestResult{
		Status:   process.RunStatusPass,
		Duration: rerun.Duration(),
	}}
	app := ui.NewApp(tree, list, hist, log, testrun.NewService(runner))
	app.SetSnapshot(buildSnapshot(t, "pkg", []string{"TestFoo"}), nil)
	app.HandleKey("Enter") // initial run
	runs := app.Runs()
	if len(runs) == 0 {
		t.Fatalf("expected one run")
	}
	appRuns := app.Runs()
	appRuns[0] = first // mutate copy to simulate last run; original remains

	app.HandleKey("r")
	if len(hist.items) < 1 {
		t.Fatalf("expected history set")
	}
}
