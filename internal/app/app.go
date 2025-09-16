package app

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"lazygotest/internal/config"
	"lazygotest/internal/core"
	"lazygotest/internal/discovery"
	"lazygotest/internal/ui"
	"lazygotest/pkg/util"
)

// App は Bubble Tea の Model 実装。
// How: Update-View ループで状態を管理し、非同期メッセージで Runner/Discovery と連携する。
// Why not: MVP では 3 ペインの見た目のみ先に用意し、機能は段階的に差し込む。
type App struct {
	cfg    config.Config
	ctx    context.Context
	cancel func()

	keys  ui.KeyMap
	focus int // 0:list 1:flags 2:logs

	// ディスカバリ結果/表示
	pkgs   []PkgView
	rows   []row
	cursor int

	listLines []string
	flagsLine string
	logsLines []string

	// ランナー
	runner    core.Runner
	runCh     chan core.TestEvent
	runCancel context.CancelFunc
	running   bool

	// 再実行用
	lastPkgs    []string
	lastRunExpr string
	curFails    *core.FailSet
}

func New(cfg config.Config) *App {
	ctx, cancel := context.WithCancel(context.Background())
	a := &App{
		cfg:       cfg,
		ctx:       ctx,
		cancel:    cancel,
		keys:      ui.DefaultKeyMap(),
		listLines: []string{"scanning packages..."},
		flagsLine: "[a:all r:run j/k:move q:quit]",
		logsLines: nil,
	}
	return a
}

func (a *App) Init() tea.Cmd { return a.discoverCmd() }

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyMsg:
		switch {
		case keyMatches(m, a.keys.Quit):
			a.cancel()
			return a, tea.Quit
		case keyMatches(m, a.keys.FocusNext):
			a.focus = (a.focus + 1) % 3
		case keyMatches(m, a.keys.FocusPrev):
			a.focus = (a.focus + 2) % 3
		case m.String() == "j":
			if a.cursor+1 < len(a.rows) {
				a.cursor++
			}
			a.refreshList()
		case m.String() == "k":
			if a.cursor > 0 {
				a.cursor--
			}
			a.refreshList()
		case m.String() == "a":
			if len(a.pkgs) == 0 || a.running {
				return a, nil
			}
			pkgs := make([]string, 0, len(a.pkgs))
			for _, p := range a.pkgs {
				pkgs = append(pkgs, p.ImportPath)
			}
			return a, a.startRun(pkgs, "")
		case m.String() == "r":
			if len(a.rows) == 0 || a.running {
				return a, nil
			}
			rw := a.rows[a.cursor]
			if rw.kind == rowPkg {
				return a, a.startRun([]string{rw.pkg}, "")
			}
			if rw.kind == rowTest {
				return a, a.startRun([]string{rw.pkg}, "^("+escapeRegex(rw.test)+")$")
			}
		case m.String() == "R":
			if a.running {
				return a, nil
			}
			pkgs, expr := a.buildFailedRerun()
			if len(pkgs) == 0 || expr == "" {
				a.appendLog("no failed tests from last run")
				return a, nil
			}
			return a, a.startRun(pkgs, expr)
		case m.String() == ".":
			if a.running {
				return a, nil
			}
			if len(a.lastPkgs) == 0 {
				a.appendLog("no previous run to repeat")
				return a, nil
			}
			return a, a.startRun(append([]string(nil), a.lastPkgs...), a.lastRunExpr)
		}
	case msgDiscoveryDone:
		if m.Err != nil {
			a.listLines = []string{"discovery error: " + m.Err.Error()}
			return a, nil
		}
		a.pkgs = m.Pkgs
		a.rebuildRows()
		a.refreshList()
		return a, nil
	case msgTestEvent:
		a.consumeEvent(m.Ev)
		return a, a.waitNext()
	case msgRunFinished:
		a.running = false
		a.runCancel = nil
		a.runCh = nil
		if a.curFails != nil {
			_ = a.curFails.Save()
		}
		return a, nil
	}
	return a, nil
}

func (a *App) View() string {
	// 3ペインを簡易レンダリング（本番は ui コンポーネントを分割）
	b := &strings.Builder{}
	fmt.Fprintln(b, "┌ Packages / Tests ────────────────────────────────────────────────────────────┐")
	for _, ln := range a.listLines {
		fmt.Fprintf(b, "│  %s\n", ln)
	}
	fmt.Fprintln(b, "└──────────────────────────────────────────────────────────────────────────────┘")
	fmt.Fprintln(b, "┌ Run/Filter [a r R / t g w c] ────────────────────────────────────────────────┐")
	fmt.Fprintf(b, "│ %s  running:%v\n", a.flagsLine, a.running)
	fmt.Fprintln(b, "└──────────────────────────────────────────────────────────────────────────────┘")
	fmt.Fprintln(b, "┌ Logs (selected test) ────────────────────────────────────────────────────────┐")
	if len(a.logsLines) == 0 {
		fmt.Fprintln(b, "│ (no logs)")
	}
	for _, ln := range a.logsLines {
		fmt.Fprintf(b, "│ %s\n", ln)
	}
	fmt.Fprintln(b, "└──────────────────────────────────────────────────────────────────────────────┘")
	return b.String()
}

// Run はアプリを起動する。外から呼ぶことで main から UI 依存を隔離する。
func Run(cfg config.Config) error {
	p := tea.NewProgram(New(cfg), tea.WithAltScreen())
	return p.Start()
}

func keyMatches(k tea.KeyMsg, keys []string) bool {
	s := k.String()
	for _, kk := range keys {
		if s == kk {
			return true
		}
	}
	return false
}

