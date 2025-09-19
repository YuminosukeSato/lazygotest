package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/YuminosukeSato/lazygotest/internal/adapter/secondary/pkgrepo"
	"github.com/YuminosukeSato/lazygotest/internal/adapter/secondary/runner"
	"github.com/YuminosukeSato/lazygotest/internal/domain"
	"github.com/YuminosukeSato/lazygotest/internal/shared/eventbus"
	"github.com/YuminosukeSato/lazygotest/internal/usecase"
	"github.com/YuminosukeSato/lazygotest/pkg/logger"
)

// FocusedPane represents which pane has focus
type FocusedPane int

const (
	PackagesPane FocusedPane = iota
	TestsPane
	DetailsPane
)

// Model represents the application state
type Model struct {
	// UI State
	focusedPane      FocusedPane
	width            int
	height           int
	packageList      list.Model
	testList         list.Model
	detailsContent   []string
	detailsScrollPos int    // Current scroll position in details pane
	detailsMaxScroll int    // Maximum scroll position
	lastKey          string // For multi-key commands like gg

	// Domain State
	packages        []*domain.Package
	selectedPackage *domain.Package
	selectedTest    *domain.TestCase
	testResults     map[domain.TestID]*domain.TestCase
	summary         *domain.TestSummary
	selectedTests   map[domain.TestID]bool // Track selected tests for batch execution

	// Dependencies
	listPkgsUC *usecase.ListPackagesUseCase
	runTestsUC *usecase.RunTestsUseCase
	eventBus   *eventbus.EventBus

	// Flags
	isRunning       bool
	showFailedOnly  bool
	watchMode       bool
	raceDetection   bool
	coverageEnabled bool

	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc
}

