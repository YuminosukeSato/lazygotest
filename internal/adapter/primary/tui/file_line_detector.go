package tui

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Pattern to match file:line format (e.g., main.go:42 or /path/to/file.go:123)
	fileLinePattern = regexp.MustCompile(`([a-zA-Z0-9_./\-]+\.go):(\d+)`)
	
	// Style for highlighting file:line references
	fileLinkStyle = lipgloss.NewStyle().
		Foreground(accentColor).
		Underline(true).
		Bold(true)
)

// HighlightFileLines applies highlighting to file:line patterns in text
func HighlightFileLines(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = fileLinePattern.ReplaceAllStringFunc(line, func(match string) string {
			return fileLinkStyle.Render(match)
		})
	}
	return strings.Join(lines, "\n")
}

// ExtractFileLineFromText extracts the first file:line pattern from text
func ExtractFileLineFromText(text string) (string, int, bool) {
	matches := fileLinePattern.FindStringSubmatch(text)
	if len(matches) < 3 {
		return "", 0, false
	}
	
	file := matches[1]
	line, err := strconv.Atoi(matches[2])
	if err != nil {
		return "", 0, false
	}
	
	return file, line, true
}

// ExtractFileLineAtPosition extracts file:line at the given position in the details
func (m *Model) ExtractFileLineAtPosition() (string, int, bool) {
	if m.focusedPane != DetailsPane {
		return "", 0, false
	}
	
	// Get the current line from details
	if m.detailsScrollPos >= 0 && m.detailsScrollPos < len(m.detailsContent) {
		currentLine := m.detailsContent[m.detailsScrollPos]
		return ExtractFileLineFromText(currentLine)
	}
	
	return "", 0, false
}

// OpenInEditor opens the specified file at the given line in $EDITOR
func OpenInEditor(file string, line int) tea.Cmd {
	return func() tea.Msg {
		// Get editor from environment, default to vim
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim"
		}
		
		// Different editors have different syntax for opening at a specific line
		var cmd *exec.Cmd
		switch {
		case strings.Contains(editor, "vim") || strings.Contains(editor, "nvim"):
			// vim +42 file.go
			cmd = exec.Command(editor, fmt.Sprintf("+%d", line), file)
		case strings.Contains(editor, "emacs"):
			// emacs +42 file.go
			cmd = exec.Command(editor, fmt.Sprintf("+%d", line), file)
		case strings.Contains(editor, "nano"):
			// nano +42 file.go
			cmd = exec.Command(editor, fmt.Sprintf("+%d", line), file)
		case strings.Contains(editor, "code") || strings.Contains(editor, "vscode"):
			// code --goto file.go:42
			cmd = exec.Command(editor, "--goto", fmt.Sprintf("%s:%d", file, line))
		case strings.Contains(editor, "subl"):
			// subl file.go:42
			cmd = exec.Command(editor, fmt.Sprintf("%s:%d", file, line))
		case strings.Contains(editor, "atom"):
			// atom file.go:42
			cmd = exec.Command(editor, fmt.Sprintf("%s:%d", file, line))
		default:
			// Generic: just open the file
			cmd = exec.Command(editor, file)
		}
		
		// Use tea.ExecProcess to run the editor
		return tea.ExecProcess(cmd, func(err error) tea.Msg {
			if err != nil {
				return errorMsg{err: err}
			}
			return nil
		})
	}
}

// handleEditorJump handles the 'o' key to open file in editor
func (m *Model) handleEditorJump() tea.Cmd {
	file, line, found := m.ExtractFileLineAtPosition()
	if !found {
		// No file:line pattern found at current position
		return nil
	}
	
	return OpenInEditor(file, line)
}