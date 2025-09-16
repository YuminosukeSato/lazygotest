package app

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"lazygotest/internal/config"
	"lazygotest/internal/core"
	"lazygotest/internal/discovery"
	"lazygotest/internal/ui"
	"lazygotest/pkg/util"
)

var (
	colorBaseBg  = lipgloss.Color("#1c1c24")
	colorBaseFg  = lipgloss.Color("#d7d7ff")
	colorBorder  = lipgloss.Color("#5f5f87")
	colorFocus   = lipgloss.Color("#87afff")
	colorPass    = lipgloss.Color("#87d787")
	colorFail    = lipgloss.Color("#ff5f5f")
	colorSkip    = lipgloss.Color("#ffd75f")
	colorRunning = lipgloss.Color("#5fafff")
	colorMuted   = lipgloss.Color("#8a8aa3")
	colorChipOn  = lipgloss.Color("#303044")
	colorToast   = lipgloss.Color("#2b2b3b")
)

var (
	baseStyle        = lipgloss.NewStyle().Foreground(colorBaseFg).Background(colorBaseBg)
	headerStyle      = lipgloss.NewStyle().Foreground(colorBaseFg).Bold(true)
	borderNormal     = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colorBorder).Padding(0, 1)
	borderFocused    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colorFocus).Padding(0, 1)
	chipOnStyle      = lipgloss.NewStyle().Foreground(colorBaseFg).Background(colorChipOn).Padding(0, 1).MarginRight(1)
	chipOffStyle     = lipgloss.NewStyle().Foreground(colorMuted).Border(lipgloss.RoundedBorder()).BorderForeground(colorMuted).Padding(0, 1).MarginRight(1)
	statusBarStyle   = lipgloss.NewStyle().Foreground(colorBaseFg).Background(lipgloss.Color("#26263a")).Padding(0, 1)
	toastStyle       = lipgloss.NewStyle().Foreground(colorBaseFg).Background(colorToast).Padding(0, 1).MarginLeft(2)
	mutedStyle       = lipgloss.NewStyle().Foreground(colorMuted)
	selectedRowStyle = lipgloss.NewStyle().Foreground(colorBaseFg).Background(lipgloss.Color("#3a3a55")).Bold(true)
	passStyle        = lipgloss.NewStyle().Foreground(colorPass)
	failStyle        = lipgloss.NewStyle().Foreground(colorFail)
	skipStyle        = lipgloss.NewStyle().Foreground(colorSkip)
	runningStyle     = lipgloss.NewStyle().Foreground(colorRunning)
)

type panel int

const (
	panelTests panel = iota
	panelFlags
	panelLogs
)

type toastTick struct{}

type App struct {
	cfg    config.Config
	ctx    context.Context
	cancel func()

	keys  ui.KeyMap
	help  help.Model
	focus panel

	width  int
	height int

	table    table.Model
	filter   textinput.Model
	logs     viewport.Model
	selected map[int]struct{}
	visible  []int
	cursor   int

	pkgs []PkgView
	rows []row

	failOnly bool

	race  bool
	cover bool
	short bool
	bench bool
	fuzz  bool
	watch bool

	tags []string

	toastMsg   string
	toastUntil time.Time

	runner       core.Runner
	runCh        chan core.TestEvent
	runCancel    context.CancelFunc
	running      bool
	runStart     time.Time
	lastDuration time.Duration

	passCount int
	failCount int
	skipCount int

	statusPkg   string
	statusIndex int
	statusTotal int
	runSeen     map[string]struct{}

	lastPkgs    []string
	lastRunExpr string
	curFails    *core.FailSet

	logLines []string
}

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
	kind     rowKind
	pkg      string
	test     string
	status   string
	duration time.Duration
	note     string
}

type msgDiscoveryDone struct {
	Pkgs []PkgView
	Err  error
}

type msgTestEvent struct{ Ev core.TestEvent }

type msgRunFinished struct{}

func tableStyles() table.Styles {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorBorder).
		BorderBottom(true).
		Bold(true).
		Foreground(colorBaseFg)
	s.Selected = selectedRowStyle
	s.Cell = lipgloss.NewStyle().Foreground(colorBaseFg)
	return s
}