// New creates a new TUI application model
func New() *Model {
	// Initialize dependencies
	bus := eventbus.New(1000)
	pkgRepo := pkgrepo.NewGoPackageRepo()
	testRunner := runner.NewTestRunner()

	listPkgsUC := usecase.NewListPackagesUseCase(pkgRepo, bus)
	runTestsUC := usecase.NewRunTestsUseCase(testRunner, bus)

	ctx, cancel := context.WithCancel(context.Background())

	m := &Model{
		focusedPane:    PackagesPane,
		packages:       make([]*domain.Package, 0),
		testResults:    make(map[domain.TestID]*domain.TestCase),
		selectedTests:  make(map[domain.TestID]bool),
		detailsContent: make([]string, 0),
		listPkgsUC:     listPkgsUC,
		runTestsUC:     runTestsUC,
		eventBus:       bus,
		ctx:            ctx,
		cancel:         cancel,
	}

	// Initialize lists with custom styles
	packageDelegate := list.NewDefaultDelegate()
	packageDelegate.ShowDescription = true
	packageDelegate.Styles.SelectedTitle = packageDelegate.Styles.SelectedTitle.
		Background(lipgloss.Color("#3C3C3C")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)
	packageDelegate.Styles.SelectedDesc = packageDelegate.Styles.SelectedDesc.
		Background(lipgloss.Color("#3C3C3C")).
		Foreground(lipgloss.Color("#CCCCCC"))

	// Use custom test delegate with colored backgrounds
	testDelegate := list.NewDefaultDelegate()
	testDelegate.ShowDescription = true

	// Custom styles for test items based on status
	// These will be overridden by item-specific rendering
	testDelegate.Styles.SelectedTitle = testDelegate.Styles.SelectedTitle.
		Background(lipgloss.Color("#3C3C3C")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)
	testDelegate.Styles.SelectedDesc = testDelegate.Styles.SelectedDesc.
		Background(lipgloss.Color("#3C3C3C")).
		Foreground(lipgloss.Color("#CCCCCC"))

	m.packageList = list.New([]list.Item{}, packageDelegate, 0, 0)
	m.packageList.Title = "" // We handle title in render
	m.packageList.SetShowStatusBar(false)
	m.packageList.SetShowHelp(false)

	m.testList = list.New([]list.Item{}, testDelegate, 0, 0)
	m.testList.Title = "" // We handle title in render
	m.testList.SetShowStatusBar(false)
	m.testList.SetShowHelp(false)

	// Subscribe to events
	m.subscribeToEvents()

	return m
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	logger.Debug("Initializing TUI model")
	return tea.Batch(
		m.loadPackages(),
		tea.EnterAltScreen,
	)
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateSizes()

	case tea.KeyMsg:
		cmds = append(cmds, m.handleKeyPress(msg))

	case packagesLoadedMsg:
		m.packages = msg.packages
		m.updatePackageList()

	case testEventMsg:
		m.handleTestEvent(msg.event)

	case errorMsg:
		logger.Error("Error occurred", "error", msg.err)
		m.appendDetail("Error: " + msg.err.Error())
	}

	// Update the focused list
	switch m.focusedPane {
	case PackagesPane:
		newList, cmd := m.packageList.Update(msg)
		m.packageList = newList
		cmds = append(cmds, cmd)
	case TestsPane:
		newList, cmd := m.testList.Update(msg)
		m.testList = newList
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the UI
func (m *Model) View() string {
	return m.render()
}

// handleKeyPress handles keyboard input
func (m *Model) handleKeyPress(msg tea.KeyMsg) tea.Cmd {
	// Global keybindings (work in any pane)
	switch msg.String() {
	case "q", "ctrl+c":
		m.cancel()
		return tea.Quit

	case "tab", "l": // l for right movement
		m.focusNextPane()
		return nil

	case "shift+tab", "h": // h for left movement
		m.focusPrevPane()
		return nil

	case "enter":
		return m.handleEnter()

	case "f", "F":
		m.showFailedOnly = !m.showFailedOnly
		m.updateTestList()
		return nil

	case "r":
		if m.selectedTest != nil {
			return m.rerunTest()
		}

	case "R":
		m.raceDetection = !m.raceDetection
		return nil

	case "C":
		m.coverageEnabled = !m.coverageEnabled
		return nil

	case "W":
		m.watchMode = !m.watchMode
		return nil

	case " ": // Space key for selection toggle
		if m.focusedPane == TestsPane {
			return m.toggleTestSelection()
		}
		return nil

	case "a": // Select all visible tests or run all tests
		if m.focusedPane == TestsPane {
			return m.selectAllTests()
		}
		// Run all tests from any pane
		return m.runAllTests()

	case "A": // Deselect all tests or run all tests
		if m.focusedPane == TestsPane {
			return m.deselectAllTests()
		}
		// Run all tests from any pane
		return m.runAllTests()
	}

	// Pane-specific Vim keybindings
	switch m.focusedPane {
	case PackagesPane, TestsPane:
		return m.handleListVimKeys(msg)
	case DetailsPane:
		return m.handleDetailsVimKeys(msg)
	}

	return nil
}

// handleListVimKeys handles Vim keys for list panes
func (m *Model) handleListVimKeys(msg tea.KeyMsg) tea.Cmd {
	var targetList *list.Model
	if m.focusedPane == PackagesPane {
		targetList = &m.packageList
	} else {
		targetList = &m.testList
	}

	switch msg.String() {
	case "j", "down":
		targetList.CursorDown()
	case "k", "up":
		targetList.CursorUp()
	case "g":
		// Wait for second 'g' for gg command
		if m.lastKey == "g" {
			// gg - go to top
			for targetList.Cursor() > 0 {
				targetList.CursorUp()
			}
			m.lastKey = ""
		} else {
			m.lastKey = "g"
		}
	case "G":
		// Go to bottom
		items := targetList.Items()
		for targetList.Cursor() < len(items)-1 {
			targetList.CursorDown()
		}
	case "ctrl+d":
		// Half page down
		for i := 0; i < 10; i++ {
			targetList.CursorDown()
		}
	case "ctrl+u":
		// Half page up
		for i := 0; i < 10; i++ {
			targetList.CursorUp()
		}
	default:
		// Reset lastKey if it's not a continuation
		if m.lastKey == "g" && msg.String() != "g" {
			m.lastKey = ""
		}
	}

	return nil
}

// handleDetailsVimKeys handles Vim keys for details pane
func (m *Model) handleDetailsVimKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "j", "down":
		if m.detailsScrollPos < m.detailsMaxScroll {
			m.detailsScrollPos++
		}
	case "k", "up":
		if m.detailsScrollPos > 0 {
			m.detailsScrollPos--
		}
	case "G":
		m.detailsScrollPos = m.detailsMaxScroll
	case "g":
		if m.lastKey == "g" {
			// gg - go to top
			m.detailsScrollPos = 0
			m.lastKey = ""
		} else {
			m.lastKey = "g"
		}
	case "ctrl+d":
		m.detailsScrollPos += 10
		if m.detailsScrollPos > m.detailsMaxScroll {
			m.detailsScrollPos = m.detailsMaxScroll
		}
	case "ctrl+u":
		m.detailsScrollPos -= 10
		if m.detailsScrollPos < 0 {
			m.detailsScrollPos = 0
		}
	default:
		if m.lastKey == "g" && msg.String() != "g" {
			m.lastKey = ""
		}
	}

	return nil
}

// KeyMap defines keybindings
type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	NextPane key.Binding
	PrevPane key.Binding
	Enter    key.Binding
	Quit     key.Binding
	Help     key.Binding
}

// DefaultKeyMap returns the default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		NextPane: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next pane"),
		),
		PrevPane: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev pane"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
	}
}
