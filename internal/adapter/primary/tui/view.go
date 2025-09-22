package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/YuminosukeSato/lazygotest/internal/domain"
)

var (
	// Colors - Okabe & Ito colorblind friendly palette for clear contrast
	primaryColor = lipgloss.Color("#0072B2") // Deep blue for main branding
	accentColor  = lipgloss.Color("#E69F00") // Warm orange accent
	successColor = lipgloss.Color("#009E73") // Teal for success state
	failureColor = lipgloss.Color("#D55E00") // Vermillion for failures
	warningColor = lipgloss.Color("#F0E442") // Vivid yellow for warnings
	mutedColor   = lipgloss.Color("#C7C7CC") // Soft neutral for secondary text

	// Neutral surfaces and emphasis tones
	surfaceColor         = lipgloss.Color("#1C1C1F")
	surfaceEmphasisColor = lipgloss.Color("#2B2B31")
	surfaceFocusFill     = lipgloss.Color("#353540")
	borderColor          = lipgloss.Color("#4A4A55")
	textPrimaryColor     = lipgloss.Color("#F7F7F7")
	textInverseColor     = lipgloss.Color("#1C1C1F")

	// State backgrounds for badges and highlights
	successBadgeBg = lipgloss.Color("#00664F")
	failureBadgeBg = lipgloss.Color("#7A2F00")
	warningBadgeBg = lipgloss.Color("#7A6B00")

	// Styles
	rootStyle = lipgloss.NewStyle().
			Background(surfaceColor).
			Foreground(textPrimaryColor)

	focusedBg = surfaceEmphasisColor

	titleStyle = lipgloss.NewStyle().
			Foreground(textPrimaryColor).
			Bold(true).
			Padding(0, 1)

	paneStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Background(surfaceColor).
			Foreground(textPrimaryColor).
			BorderTop(true).
			BorderBottom(true).
			BorderLeft(true).
			BorderRight(true)

	focusedPaneStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.DoubleBorder()).
				BorderForeground(accentColor).
				Background(focusedBg).
				Foreground(textPrimaryColor).
				Bold(true).
				BorderTop(true).
				BorderBottom(true).
				BorderLeft(true).
				BorderRight(true)

	paneTitleStyle = lipgloss.NewStyle().
			Background(surfaceEmphasisColor).
			Foreground(textPrimaryColor).
			Bold(true).
			Padding(0, 1)

	paneTitleFocusedStyle = paneTitleStyle.Copy().
				Background(accentColor).
				Foreground(textInverseColor)

	paneBodyStyle = lipgloss.NewStyle().
			Background(surfaceEmphasisColor).
			Foreground(textPrimaryColor).
			Padding(0, 1)

	paneBodyFocusedStyle = paneBodyStyle.Copy().
				Background(surfaceFocusFill)

	paneMetaStyle = lipgloss.NewStyle().
			Background(surfaceColor).
			Foreground(mutedColor).
			Padding(0, 1).
			Italic(true).
			Align(lipgloss.Right)

	paneMetaFocusedStyle = paneMetaStyle.Copy().
				Background(surfaceFocusFill).
				Foreground(textPrimaryColor)

	headerStyle = lipgloss.NewStyle().
			Background(primaryColor).
			Foreground(textPrimaryColor).
			Padding(0, 1).
			Bold(true)

	footerStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Background(surfaceColor).
			Padding(0, 1)

	statusPassStyle = lipgloss.NewStyle().
			Background(successBadgeBg).
			Foreground(textPrimaryColor).
			Bold(true).
			Padding(0, 1)

	statusFailStyle = lipgloss.NewStyle().
			Background(failureBadgeBg).
			Foreground(textPrimaryColor).
			Bold(true).
			Padding(0, 1)

	statusRunningStyle = lipgloss.NewStyle().
				Background(warningBadgeBg).
				Foreground(textPrimaryColor).
				Bold(true).
				Padding(0, 1)
)

// render renders the complete UI
func (m *Model) render() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	header := m.renderHeader()
	footer := m.renderFooter()
	content := m.renderContent()

	layout := lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		content,
		footer,
	)

	return rootStyle.Render(layout)
}