func New(cfg config.Config) *App {
	ctx, cancel := context.WithCancel(context.Background())

	cols := []table.Column{
		{Title: " ", Width: 2},
		{Title: "NAME", Width: 38},
		{Title: "DUR", Width: 7},
		{Title: "STATUS", Width: 8},
		{Title: "NOTE", Width: 30},
	}

	tbl := table.New(
		table.WithColumns(cols),
		table.WithRows([]table.Row{{"▸", "scanning packages...", "", "", ""}}),
		table.WithFocused(true),
		table.WithHeight(12),
	)
	tbl.SetStyles(tableStyles())

	filter := textinput.New()
	filter.Prompt = "/"
	filter.Placeholder = "Div|Add"
	filter.CharLimit = 120
	filter.PromptStyle = mutedStyle
	filter.TextStyle = lipgloss.NewStyle().Foreground(colorBaseFg)
	filter.PlaceholderStyle = mutedStyle

	logs := viewport.New(0, 0)
	logs.SetContent("(no logs)")

	h := help.New()
	h.ShowAll = false

	return &App{
		cfg:      cfg,
		ctx:      ctx,
		cancel:   cancel,
		keys:     ui.DefaultKeyMap(),
		help:     h,
		focus:    panelTests,
		table:    tbl,
		filter:   filter,
		logs:     logs,
		selected: map[int]struct{}{},
		visible:  make([]int, 0),
		race:     cfg.Race,
		cover:    cfg.Cover,
		short:    cfg.Short,
		watch:    cfg.Watch,
		tags:     append([]string(nil), cfg.Tags...),
	}
}

func (a *App) Init() tea.Cmd {
	return a.discoverCmd()
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		a.resize(m.Width, m.Height)
		return a, nil
	case tea.KeyMsg:
		cmd := a.handleKeyMsg(m)
		return a, cmd
	case msgDiscoveryDone:
		a.handleDiscovery(m)
		return a, nil
	case msgTestEvent:
		a.consumeEvent(m.Ev)
		return a, a.waitNext()
	case msgRunFinished:
		a.finishRun()
		return a, nil
	case toastTick:
		if time.Now().After(a.toastUntil) {
			a.toastMsg = ""
			return a, nil
		}
		return a, a.scheduleToastTick()
	}
	return a, nil
}

func (a *App) View() string {
	tests := a.renderTests()
	flags := a.renderFlags()
	logs := a.renderLogs()
	status := a.renderStatus()

	view := lipgloss.JoinVertical(lipgloss.Left, tests, flags, logs, status)
	if a.toastMsg != "" {
		toast := toastStyle.Render(a.toastMsg)
		view = lipgloss.JoinVertical(lipgloss.Left, toast, view)
	}
	return baseStyle.Padding(0, 1).Render(view)
}

