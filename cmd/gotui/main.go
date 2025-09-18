package main

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"lazygotest/internal/adapter/primary/tui"
	"lazygotest/pkg/logger"
)

func main() {
	// Initialize debug logger
	if err := logger.Init("debug.log"); err != nil {
		logger.Error("Failed to initialize logger", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := logger.Close(); err != nil {
			logger.Error("Failed to close logger", "error", err)
		}
	}()

	logger.Debug("Starting gotui application")

	// Create and run the TUI application
	app := tui.New()
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		logger.Error("Error running program", "error", err)
		os.Exit(1)
	}
}
