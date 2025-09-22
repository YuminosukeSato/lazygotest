package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/YuminosukeSato/lazygotest/internal/domain"
	"github.com/charmbracelet/lipgloss"
)

// TreeRenderer handles the rendering of test trees
type TreeRenderer struct {
	width int
	// Styles for different test states
	passStyle    lipgloss.Style
	failStyle    lipgloss.Style
	skipStyle    lipgloss.Style
	runningStyle lipgloss.Style
	pendingStyle lipgloss.Style
	timeStyle    lipgloss.Style
}

// NewTreeRenderer creates a new tree renderer
func NewTreeRenderer(width int) *TreeRenderer {
	return &TreeRenderer{
		width:        width,
		passStyle:    lipgloss.NewStyle().Foreground(successColor).Bold(true),
		failStyle:    lipgloss.NewStyle().Foreground(failureColor).Bold(true),
		skipStyle:    lipgloss.NewStyle().Foreground(warningColor),
		runningStyle: lipgloss.NewStyle().Foreground(accentColor).Bold(true),
		pendingStyle: lipgloss.NewStyle().Foreground(mutedColor),
		timeStyle:    lipgloss.NewStyle().Foreground(mutedColor).Italic(true),
	}
}

// RenderTree renders the entire test tree
func (r *TreeRenderer) RenderTree(tree *domain.TestTree) string {
	if tree == nil || len(tree.Tests) == 0 {
		return lipgloss.NewStyle().Foreground(mutedColor).Render("No tests available")
	}

	var lines []string

	// Render package-level information if available
	if tree.Package != "" {
		pkgLine := r.renderPackageHeader(tree)
		lines = append(lines, pkgLine)
	}

	// Render each top-level test
	for _, test := range tree.Tests {
		testLines := r.renderNode(test, 0, false)
		lines = append(lines, testLines...)
	}

	return strings.Join(lines, "\n")
}

// renderPackageHeader renders the package header with overall status
func (r *TreeRenderer) renderPackageHeader(tree *domain.TestTree) string {
	icon := r.getStatusIcon(tree.Status)
	duration := r.formatDuration(tree.Duration)

	header := fmt.Sprintf("%s Package: %s", icon, tree.Package)
	if duration != "" {
		// Right-align the duration
		spacesNeeded := r.width - lipgloss.Width(header) - lipgloss.Width(duration) - 2
		if spacesNeeded > 0 {
			header += strings.Repeat(" ", spacesNeeded) + duration
		}
	}

	return lipgloss.NewStyle().Bold(true).Render(header)
}

// renderNode recursively renders a test node and its children
func (r *TreeRenderer) renderNode(node *domain.TestNode, depth int, isLast bool) []string {
	if node == nil {
		return nil
	}

	var lines []string

	// Build the tree prefix
	prefix := r.buildTreePrefix(depth, isLast)

	// Get status icon and style
	icon := r.getStatusIcon(node.Status)

	// Format the test name with icon
	testName := fmt.Sprintf("%s%s %s", prefix, icon, node.Name)

	// Format duration if available
	duration := r.formatDuration(node.Duration)

	// Build the complete line with right-aligned duration
	line := testName
	if duration != "" {
		spacesNeeded := r.width - lipgloss.Width(testName) - lipgloss.Width(duration) - 2
		if spacesNeeded > 0 {
			line += strings.Repeat(" ", spacesNeeded) + r.timeStyle.Render(duration)
		}
	}

	// Apply status styling
	line = r.applyStatusStyle(line, node.Status)
	lines = append(lines, line)

	// Render subtests if they exist
	if len(node.SubTests) > 0 {
		subTestCount := len(node.SubTests)
		i := 0
		for _, subTest := range node.SubTests {
			isLastSubTest := i == subTestCount-1
			subLines := r.renderNode(subTest, depth+1, isLastSubTest)
			lines = append(lines, subLines...)
			i++
		}
	}

	return lines
}

// buildTreePrefix builds the tree structure prefix for indentation
func (r *TreeRenderer) buildTreePrefix(depth int, isLast bool) string {
	if depth == 0 {
		return ""
	}

	prefix := strings.Repeat("  ", depth-1)
	if isLast {
		prefix += "└─ "
	} else {
		prefix += "├─ "
	}

	return prefix
}

// getStatusIcon returns the appropriate icon for a test status
func (r *TreeRenderer) getStatusIcon(status domain.TestStatus) string {
	switch status {
	case domain.StatusPassed:
		return r.passStyle.Render("✓")
	case domain.StatusFailed:
		return r.failStyle.Render("✗")
	case domain.StatusSkipped:
		return r.skipStyle.Render("–")
	case domain.StatusRunning:
		return r.runningStyle.Render("⟳")
	default:
		return r.pendingStyle.Render("○")
	}
}

// applyStatusStyle applies the appropriate style based on test status
func (r *TreeRenderer) applyStatusStyle(text string, status domain.TestStatus) string {
	switch status {
	case domain.StatusFailed:
		// Highlight failed tests with background color
		return lipgloss.NewStyle().
			Background(failureBadgeBg).
			Foreground(textPrimaryColor).
			Width(r.width).
			Render(text)
	case domain.StatusRunning:
		// Highlight running tests
		return lipgloss.NewStyle().
			Background(warningBadgeBg).
			Foreground(textPrimaryColor).
			Width(r.width).
			Render(text)
	default:
		return text
	}
}

// formatDuration formats a duration for display
func (r *TreeRenderer) formatDuration(d time.Duration) string {
	if d == 0 {
		return ""
	}

	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%.0fms", d.Seconds()*1000)
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

// SetWidth updates the renderer width
func (r *TreeRenderer) SetWidth(width int) {
	r.width = width
}