func Run(cfg config.Config) error {
	p := tea.NewProgram(New(cfg), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func (a *App) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	if key.Matches(msg, a.keys.Quit) {
		a.cancel()
		return tea.Quit
	}
	if key.Matches(msg, a.keys.Help) {
		a.help.ShowAll = !a.help.ShowAll
		return nil
	}

	if a.filter.Focused() {
		if msg.Type == tea.KeyEsc {
			a.filter.Blur()
			return nil
		}
		var cmd tea.Cmd
		a.filter, cmd = a.filter.Update(msg)
		a.refreshTable()
		return cmd
	}

	switch {
	case key.Matches(msg, a.keys.Next):
		a.cycleFocus(1)
		return nil
	case key.Matches(msg, a.keys.Prev):
		a.cycleFocus(-1)
		return nil
	case key.Matches(msg, a.keys.RunAll):
		return a.runAll()
	case key.Matches(msg, a.keys.RunFailed):
		return a.runFailed()
	case key.Matches(msg, a.keys.Repeat):
		return a.repeatLastRun()
	case key.Matches(msg, a.keys.Run):
		return a.runSelection()
	case key.Matches(msg, a.keys.Filter):
		a.focus = panelTests
		a.filter.Focus()
		a.filter.SetCursor(len(a.filter.Value()))
		return nil
	case key.Matches(msg, a.keys.FailOnly):
		a.failOnly = !a.failOnly
		a.refreshTable()
		return nil
	case key.Matches(msg, a.keys.Race):
		a.race = !a.race
		a.cfg.Race = a.race
		return a.toast(labelToggle("-race", a.race))
	case key.Matches(msg, a.keys.Cover):
		a.cover = !a.cover
		a.cfg.Cover = a.cover
		return a.toast(labelToggle("-cover", a.cover))
	case key.Matches(msg, a.keys.Bench):
		a.bench = !a.bench
		return a.toast(labelToggle("bench", a.bench))
	case key.Matches(msg, a.keys.Fuzz):
		a.fuzz = !a.fuzz
		return a.toast(labelToggle("fuzz", a.fuzz))
	case key.Matches(msg, a.keys.Tags):
		return a.toast("Tag input is not implemented yet")
	case key.Matches(msg, a.keys.Watch):
		a.watch = !a.watch
		return a.toast(labelToggle("watch", a.watch))
	case key.Matches(msg, a.keys.Open):
		return a.toast("Editor integration is not implemented yet")
	case key.Matches(msg, a.keys.Save):
		return a.toast("Log save is not implemented yet")
	case key.Matches(msg, a.keys.Search):
		return a.toast("Log search is not implemented yet")
	case key.Matches(msg, a.keys.FocusLogs):
		a.focus = panelLogs
		return nil
	}

	switch a.focus {
	case panelTests:
		return a.handleTestsKey(msg)
	case panelLogs:
		var cmd tea.Cmd
		a.logs, cmd = a.logs.Update(msg)
		return cmd
	}
	return nil
}

func (a *App) handleTestsKey(msg tea.KeyMsg) tea.Cmd {
	if key.Matches(msg, a.keys.Select) {
		idx, ok := a.currentRowIndex()
		if !ok {
			return nil
		}
		if _, exists := a.selected[idx]; exists {
			delete(a.selected, idx)
		} else {
			a.selected[idx] = struct{}{}
		}
		a.refreshTable()
		return nil
	}
	if msg.String() == "h" || msg.String() == "l" {
		a.toast("Collapse is not implemented yet")
		return nil
	}
	var cmd tea.Cmd
	a.table, cmd = a.table.Update(msg)
	a.cursor = a.table.Cursor()
	return cmd
}

func (a *App) currentRowIndex() (int, bool) {
	if len(a.visible) == 0 {
		return 0, false
	}
	cur := a.table.Cursor()
	if cur < 0 || cur >= len(a.visible) {
		return 0, false
	}
	return a.visible[cur], true
}

func (a *App) handleDiscovery(m msgDiscoveryDone) {
	if m.Err != nil {
		a.pkgs = nil
		a.rows = nil
		a.table.SetRows([]table.Row{{"", failStyle.Render("discovery error: " + m.Err.Error()), "", "", ""}})
		a.visible = nil
		return
	}
	a.pkgs = m.Pkgs
	a.rebuildRows()
	a.refreshTable()
}

func (a *App) rebuildRows() {
	rows := make([]row, 0, len(a.pkgs)*8)
	for _, p := range a.pkgs {
		rows = append(rows, row{kind: rowPkg, pkg: p.ImportPath})
		for _, t := range p.Tests {
			rows = append(rows, row{kind: rowTest, pkg: p.ImportPath, test: t})
		}
	}
	a.rows = rows
	a.cursor = 0
}

func (a *App) refreshTable() {
	filter := strings.TrimSpace(a.filter.Value())
	rows := make([]table.Row, 0, len(a.rows))
	visible := make([]int, 0, len(a.rows))
	lower := strings.ToLower(filter)

	for idx, r := range a.rows {
		if a.failOnly && r.status != "FAIL" {
			continue
		}
		if filter != "" && !rowMatches(r, lower) {
			continue
		}
		rows = append(rows, a.rowToTableRow(idx, r))
		visible = append(visible, idx)
	}

	if len(rows) == 0 {
		rows = []table.Row{{"", mutedStyle.Render("(no matches)"), "", "", ""}}
		visible = visible[:0]
		a.table.SetRows(rows)
		a.table.SetCursor(0)
		a.visible = visible
		return
	}

	a.visible = visible
	a.table.SetRows(rows)
	if a.cursor >= len(a.visible) {
		a.cursor = len(a.visible) - 1
	}
	if a.cursor < 0 {
		a.cursor = 0
	}
	a.table.SetCursor(a.cursor)
}

func rowMatches(r row, filter string) bool {
	if filter == "" {
		return true
	}
	if strings.Contains(strings.ToLower(r.pkg), filter) {
		return true
	}
	if strings.Contains(strings.ToLower(r.test), filter) {
		return true
	}
	return false
}

func (a *App) rowToTableRow(idx int, r row) table.Row {
	indicator := " "
	if _, ok := a.selected[idx]; ok {
		indicator = "●"
	} else if r.kind == rowPkg {
		indicator = "▸"
	} else {
		switch r.status {
		case "PASS":
			indicator = "✓"
		case "FAIL":
			indicator = "✗"
		case "SKIP":
			indicator = "◇"
		default:
			indicator = " "
		}
	}

	name := r.pkg
	if r.kind == rowTest {
		name = "  " + r.test
	}

	duration := formatDuration(r.duration)
	status := renderStatusLabel(r.status)
	note := r.note

	return table.Row{indicator, name, duration, status, note}
}

func formatDuration(d time.Duration) string {
	if d <= 0 {
		return ""
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

func renderStatusLabel(status string) string {
	switch status {
	case "PASS":
		return passStyle.Render("PASS")
	case "FAIL":
		return failStyle.Render("FAIL")
	case "SKIP":
		return skipStyle.Render("SKIP")
	default:
		return ""
	}
}

func (a *App) renderTests() string {
	width := max(a.width - 2)

	chips := boolChip("failed", a.failOnly) + boolChip("short", a.short)
	filterLine := ""
	if a.filter.Focused() {
		filterLine = lipgloss.JoinHorizontal(lipgloss.Left, a.filter.View(), "  ", chips)
	} else {
		fv := a.filter.Value()
		if fv == "" {
			filterLine = lipgloss.JoinHorizontal(lipgloss.Left, mutedStyle.Render("filter: (/)"), "  ", chips)
		} else {
			filterLine = lipgloss.JoinHorizontal(lipgloss.Left, fmt.Sprintf("filter: /%s/", fv), "  ", chips)
		}
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		headerStyle.Render(" TESTS (j/k move, Space select, / filter, f failed-only, ? help) "),
		a.table.View(),
		filterLine,
		a.help.View(a.keys),
	)

	box := borderNormal
	if a.focus == panelTests {
		box = borderFocused
	}
	return box.Width(width).Render(content)
}

func (a *App) renderFlags() string {
	width := max(a.width - 2)

	runExpr := "(all)"
	if a.lastRunExpr != "" {
		runExpr = fmt.Sprintf("/%s/", a.lastRunExpr)
	}
	tags := "(none)"
	if len(a.tags) > 0 {
		tags = strings.Join(a.tags, ",")
	}

	lines := []string{
		headerStyle.Render(" FLAGS (t tags g:-race c cover b bench z fuzz w watch o open-editor) "),
		fmt.Sprintf("  -run=%s   -tags=%s   -race=%s   -cover=%s   -short=%s",
			mutedStyle.Render(runExpr),
			mutedStyle.Render(tags),
			mutedStyle.Render(onOff(a.race)),
			mutedStyle.Render(onOff(a.cover)),
			mutedStyle.Render(onOff(a.short)),
		),
		"  " + strings.Join([]string{
			boolChip("race", a.race),
			boolChip("cover", a.cover),
			boolChip("bench", a.bench),
			boolChip("fuzz", a.fuzz),
			boolChip("watch", a.watch),
		}, ""),
	}

	box := borderNormal
	if a.focus == panelFlags {
		box = borderFocused
	}
	return box.Width(width).Render(strings.Join(lines, "\n"))
}

func (a *App) renderLogs() string {
	width := max(a.width - 2)
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		headerStyle.Render(" LOGS (Enter/Tab focus, ↑↓ scroll, s save, C-f search) "),
		a.logs.View(),
	)
	box := borderNormal
	if a.focus == panelLogs {
		box = borderFocused
	}
	return box.Width(width).Render(content)
}

func (a *App) renderStatus() string {
	width := max(a.width - 2)

	indicator := runningStyle.Render("● Running")
	duration := a.lastDuration
	if a.running {
		duration = time.Since(a.runStart)
	}
	if !a.running {
		indicator = mutedStyle.Render("○ Idle")
	}
	pkgLabel := "pkg:- (0/0)"
	if a.statusTotal > 0 {
		pkg := a.statusPkg
		if pkg == "" {
			pkg = "-"
		}
		pkgLabel = fmt.Sprintf("pkg:%s (%d/%d)", pkg, a.statusIndex, a.statusTotal)
	}
	durLabel := "⏱ --"
	if duration > 0 {
		durLabel = fmt.Sprintf("⏱ %s", duration.Round(100*time.Millisecond))
	}

	left := strings.Join([]string{
		indicator,
		pkgLabel,
		fmt.Sprintf("%s:%d %s:%d %s:%d",
			passStyle.Render("PASS"), a.passCount,
			failStyle.Render("FAIL"), a.failCount,
			skipStyle.Render("SKIP"), a.skipCount,
		),
		durLabel,
	}, " | ")
	right := mutedStyle.Render("R:run failed  .:repeat last  q:quit")

	return statusBarStyle.Width(width).Render(
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			left,
			lipgloss.PlaceHorizontal(width, 1.0, right),
		),
	)
}

func boolChip(label string, on bool) string {
	if on {
		return chipOnStyle.Render(label)
	}
	return chipOffStyle.Render(label)
}

func onOff(v bool) string {
	if v {
		return "ON"
	}
	return "OFF"
}

func labelToggle(label string, on bool) string {
	state := "OFF"
	if on {
		state = "ON"
	}
	return fmt.Sprintf("%s %s", label, state)
}

func (a *App) resize(width, height int) {
	if width <= 0 || height <= 0 {
		return
	}
	a.width = width
	a.height = height

	logsHeight := height / 3
	if logsHeight < 8 {
		logsHeight = 8
	}
	testsHeight := height - logsHeight - 6
	if testsHeight < 10 {
		testsHeight = 10
	}

	a.table.SetHeight(testsHeight - 4)
	a.logs.Width = max(width - 4)
	a.logs.Height = logsHeight - 3
	a.help.Width = max(width - 4)
}

func (a *App) cycleFocus(delta int) {
	a.focus = panel((int(a.focus) + delta + 3) % 3)
	if a.focus != panelTests {
		a.filter.Blur()
	}
}

func (a *App) runAll() tea.Cmd {
	if a.running || len(a.pkgs) == 0 {
		return a.toast("No packages to run")
	}
	pkgs := make([]string, 0, len(a.pkgs))
	for _, p := range a.pkgs {
		pkgs = append(pkgs, p.ImportPath)
	}
	return a.startRun(pkgs, "")
}

func (a *App) runSelection() tea.Cmd {
	if a.running {
		return nil
	}
	indices := make([]int, 0, len(a.selected))
	for idx := range a.selected {
		indices = append(indices, idx)
	}
	if len(indices) == 0 {
		idx, ok := a.currentRowIndex()
		if !ok {
			return nil
		}
		indices = append(indices, idx)
	}
	sort.Ints(indices)

	pkgSet := map[string]struct{}{}
	testsOnly := true
	testNames := make([]string, 0, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= len(a.rows) {
			continue
		}
		r := a.rows[idx]
		pkgSet[r.pkg] = struct{}{}
		if r.kind != rowTest {
			testsOnly = false
		} else {
			testNames = append(testNames, r.test)
		}
	}
	if len(pkgSet) == 0 {
		return a.toast("No selections found")
	}
	pkgs := make([]string, 0, len(pkgSet))
	for pkg := range pkgSet {
		pkgs = append(pkgs, pkg)
	}
	sort.Strings(pkgs)

	var regex string
	if testsOnly && len(testNames) > 0 {
		regex = util.BuildRunRegex(testNames)
	}
	return a.startRun(pkgs, regex)
}

func (a *App) runFailed() tea.Cmd {
	if a.running {
		return nil
	}
	pkgs, expr := a.buildFailedRerun()
	if len(pkgs) == 0 {
		return a.toast("No failed tests available")
	}
	return a.startRun(pkgs, expr)
}

func (a *App) repeatLastRun() tea.Cmd {
	if a.running {
		return nil
	}
	if len(a.lastPkgs) == 0 {
		return a.toast("No previous run")
	}
	return a.startRun(append([]string(nil), a.lastPkgs...), a.lastRunExpr)
}

func (a *App) startRun(pkgs []string, runRegex string) tea.Cmd {
	if len(pkgs) == 0 {
		return a.toast("No targets to run")
	}
	if a.running {
		return nil
	}

	a.running = true
	a.runStart = time.Now()
	a.lastDuration = 0
	a.passCount = 0
	a.failCount = 0
	a.skipCount = 0
	a.statusPkg = ""
	a.statusIndex = 0
	a.statusTotal = len(pkgs)
	a.runSeen = map[string]struct{}{}
	a.selected = map[int]struct{}{}
	a.logLines = nil
	a.logs.SetContent("(no logs)")

	for i := range a.rows {
		a.rows[i].status = ""
		a.rows[i].duration = 0
	}
	a.refreshTable()

	a.runCh = make(chan core.TestEvent, 256)
	ctx, cancel := context.WithCancel(a.ctx)
	a.runCancel = cancel
	a.curFails = core.NewFailSet()

	flags := core.RunFlags{
		Race:    a.race,
		Cover:   a.cover,
		Short:   a.short,
		Tags:    a.tags,
		Timeout: a.cfg.Timeout.String(),
	}

	a.lastPkgs = append([]string(nil), pkgs...)
	a.lastRunExpr = runRegex

	go func() {
		if err := a.runner.Run(ctx, pkgs, runRegex, flags, a.runCh); err != nil {
			// Log error but continue - channel will be closed anyway
			// TODO: Add proper error logging in production
			_ = err // Explicitly ignore for now
		}
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
	if ev.Package != "" {
		if a.runSeen == nil {
			a.runSeen = map[string]struct{}{}
		}
		if _, ok := a.runSeen[ev.Package]; !ok {
			a.runSeen[ev.Package] = struct{}{}
			a.statusIndex = len(a.runSeen)
		}
		a.statusPkg = ev.Package
	}

	switch ev.Action {
	case "output":
		line := strings.TrimRight(ev.Output, "\n")
		if line != "" {
			a.appendLog(line)
		}
	case "run":
		if ev.Test != "" {
			a.appendLog(fmt.Sprintf("=== %s · %s", ev.Package, ev.Test))
		}
	case "pass":
		if ev.Test != "" {
			a.passCount++
			a.updateRowStatus(ev.Package, ev.Test, "PASS", ev.Elapsed)
		}
	case "fail":
		if ev.Test != "" {
			a.failCount++
			a.updateRowStatus(ev.Package, ev.Test, "FAIL", ev.Elapsed)
			if a.curFails != nil {
				a.curFails.Add(ev.Package, ev.Test)
			}
		} else {
			a.appendLog(fmt.Sprintf("FAIL %s", ev.Package))
		}
	case "skip":
		if ev.Test != "" {
			a.skipCount++
			a.updateRowStatus(ev.Package, ev.Test, "SKIP", ev.Elapsed)
		}
	}
}

func (a *App) updateRowStatus(pkg, test, status string, elapsed float64) {
	dur := time.Duration(elapsed * float64(time.Second))
	for i := range a.rows {
		r := &a.rows[i]
		if r.pkg != pkg {
			continue
		}
		if r.kind == rowTest && r.test == test {
			r.status = status
			if elapsed > 0 {
				r.duration = dur
			}
			break
		}
	}
	a.refreshTable()
}

func (a *App) appendLog(line string) {
	const maxLogs = 400
	a.logLines = append(a.logLines, line)
	if len(a.logLines) > maxLogs {
		a.logLines = a.logLines[len(a.logLines)-maxLogs:]
	}
	a.logs.SetContent(strings.Join(a.logLines, "\n"))
	a.logs.GotoBottom()
}

func (a *App) finishRun() {
	a.running = false
	a.runCancel = nil
	a.runCh = nil
	if !a.runStart.IsZero() {
		a.lastDuration = time.Since(a.runStart)
	}
	if a.curFails != nil {
		if err := a.curFails.Save(); err != nil {
			// Log error but continue - non-critical failure
			// TODO: Add proper error logging in production
			_ = err // Explicitly ignore for now
		}
	}
}

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
			names, err := discovery.ListTests(a.ctx, p.ImportPath)
			if err != nil {
				// Continue with empty test list for this package
				names = nil
			}
			out = append(out, PkgView{ImportPath: p.ImportPath, Tests: names})
		}
		return msgDiscoveryDone{Pkgs: out}
	}
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

func (a *App) toast(message string) tea.Cmd {
	a.toastMsg = message
	a.toastUntil = time.Now().Add(1500 * time.Millisecond)
	return a.scheduleToastTick()
}

func (a *App) scheduleToastTick() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(time.Time) tea.Msg { return toastTick{} })
}

func max(v int) int {
	if v < 0 {
		return 0
	}
	return v
}
