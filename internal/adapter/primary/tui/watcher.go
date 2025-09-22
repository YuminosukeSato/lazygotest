package tui

import (
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
	"github.com/YuminosukeSato/lazygotest/pkg/logger"
)

// fileChangeMsg is sent when a watched file changes
type fileChangeMsg struct {
	path string
	op   fsnotify.Op
}

// watcherStartedMsg is sent when the watcher starts successfully
type watcherStartedMsg struct{}

// watcherStoppedMsg is sent when the watcher stops
type watcherStoppedMsg struct{}

// watcherErrorMsg is sent when the watcher encounters an error
type watcherErrorMsg struct {
	err error
}

// startWatcher starts file system watching
func (m *Model) startWatcher() tea.Cmd {
	return func() tea.Msg {
		// Create new watcher
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			logger.Error("Failed to create watcher", "error", err)
			return watcherErrorMsg{err: err}
		}
		
		m.watcher = watcher
		m.watcherStop = make(chan struct{})
		
		// Add directories to watch
		err = filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			
			// Skip hidden directories and vendor
			if d.IsDir() {
				name := d.Name()
				if strings.HasPrefix(name, ".") || name == "vendor" || name == "node_modules" {
					return filepath.SkipDir
				}
				
				// Watch this directory
				if err := watcher.Add(path); err != nil {
					logger.Warn("Failed to watch directory", "path", path, "error", err)
				}
			}
			
			return nil
		})
		
		if err != nil {
			logger.Error("Failed to walk directories", "error", err)
			return watcherErrorMsg{err: err}
		}
		
		// Start goroutine to handle events
		go m.watchFiles()
		
		logger.Info("File watcher started")
		return watcherStartedMsg{}
	}
}

// stopWatcher stops file system watching
func (m *Model) stopWatcher() tea.Cmd {
	return func() tea.Msg {
		if m.watcher != nil {
			// Signal the watcher goroutine to stop
			if m.watcherStop != nil {
				close(m.watcherStop)
				m.watcherStop = nil
			}
			
			// Close the watcher
			if err := m.watcher.Close(); err != nil {
				logger.Error("Failed to close watcher", "error", err)
			}
			m.watcher = nil
			
			logger.Info("File watcher stopped")
		}
		return watcherStoppedMsg{}
	}
}

// watchFiles is the goroutine that monitors file changes
func (m *Model) watchFiles() {
	// Debounce timer to avoid multiple rapid events
	var debounceTimer *time.Timer
	debounceDelay := 300 * time.Millisecond
	
	// Track recently seen events to avoid duplicates
	recentEvents := make(map[string]time.Time)
	
	for {
		select {
		case event, ok := <-m.watcher.Events:
			if !ok {
				return
			}
			
			// Filter only Go files and test files
			if !strings.HasSuffix(event.Name, ".go") {
				continue
			}
			
			// Skip if this is a recent duplicate event
			if lastTime, exists := recentEvents[event.Name]; exists {
				if time.Since(lastTime) < 100*time.Millisecond {
					continue
				}
			}
			recentEvents[event.Name] = time.Now()
			
			// Only react to write and create events
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				logger.Debug("File changed", "file", event.Name, "op", event.Op)
				
				// Reset debounce timer
				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				
				// Set new timer
				debounceTimer = time.AfterFunc(debounceDelay, func() {
					// Send file change message to trigger test re-run
					if m.program != nil {
						m.program.Send(fileChangeMsg{
							path: event.Name,
							op:   event.Op,
						})
					}
				})
			}
			
		case err, ok := <-m.watcher.Errors:
			if !ok {
				return
			}
			logger.Error("Watcher error", "error", err)
			if m.program != nil {
				m.program.Send(watcherErrorMsg{err: err})
			}
			
		case <-m.watcherStop:
			// Stop signal received
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			return
		}
	}
}

// handleFileChange handles file change events
func (m *Model) handleFileChange(msg fileChangeMsg) tea.Cmd {
	// Don't trigger if tests are already running
	if m.isRunning {
		return nil
	}
	
	// Log the change
	m.appendDetail("File changed: " + msg.path)
	m.appendDetail("Re-running tests...")
	
	// Determine which package to test based on the changed file
	pkgPath := filepath.Dir(msg.path)
	
	// If we have a selected package and the file is in it, test that package
	if m.selectedPackage != nil {
		selectedPath := string(m.selectedPackage.ID)
		if strings.HasPrefix(pkgPath, selectedPath) {
			return m.runSelectedPackage()
		}
	}
	
	// Otherwise, run all tests
	return m.runAllTests()
}