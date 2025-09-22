package tui

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/YuminosukeSato/lazygotest/internal/domain"
	"github.com/YuminosukeSato/lazygotest/internal/shared/eventbus"
	"github.com/YuminosukeSato/lazygotest/pkg/logger"
)

// formatInt converts int to string
func formatInt(n int) string {
	return strconv.Itoa(n)
}

// Message types
type packagesLoadedMsg struct {
	packages []*domain.Package
}

type testEventMsg struct {
	event domain.TestEvent
}

type errorMsg struct {
	err error
}

// loadPackages loads the list of packages
func (m *Model) loadPackages() tea.Cmd {
	return func() tea.Msg {
		packages, err := m.listPkgsUC.Execute(m.ctx)
		if err != nil {
			return errorMsg{err: err}
		}
		return packagesLoadedMsg{packages: packages}
	}
}

// runAllTests runs all tests
func (m *Model) runAllTests() tea.Cmd {
	if m.isRunning {
		return nil
	}

	m.isRunning = true
	m.detailsBuffer.Clear()
	m.detailsBuffer.Add("Running all tests...")
	m.detailsContent = m.detailsBuffer.GetLines()

	return func() tea.Msg {
		err := m.runTestsUC.ExecuteAll(m.ctx)
		if err != nil {
			return errorMsg{err: err}
		}
		return nil
	}
}

// runAllPackagesParallel runs all packages in parallel
func (m *Model) runAllPackagesParallel() tea.Cmd {
	if m.isRunning || len(m.packages) == 0 {
		return nil
	}

	m.isRunning = true
	m.detailsBuffer.Clear()
	m.detailsBuffer.Add("Running all packages in parallel...")
	m.detailsBuffer.Add("Max workers: " + formatInt(m.parallelRunner.GetMaxWorkers()))
	m.detailsContent = m.detailsBuffer.GetLines()

	return func() tea.Msg {
		err := m.parallelRunner.RunPackagesParallel(m.ctx, m.packages, m.raceDetection, m.coverageEnabled)
		if err != nil {
			m.isRunning = false
			return errorMsg{err: err}
		}

		// Get aggregated summary
		m.summary = m.parallelRunner.GetTotalSummary()
		m.isRunning = false

		// Log completion
		m.appendDetail("Parallel execution completed")
		m.appendDetail("Total: " + formatInt(m.summary.Total) + 
			" | Passed: " + formatInt(m.summary.Passed) +
			" | Failed: " + formatInt(m.summary.Failed) +
			" | Skipped: " + formatInt(m.summary.Skipped))

		return nil
	}
}

// runSelectedPackage runs tests for the selected package
func (m *Model) runSelectedPackage() tea.Cmd {
	if m.selectedPackage == nil || m.isRunning {
		return nil
	}

	m.isRunning = true
	m.detailsBuffer.Clear()
	m.detailsBuffer.Add("Running tests in " + m.selectedPackage.Name + "...")
	m.detailsContent = m.detailsBuffer.GetLines()

	return func() tea.Msg {
		err := m.runTestsUC.ExecutePackage(m.ctx, m.selectedPackage.ID)
		if err != nil {
			return errorMsg{err: err}
		}
		return nil
	}
}

// rerunTest reruns the selected test
func (m *Model) rerunTest() tea.Cmd {
	if m.selectedTest == nil || m.isRunning {
		return nil
	}

	m.isRunning = true
	m.detailsBuffer.Clear()
	m.detailsBuffer.Add("Running " + m.selectedTest.ID.Name + "...")
	m.detailsContent = m.detailsBuffer.GetLines()

	return func() tea.Msg {
		err := m.runTestsUC.ExecuteTest(m.ctx, m.selectedTest.ID)
		if err != nil {
			return errorMsg{err: err}
		}
		return nil
	}
}

// handleEnter handles the enter key based on focused pane
func (m *Model) handleEnter() tea.Cmd {
	switch m.focusedPane {
	case PackagesPane:
		if i, ok := m.packageList.SelectedItem().(packageItem); ok {
			m.selectedPackage = i.pkg
			m.updateTestsForPackage(i.pkg)
			return m.runSelectedPackage()
		}

	case TestsPane:
		// Run selected tests or current test
		return m.runSelectedTests()
	}

	return nil
}