// renderHeader renders the header bar
func (m *Model) renderHeader() string {
	title := "lazygotest v0.1"

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

	if m.parallelMode {
		flags = append(flags, "[parallel:ON]")
	} else {
		flags = append(flags, "[parallel:OFF]")
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
			if isSelected && m.selectedPackage != nil && domain.PkgID(testID.Pkg) == m.selectedPackage.ID {
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
		"P:Parallel",
		"p:Run∥",
	}

	allKeys := append(commonKeys, paneKeys...)
	allKeys = append(allKeys, actionKeys...)

	return footerStyle.Width(m.width).Render(strings.Join(allKeys, " | "))
}

// renderContent renders the main content area with three panes
func (m *Model) renderContent() string {
	// Calculate pane widths with 25:40:35 ratio for better readability
	// Packages (25%), Tests (40%), Logs (35%)
	packagesWidth := int(float64(m.width) * 0.25)
	testsWidth := int(float64(m.width) * 0.40)
	detailsWidth := m.width - packagesWidth - testsWidth

	paneHeight := m.height - 4 // Account for header and footer

	// Render each pane
	packagesPane := m.renderPackagesPane(packagesWidth, paneHeight)
	testsPane := m.renderTestsPane(testsWidth, paneHeight)
	detailsPane := m.renderDetailsPane(detailsWidth, paneHeight)

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

	contentWidth := width - 2
	if contentWidth < 0 {
		contentWidth = 0
	}
	contentHeight := height - 4 // Leave space for borders, title, and footer line
	if contentHeight < 0 {
		contentHeight = 0
	}
	innerWidth := contentWidth - 2 // Account for pane body horizontal padding
	if innerWidth < 0 {
		innerWidth = 0
	}

	// Update list size with inner content dimensions
	m.packageList.SetSize(innerWidth, contentHeight)

	// Build title with focus indicator
	title := "Packages"
	if isFocused {
		title = "▶ " + title
	}
	titleBarStyle := paneTitleStyle
	if isFocused {
		titleBarStyle = paneTitleFocusedStyle
	}
	titleContent := titleBarStyle.Width(contentWidth).Render(title)

	// Position indicator
	items := m.packageList.Items()
	cursor := m.packageList.Cursor()
	total := len(items)
	positionText := "[0/0]"
	if total > 0 {
		positionText = "[" + intToString(cursor+1) + "/" + intToString(total) + "]"
	}
	metaStyle := paneMetaStyle
	if isFocused {
		metaStyle = paneMetaFocusedStyle
	}
	positionLine := metaStyle.Width(contentWidth).Render(positionText)

	// Body content with consistent background
	bodyStyle := paneBodyStyle
	if isFocused {
		bodyStyle = paneBodyFocusedStyle
	}
	content := bodyStyle.Width(contentWidth).Height(contentHeight).Render(m.packageList.View())

	// Combine all elements
	fullContent := lipgloss.JoinVertical(
		lipgloss.Top,
		titleContent,
		content,
		positionLine,
	)

	return style.Width(width).Height(height).Render(fullContent)
}

// renderTestsPane renders the tests list or tree view
func (m *Model) renderTestsPane(width, height int) string {
	style := paneStyle
	isFocused := m.focusedPane == TestsPane
	if isFocused {
		style = focusedPaneStyle
	}

	contentWidth := width - 2
	if contentWidth < 0 {
		contentWidth = 0
	}
	contentHeight := height - 4
	if contentHeight < 0 {
		contentHeight = 0
	}
	innerWidth := contentWidth - 2
	if innerWidth < 0 {
		innerWidth = 0
	}

	// Build title with focus indicator
	title := "Tests"
	if m.selectedPackage != nil {
		title = "Tests in " + m.selectedPackage.Name
	}
	if m.treeViewMode {
		title += " [Tree]"
	} else {
		title += " [List]"
	}
	if isFocused {
		title = "▶ " + title
	}
	titleBarStyle := paneTitleStyle
	if isFocused {
		titleBarStyle = paneTitleFocusedStyle
	}
	titleContent := titleBarStyle.Width(contentWidth).Render(title)

	var bodyContent string
	var positionText string
	bodyStyle := paneBodyStyle
	if isFocused {
		bodyStyle = paneBodyFocusedStyle
	}

	if m.treeViewMode && m.selectedPackage != nil {
		if m.treeRenderer != nil {
			m.treeRenderer.SetWidth(innerWidth)
		}

		pkgID := string(m.selectedPackage.ID)
		testTree, exists := m.testTrees[pkgID]
		if !exists || testTree == nil {
			testTree = &domain.TestTree{
				Package: m.selectedPackage.Name,
				Tests:   make(map[string]*domain.TestNode),
			}
			m.testTrees[pkgID] = testTree
		}

		if m.treeRenderer != nil {
			bodyContent = m.treeRenderer.RenderTree(testTree)
		} else {
			bodyContent = "Tree renderer not initialized"
		}

		nodeCount := len(testTree.Tests)
		positionText = "[" + intToString(nodeCount) + " tests]"
	} else {
		m.testList.SetSize(innerWidth, contentHeight)

		items := m.testList.Items()
		cursor := m.testList.Cursor()
		total := len(items)
		if total > 0 {
			positionText = "[" + intToString(cursor+1) + "/" + intToString(total) + "]"
		} else {
			positionText = "[0/0]"
		}

		bodyContent = m.testList.View()
	}

	contentBlock := bodyStyle.Width(contentWidth).Height(contentHeight).Render(bodyContent)

	metaStyle := paneMetaStyle
	if isFocused {
		metaStyle = paneMetaFocusedStyle
	}
	positionLine := metaStyle.Width(contentWidth).Render(positionText)

	fullContent := lipgloss.JoinVertical(
		lipgloss.Top,
		titleContent,
		contentBlock,
		positionLine,
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

	contentWidth := width - 2
	if contentWidth < 0 {
		contentWidth = 0
	}
	contentHeight := height - 4
	if contentHeight < 0 {
		contentHeight = 0
	}

	// Calculate visible content area (accounting for title and position)
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

	// Apply file:line highlighting to visible lines
	for i, line := range visibleLines {
		visibleLines[i] = HighlightFileLines(line)
	}

	content := strings.Join(visibleLines, "\n")

	// Build title with focus indicator and test status
	titleText := "Details / Logs"
	statusBadge := ""
	if m.selectedTest != nil {
		titleText = "Logs: " + m.selectedTest.ID.Name
		switch m.selectedTest.Status {
		case domain.StatusFailed:
			statusBadge = statusFailStyle.Render("✗ ")
		case domain.StatusPassed:
			statusBadge = statusPassStyle.Render("✓ ")
		case domain.StatusRunning:
			statusBadge = statusRunningStyle.Render("⟳ ")
		}
	}
	if isFocused {
		titleText = "▶ " + titleText
	}
	titleBarStyle := paneTitleStyle
	if isFocused {
		titleBarStyle = paneTitleFocusedStyle
	}

	if statusBadge != "" {
		statusBadge += " "
	}

	titleContent := ""
	if statusBadge != "" {
		badgeWidth := lipgloss.Width(statusBadge)
		remainingWidth := contentWidth - badgeWidth
		if remainingWidth < 0 {
			remainingWidth = 0
		}
		titleContent = lipgloss.JoinHorizontal(
			lipgloss.Left,
			statusBadge,
			titleBarStyle.Width(remainingWidth).Render(titleText),
		)
	} else {
		titleContent = titleBarStyle.Width(contentWidth).Render(titleText)
	}

	// Position/scroll indicator
	positionText := "[Empty]"
	if len(m.detailsContent) > 0 {
		currentLine := m.detailsScrollPos + 1
		totalLines := len(m.detailsContent)
		scrollPercent := 0
		if m.detailsMaxScroll > 0 {
			scrollPercent = (m.detailsScrollPos * 100) / m.detailsMaxScroll
		}
		positionText = "[Line " + intToString(currentLine) + "/" + intToString(totalLines) + " (" + intToString(scrollPercent) + "%)]"
	}
	metaStyle := paneMetaStyle
	if isFocused {
		metaStyle = paneMetaFocusedStyle
	}
	positionLine := metaStyle.Width(contentWidth).Render(positionText)

	bodyStyle := paneBodyStyle
	if isFocused {
		bodyStyle = paneBodyFocusedStyle
	}
	bodyContent := bodyStyle.Width(contentWidth).Height(contentHeight).Render(content)

	// Combine all elements
	fullContent := lipgloss.JoinVertical(
		lipgloss.Top,
		titleContent,
		bodyContent,
		positionLine,
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
