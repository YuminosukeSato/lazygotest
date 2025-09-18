package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"lazygotest/internal/domain"
)

var (
	// Colors - Colorblind friendly palette
	// Using blue/orange as primary contrast (safe for all types of color blindness)
	primaryColor = lipgloss.Color("#0066CC") // Blue (safe)
	accentColor  = lipgloss.Color("#FF9500") // Orange (distinguishable from blue)
	successColor = lipgloss.Color("#007AFF") // Bright blue for success
	failureColor = lipgloss.Color("#FF3B30") // Red with icon for fail
	warningColor = lipgloss.Color("#FF9500") // Orange for warnings
	mutedColor   = lipgloss.Color("#8E8E93") // Gray for muted text

	// Background colors with different brightness levels
	focusedBg = lipgloss.Color("#2C2C2E") // Medium gray for focused

	// Styles
	titleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Padding(0, 1)

	paneStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(mutedColor).
			BorderTop(true).
			BorderBottom(true).
			BorderLeft(true).
			BorderRight(true)

	focusedPaneStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.DoubleBorder()).
				BorderForeground(accentColor).
				Bold(true).
				Background(focusedBg).
				BorderTop(true).
				BorderBottom(true).
				BorderLeft(true).
				BorderRight(true)

	headerStyle = lipgloss.NewStyle().
			Background(primaryColor).
			Foreground(lipgloss.Color("#FAFAFA")).
			Padding(0, 1).
			Bold(true)

	footerStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Padding(0, 1)

	statusPassStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	statusFailStyle = lipgloss.NewStyle().
			Foreground(failureColor).
			Bold(true)

	statusRunningStyle = lipgloss.NewStyle().
				Foreground(warningColor).
				Bold(true)

	positionIndicatorStyle = lipgloss.NewStyle().
				Foreground(mutedColor).
				Italic(true)
)

// render renders the complete UI
func (m *Model) render() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	header := m.renderHeader()
	footer := m.renderFooter()
	content := m.renderContent()

	return lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		content,
		footer,
	)
}

// renderHeader renders the header bar
func (m *Model) renderHeader() string {
	title := "gotui v0.1"

	flags := []string{}
	if m.raceDetection {
		flags = append(flags, "[race:ON]")
	} else {
		flags = append(flags, "[race:OFF]")
	}

	if m.coverageEnabled {
		flags = append(flags, "[cover:ON]")
	} else {
		flags = append(flags, "[cover:OFF]")
	}

	if m.watchMode {
		flags = append(flags, "[watch:ON]")
	} else {
		flags = append(flags, "[watch:OFF]")
	}

	if m.showFailedOnly {
		flags = append(flags, "[failed-only]")
	}

	flagsStr := strings.Join(flags, " ")

	// Status summary
	status := ""
	if m.summary != nil {
		pass := statusPassStyle.Render("PASS " + intToString(m.summary.Passed))
		fail := statusFailStyle.Render("FAIL " + intToString(m.summary.Failed))
		skip := lipgloss.NewStyle().Foreground(mutedColor).Render("SKIP " + intToString(m.summary.Skipped))
		status = pass + " | " + fail + " | " + skip
	} else if m.isRunning {
		status = statusRunningStyle.Render("⟳ Running...")
	}

	headerContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		titleStyle.Render(title),
		lipgloss.NewStyle().Width(10).Render(" "),
		flagsStr,
		lipgloss.NewStyle().Width(m.width-lipgloss.Width(title)-lipgloss.Width(flagsStr)-lipgloss.Width(status)-15).Render(" "),
		status,
	)

	return headerStyle.Width(m.width).Render(headerContent)
}

// renderFooter renders the footer with keybindings
func (m *Model) renderFooter() string {
	// Show different keys based on focused pane
	commonKeys := []string{
		"q:Quit",
		"h/l:←/→ Pane",
	}

	paneKeys := []string{}
	switch m.focusedPane {
	case PackagesPane:
		paneKeys = []string{
			"j/k:↓/↑",
			"gg/G:Top/Bot",
			"Enter:Run",
			"/:Search",
		}
	case TestsPane:
		paneKeys = []string{
			"j/k:↓/↑",
			"Space:Toggle",
			"a/A:All/None",
			"Enter:Run",
		}
		// Add selection count if tests are selected
		selectedCount := 0
		for testID, isSelected := range m.selectedTests {
			if isSelected && m.selectedPackage != nil && testID.Pkg == m.selectedPackage.ID {
				selectedCount++
			}
		}
		if selectedCount > 0 {
			paneKeys = append(paneKeys, "["+intToString(selectedCount)+" selected]")
		}
	case DetailsPane:
		paneKeys = []string{
			"j/k:Scroll ↓/↑",
			"gg/G:Top/Bot",
			"^d/^u:PageDn/Up",
		}
	}

	actionKeys := []string{
		"A:All",
		"F:Failed",
		"W:Watch",
		"R:Race",
		"C:Cover",
	}

	allKeys := append(commonKeys, paneKeys...)
	allKeys = append(allKeys, actionKeys...)

	return footerStyle.Width(m.width).Render(strings.Join(allKeys, " | "))
}

// renderContent renders the main content area with three panes
func (m *Model) renderContent() string {
	paneWidth := m.width / 3
	paneHeight := m.height - 4 // Account for header and footer

	// Render each pane
	packagesPane := m.renderPackagesPane(paneWidth, paneHeight)
	testsPane := m.renderTestsPane(paneWidth, paneHeight)
	detailsPane := m.renderDetailsPane(paneWidth, paneHeight)

	// Join panes horizontally
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		packagesPane,
		testsPane,
		detailsPane,
	)
}