// subscribeToEvents sets up event handlers
func (m *Model) subscribeToEvents() {
	// Subscribe to test events
	m.eventBus.Subscribe(eventbus.TopicTestEvent, func(ctx context.Context, event interface{}) {
		if e, ok := event.(domain.TestEvent); ok {
			m.handleTestEventAsync(e)
		}
	})

	// Subscribe to test completion
	m.eventBus.Subscribe(eventbus.TopicTestCompleted, func(ctx context.Context, event interface{}) {
		if summary, ok := event.(*domain.TestSummary); ok {
			m.summary = summary
			m.isRunning = false
			logger.Info("Tests completed", "summary", summary)
		}
	})

	// Subscribe to errors
	m.eventBus.Subscribe(eventbus.TopicError, func(ctx context.Context, event interface{}) {
		if err, ok := event.(error); ok {
			logger.Error("Event bus error", "error", err)
		}
	})
}

// handleTestEventAsync processes test events asynchronously
func (m *Model) handleTestEventAsync(event domain.TestEvent) {
	// This would normally send a message through the Bubble Tea program
	// For now, we'll update the model directly (in a real app, use tea.Cmd)
	m.handleTestEvent(event)
}

// handleTestEvent processes a test event
func (m *Model) handleTestEvent(event domain.TestEvent) {
	logger.Debug("Handling test event", "event", event)

	// Build/update the test tree
	m.BuildTestTree(event)

	// Update test results (for backward compatibility with list view)
	if event.Test != "" {
		testID := domain.TestID{
			Pkg:  event.Package,
			Name: event.Test,
		}

		test, exists := m.testResults[testID]
		if !exists {
			test = &domain.TestCase{
				ID:   testID,
				Logs: []string{},
			}
			m.testResults[testID] = test
		}

		switch event.Action {
		case "run":
			test.Status = domain.StatusRunning
		case "pass":
			test.Status = domain.StatusPassed
			if event.Elapsed > 0 {
				test.Duration = time.Duration(event.Elapsed * float64(time.Second))
			}
		case "fail":
			test.Status = domain.StatusFailed
			if test.LastFail == nil {
				test.LastFail = &domain.FailInfo{}
			}
			test.LastFail.FullLog = strings.Join(test.Logs, "\n")
			if event.Elapsed > 0 {
				test.Duration = time.Duration(event.Elapsed * float64(time.Second))
			}
		case "skip":
			test.Status = domain.StatusSkipped
		case "output":
			test.Logs = append(test.Logs, event.Output)
			m.appendDetail(event.Output)
		}
	} else if event.Output != "" {
		// Package-level output
		m.appendDetail(event.Output)
	}

	// Update UI
	m.updateTestList()
}

// appendDetail adds a line to the details pane
func (m *Model) appendDetail(line string) {
	// Add to ring buffer
	m.detailsBuffer.Add(line)
	// Update cached content for rendering
	m.detailsContent = m.detailsBuffer.GetLines()
}

// focusNextPane moves focus to the next pane
func (m *Model) focusNextPane() {
	m.focusedPane = (m.focusedPane + 1) % 3
}

// focusPrevPane moves focus to the previous pane
func (m *Model) focusPrevPane() {
	m.focusedPane = (m.focusedPane + 2) % 3
}

// updateSizes updates component sizes based on window size
func (m *Model) updateSizes() {
	paneWidth := m.width / 3
	paneHeight := m.height - 4 // Leave room for header and footer

	m.packageList.SetSize(paneWidth, paneHeight)
	m.testList.SetSize(paneWidth, paneHeight)
}

// List items for packages and tests
type packageItem struct {
	pkg *domain.Package
}

func (i packageItem) Title() string       { return i.pkg.Name }
func (i packageItem) Description() string { return string(i.pkg.ID) }
func (i packageItem) FilterValue() string { return i.pkg.Name }

type testItem struct {
	test       *domain.TestCase
	isSelected bool
}

