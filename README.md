# lazygotest

A beautiful TUI for running Go tests interactively, inspired by lazygit.

## Features
- **lazygit-inspired color scheme** - Easy on the eyes with WCAG-compliant contrast
- **Smart directory detection** - Auto-finds go.mod in parent directories
- **Interactive test runner** - Browse and run tests with vim-style keybindings
- **Real-time feedback** - See test results and logs as they happen
- **Accessible design** - Color-blind friendly with icons and clear focus indicators

## Installation

### Via go install (recommended)

```bash
go install github.com/s21066/lazygotest/cmd/lazygotest@latest
```

### From source

```bash
git clone https://github.com/s21066/lazygotest.git
cd lazygotest
make install
```

## Quick Start

```bash
# Run in current directory (searches for go.mod in parents)
lazygotest

# Run in specific directory
lazygotest .
lazygotest /path/to/project
```

### For Development

```bash
# Build the application
make build

# Run the TUI application (requires interactive terminal)
make run

# Or run directly
./bin/lazygotest
```

### Available Make Targets

```bash
make help        # Show all available targets
make build       # Build the binary
make test        # Run all tests
make test-tview  # Run tview-specific tests
make run         # Build and run the application
make install     # Install lazygotest to $GOPATH/bin
make clean       # Remove build artifacts and cache
make lint        # Run golangci-lint
make fmt         # Format code
make all         # Format, lint, test, and build
```

## TUI Features

### lazygit-inspired Theme
- **Catppuccin Mocha colors** - Soothing dark theme with excellent contrast
- **Soft pastels** - Easy on the eyes during long coding sessions
- **Consistent color language** - Success/failure/running states use distinct, recognizable colors

### Universal Design Improvements
- **WCAG 2.1 compliant** - All text meets AA contrast ratios (4.5:1+)
- **Visual Status Indicators**: Icons (✓/✗/⏳) complement colors for color-blind accessibility
- **Focus Indicators**: Bright blue borders show which pane has focus
- **Titled Panes**: Each panel clearly labeled ( Packages / Tests / History / Log )
- **Status Bar**: Context-sensitive key bindings always visible
- **Multiple Input Methods**: Arrow keys and PageUp/PageDown alongside vim-style navigation

### Terminal Compatibility
- ✅ macOS Terminal.app
- ✅ iTerm2
- ✅ Warp
- ✅ Alacritty
- ✅ kitty
- ⚠️ Requires interactive TTY (won't work in CI/CD without TTY)

### Keyboard Shortcuts

**Navigation:**
- `↑↓` or `j/k` - Move up/down
- `←→` or `h/l` - Move left/right (between panes)
- `gg` or `PageUp` - Jump to top
- `G` or `PageDown` - Jump to bottom
- `Tab` / `Shift-Tab` - Switch focus between panes

**Actions:**
- `Enter` - Run selected test
- `r` - Rerun last test
- `/` - Open filter dialog
- `?` - Show help
- `q` or `Esc` - Quit

**Visual Feedback:**
- Focused pane shows bright blue border
- Test results show icons: ✓ (pass), ✗ (fail), ⏳ (running)
- Status bar displays context-sensitive key bindings

## Usage
1. Run `./scripts/check.sh` (defaults to PROFILE=unit).
   - `PROFILE=race` adds `-race`, `PROFILE=cover` adds `-cover`, `PROFILE=short` adds `-short`, `RUN=TestName` adds `-run ^TestName$`.
2. Start the TUI: `make run` or `./bin/lazygotest`
3. Use keyboard shortcuts above to navigate and run tests.

## Development
- Follow TDD: add a failing test, implement the minimum, then refactor.
- Keep PRs small (≤20 lines) and avoid `git add .`; stage files explicitly.
- No external network required once module cache is prepared.