// ----- 内部: ディスカバリ -----

type PkgView struct {
	ImportPath string
	Tests      []string
}

type rowKind int

const (
	rowPkg rowKind = iota
	rowTest
)

type row struct {
	kind rowKind
	pkg  string
	test string
}

type msgDiscoveryDone struct {
	Pkgs []PkgView
	Err  error
}
type msgTestEvent struct{ Ev core.TestEvent }
type msgRunFinished struct{}

func (a *App) discoverCmd() tea.Cmd {
	patterns := a.cfg.Packages
	if len(patterns) == 0 {
		patterns = []string{"./..."}
	}
	return func() tea.Msg {
		pkgs, err := discovery.ListPackages(a.ctx, patterns)
		if err != nil {
			return msgDiscoveryDone{Err: err}
		}
		out := make([]PkgView, 0, len(pkgs))
		for _, p := range pkgs {
			names, _ := discovery.ListTests(a.ctx, p.ImportPath)
			out = append(out, PkgView{ImportPath: p.ImportPath, Tests: names})
		}
		return msgDiscoveryDone{Pkgs: out}
	}
}

func (a *App) rebuildRows() {
	rows := make([]row, 0, 64)
	for _, p := range a.pkgs {
		rows = append(rows, row{kind: rowPkg, pkg: p.ImportPath})
		for _, t := range p.Tests {
			rows = append(rows, row{kind: rowTest, pkg: p.ImportPath, test: t})
		}
	}
	a.rows = rows
	if a.cursor >= len(a.rows) {
		a.cursor = len(a.rows) - 1
	}
	if a.cursor < 0 {
		a.cursor = 0
	}
}

func (a *App) refreshList() {
	lines := make([]string, 0, len(a.rows))
	for i, r := range a.rows {
		cur := " "
		if i == a.cursor {
			cur = ">"
		}
		switch r.kind {
		case rowPkg:
			lines = append(lines, fmt.Sprintf("%s %s", cur, r.pkg))
		case rowTest:
			lines = append(lines, fmt.Sprintf("%s   %s", cur, r.test))
		}
	}
	if len(lines) == 0 {
		lines = []string{"(no packages)"}
	}
	a.listLines = lines
}

// ----- 内部: 実行 -----

func (a *App) startRun(pkgs []string, runRegex string) tea.Cmd {
	if a.running {
		return nil
	}
	a.running = true
	a.logsLines = nil
	a.runCh = make(chan core.TestEvent, 256)
	ctx, cancel := context.WithCancel(a.ctx)
	a.runCancel = cancel
	flags := core.RunFlags{Race: a.cfg.Race, Cover: a.cfg.Cover, Short: a.cfg.Short, Tags: a.cfg.Tags, Timeout: a.cfg.Timeout.String()}
	a.lastPkgs = append([]string(nil), pkgs...)
	a.lastRunExpr = runRegex
	a.curFails = core.NewFailSet()
	go func() {
		_ = a.runner.Run(ctx, pkgs, runRegex, flags, a.runCh)
		close(a.runCh)
	}()
	return a.waitNext()
}

func (a *App) waitNext() tea.Cmd {
	return func() tea.Msg {
		if a.runCh == nil {
			return msgRunFinished{}
		}
		ev, ok := <-a.runCh
		if !ok {
			return msgRunFinished{}
		}
		return msgTestEvent{Ev: ev}
	}
}

func (a *App) consumeEvent(ev core.TestEvent) {
	switch ev.Action {
	case "output":
		a.appendLog(strings.TrimRight(ev.Output, "\n"))
	case "run":
		if ev.Test != "" {
			a.appendLog(fmt.Sprintf("=== %s %s", ev.Package, ev.Test))
		}
	case "pass":
		if ev.Test != "" {
			a.appendLog(fmt.Sprintf("PASS %s %s (%.2fs)", ev.Package, ev.Test, ev.Elapsed))
		}
	case "fail":
		if ev.Test != "" {
			a.appendLog(fmt.Sprintf("FAIL %s %s (%.2fs)", ev.Package, ev.Test, ev.Elapsed))
			if a.curFails != nil {
				a.curFails.Add(ev.Package, ev.Test)
			}
		}
	}
}

func (a *App) appendLog(line string) {
	const max = 400
	a.logsLines = append(a.logsLines, line)
	if len(a.logsLines) > max {
		a.logsLines = a.logsLines[len(a.logsLines)-max:]
	}
}

func escapeRegex(s string) string {
	repl := []struct{ old, new string }{
		{"\\", "\\\\"}, {"(", "\\("}, {")", "\\)"}, {"[", "\\["}, {"]", "\\]"}, {".", "\\."}, {"+", "\\+"}, {"*", "\\*"}, {"^", "\\^"}, {"$", "\\$"}, {"|", "\\|"},
	}
	out := s
	for _, r := range repl {
		out = strings.ReplaceAll(out, r.old, r.new)
	}
	return out
}

func (a *App) buildFailedRerun() ([]string, string) {
	f := a.curFails
	if f == nil || len(f.Items) == 0 {
		if lf, err := core.LoadFailSet(); err == nil {
			f = lf
		}
	}
	if f == nil || len(f.Items) == 0 {
		return nil, ""
	}
	pkgs := make([]string, 0, len(f.Items))
	names := make([]string, 0, 32)
	for p, m := range f.Items {
		pkgs = append(pkgs, p)
		for name := range m {
			names = append(names, name)
		}
	}
	return pkgs, util.BuildRunRegex(names)
}