func (i testItem) Title() string {
	// Selection indicator with better visibility
	checkbox := "[ ]"
	if i.isSelected {
		checkbox = "[■]" // Filled square for selected
	}

	// Status indicator with icons
	status := ""
	statusStyle := lipgloss.NewStyle()

	switch i.test.Status {
	case domain.StatusPassed:
		status = "✓"
		statusStyle = statusStyle.Background(successBadgeBg).
			Foreground(textPrimaryColor)
	case domain.StatusFailed:
		status = "✗"
		statusStyle = statusStyle.Background(failureBadgeBg).
			Foreground(textPrimaryColor)
	case domain.StatusRunning:
		status = "⟳"
		statusStyle = statusStyle.Background(warningBadgeBg).
			Foreground(textPrimaryColor)
	case domain.StatusSkipped:
		status = "-"
	default:
		status = " "
	}

	// Apply styling to the entire line
	fullText := checkbox + " " + status + " " + i.test.ID.Name

	// Apply background color if test has been run
	if i.test.Status == domain.StatusPassed || i.test.Status == domain.StatusFailed || i.test.Status == domain.StatusRunning {
		return statusStyle.Width(40).Render(fullText)
	}

	return fullText
}

func (i testItem) Description() string {
	if i.test.Duration > 0 {
		return strconv.FormatFloat(i.test.Duration.Seconds(), 'f', 2, 64) + "s"
	}
	return ""
}

func (i testItem) FilterValue() string { return i.test.ID.Name }

// updatePackageList updates the package list UI
func (m *Model) updatePackageList() {
	items := make([]list.Item, len(m.packages))
	for i, pkg := range m.packages {
		items[i] = packageItem{pkg: pkg}
	}
	m.packageList.SetItems(items)
}

// updateTestList updates the test list UI
func (m *Model) updateTestList() {
	if m.selectedPackage == nil {
		m.testList.SetItems([]list.Item{})
		return
	}

	items := make([]list.Item, 0)

	// Get tests for selected package
	for _, test := range m.testResults {
		if domain.PkgID(test.ID.Pkg) != m.selectedPackage.ID {
			continue
		}

		if m.showFailedOnly && test.Status != domain.StatusFailed {
			continue
		}

		// Check if test is selected
		isSelected := m.selectedTests[test.ID]
		items = append(items, testItem{
			test:       test,
			isSelected: isSelected,
		})
	}

	m.testList.SetItems(items)
}

// updateTestsForPackage loads tests for a package
func (m *Model) updateTestsForPackage(pkg *domain.Package) {
	// In a real implementation, this would discover tests
	// For now, we'll wait for test events to populate the list
	m.testList.Title = "Tests in " + pkg.Name

	// TODO: Call runner.ListTests to get available tests
	// m.availableTests = runner.ListTests(ctx, pkg.ID)
}

// toggleTestSelection toggles the selection state of the current test
func (m *Model) toggleTestSelection() tea.Cmd {
	if item, ok := m.testList.SelectedItem().(testItem); ok {
		testID := item.test.ID
		m.selectedTests[testID] = !m.selectedTests[testID]
		m.updateTestList()
	}
	return nil
}

// selectAllTests selects all visible tests
func (m *Model) selectAllTests() tea.Cmd {
	items := m.testList.Items()
	for _, item := range items {
		if ti, ok := item.(testItem); ok {
			m.selectedTests[ti.test.ID] = true
		}
	}
	m.updateTestList()
	return nil
}

// deselectAllTests deselects all tests
func (m *Model) deselectAllTests() tea.Cmd {
	// Clear all selections
	m.selectedTests = make(map[domain.TestID]bool)
	m.updateTestList()
	return nil
}

// runSelectedTests runs all selected tests or current test if none selected
func (m *Model) runSelectedTests() tea.Cmd {
	if m.isRunning || m.selectedPackage == nil {
		return nil
	}

	// Get selected test IDs
	var selectedIDs []domain.TestID
	for testID, isSelected := range m.selectedTests {
		if isSelected && domain.PkgID(testID.Pkg) == m.selectedPackage.ID {
			selectedIDs = append(selectedIDs, testID)
		}
	}

	// If no tests selected, run the current test
	if len(selectedIDs) == 0 {
		if item, ok := m.testList.SelectedItem().(testItem); ok {
			selectedIDs = []domain.TestID{item.test.ID}
		}
	}

	if len(selectedIDs) == 0 {
		return nil
	}

	m.isRunning = true
	m.detailsContent = []string{"Running selected tests..."}

	return func() tea.Msg {
		// Execute multiple tests
		err := m.runTestsUC.ExecuteMultipleTests(m.ctx, selectedIDs)
		if err != nil {
			return errorMsg{err: err}
		}
		return nil
	}
}
