package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"lazygotest/internal/adapter/primary/tui"
	"lazygotest/pkg/logger"
)

// Build-time variables injected via ldflags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Parse command line flags
	versionFlag := flag.Bool("version", false, "Print version information")
	flag.Parse()

	// Handle --version flag
	if *versionFlag {
		fmt.Printf("lazygotest version %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built:  %s\n", date)
		os.Exit(0)
	}

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

	logger.Debug("Starting lazygotest application")

	// Create and run the TUI application
	app := tui.New()
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		logger.Error("Error running program", "error", err)
		os.Exit(1)
	}
}