// renderPackagesPane renders the packages list
func (m *Model) renderPackagesPane(width, height int) string {
	style := paneStyle
	isFocused := m.focusedPane == PackagesPane
	if isFocused {
		style = focusedPaneStyle
	}

	// Update list size
	m.packageList.SetSize(width-2, height-4) // Leave space for title and position

	// Build title with focus indicator
	title := "Packages"
	if isFocused {
		title = "▶ " + title
	}
	titleContent := titleStyle.Render(title)

	// Position indicator
	items := m.packageList.Items()
	cursor := m.packageList.Cursor()
	total := len(items)
	position := ""
	if total > 0 {
		position = positionIndicatorStyle.Render("[" + intToString(cursor+1) + "/" + intToString(total) + "]")
	} else {
		position = positionIndicatorStyle.Render("[0/0]")
	}

	content := m.packageList.View()

	// Combine all elements
	fullContent := lipgloss.JoinVertical(
		lipgloss.Top,
		titleContent,
		content,
		position,
	)

	return style.Width(width).Height(height).Render(fullContent)
}

// renderTestsPane renders the tests list
func (m *Model) renderTestsPane(width, height int) string {
	style := paneStyle
	isFocused := m.focusedPane == TestsPane
	if isFocused {
		style = focusedPaneStyle
	}

	// Update list size
	m.testList.SetSize(width-2, height-4) // Leave space for title and position

	// Build title with focus indicator
	title := "Tests"
	if m.selectedPackage != nil {
		title = "Tests in " + m.selectedPackage.Name
	}
	if isFocused {
		title = "▶ " + title
	}
	titleContent := titleStyle.Render(title)

	// Position indicator
	items := m.testList.Items()
	cursor := m.testList.Cursor()
	total := len(items)
	position := ""
	if total > 0 {
		position = positionIndicatorStyle.Render("[" + intToString(cursor+1) + "/" + intToString(total) + "]")
	} else {
		position = positionIndicatorStyle.Render("[0/0]")
	}

	content := m.testList.View()

	// Combine all elements
	fullContent := lipgloss.JoinVertical(
		lipgloss.Top,
		titleContent,
		content,
		position,
	)

	return style.Width(width).Height(height).Render(fullContent)
}

// renderDetailsPane renders the details/logs pane
func (m *Model) renderDetailsPane(width, height int) string {
	style := paneStyle
	isFocused := m.focusedPane == DetailsPane
	if isFocused {
		style = focusedPaneStyle
	}

	// Calculate visible content area (accounting for title and position)
	contentHeight := height - 4

	// Update max scroll
	m.detailsMaxScroll = len(m.detailsContent) - contentHeight
	if m.detailsMaxScroll < 0 {
		m.detailsMaxScroll = 0
	}

	// Build content from details with scroll position
	startIdx := m.detailsScrollPos
	endIdx := startIdx + contentHeight
	if endIdx > len(m.detailsContent) {
		endIdx = len(m.detailsContent)
	}
	if startIdx >= len(m.detailsContent) && len(m.detailsContent) > 0 {
		startIdx = len(m.detailsContent) - contentHeight
		if startIdx < 0 {
			startIdx = 0
		}
	}

	visibleLines := []string{}
	if startIdx < len(m.detailsContent) {
		visibleLines = m.detailsContent[startIdx:endIdx]
	}
	content := strings.Join(visibleLines, "\n")

	// Build title with focus indicator and test status
	title := "Details / Logs"
	if m.selectedTest != nil {
		title = "Logs: " + m.selectedTest.ID.Name
		switch m.selectedTest.Status {
		case domain.StatusFailed:
			title = statusFailStyle.Render("✗ ") + title
		case domain.StatusPassed:
			title = statusPassStyle.Render("✓ ") + title
		case domain.StatusRunning:
			title = statusRunningStyle.Render("⟳ ") + title
		}
	}
	if isFocused {
		title = "▶ " + title
	}
	titleContent := titleStyle.Render(title)

	// Position/scroll indicator
	position := ""
	if len(m.detailsContent) > 0 {
		currentLine := m.detailsScrollPos + 1
		totalLines := len(m.detailsContent)
		scrollPercent := 0
		if m.detailsMaxScroll > 0 {
			scrollPercent = (m.detailsScrollPos * 100) / m.detailsMaxScroll
		}
		position = positionIndicatorStyle.Render(
			"[Line " + intToString(currentLine) + "/" + intToString(totalLines) +
				" (" + intToString(scrollPercent) + "%)]",
		)
	} else {
		position = positionIndicatorStyle.Render("[Empty]")
	}

	// Combine all elements
	fullContent := lipgloss.JoinVertical(
		lipgloss.Top,
		titleContent,
		content,
		position,
	)

	return style.Width(width).Height(height).Render(fullContent)
}

// Helper to convert int to string without fmt
func intToString(n int) string {
	if n == 0 {
		return "0"
	}

	var result []byte
	negative := n < 0
	if negative {
		n = -n
	}

	for n > 0 {
		digit := n % 10
		result = append([]byte{byte('0' + digit)}, result...)
		n /= 10
	}

	if negative {
		result = append([]byte{'-'}, result...)
	}

	return string(result)
}
