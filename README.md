# gotui

An interactive terminal-based Go test runner with a modern TUI (Terminal User Interface) built using Bubble Tea and Clean Architecture.

## Features

- **Interactive Test Selection**: Browse and select tests to run with keyboard navigation
- **Real-time Test Results**: Watch test results as they stream in
- **Failed Test Management**: Easily re-run only failed tests
- **Powerful Filtering**: Filter tests by name or package
- **Test Flags Support**: Toggle common test flags like `-race`, `-cover`, `-short`
- **Multi-panel Interface**: Split view for tests, flags, and logs
- **Keyboard-driven Workflow**: Efficient navigation without leaving the keyboard

## Installation

### Using go install

```bash
go install github.com/yourusername/gotui/cmd/gotui@latest
```

### Building from source

```bash
git clone https://github.com/yourusername/gotui.git
cd gotui
go build -o gotui ./cmd/gotui
```

## Usage

### Basic Usage

Run in your Go project directory:

```bash
gotui ./...
```

Or specify packages:

```bash
gotui ./internal/... ./pkg/...
```

### Command Line Options

```bash
gotui [flags] [packages]

Flags:
  -watch          Enable file watch mode
  -cover          Enable coverage reporting
  -race           Enable race detector
  -short          Run short tests only
  -timeout duration    Test timeout (default 2m0s)
  -tags string    Build tags (comma separated)
  -editor string  Editor command (default $EDITOR or nvim)
  -p int          Package-level parallelism (default 1)
  -debug          Enable debug logging
```

### Keyboard Shortcuts

#### Navigation
- `Tab` / `Shift+Tab` - Switch between panels
- `j` / `k` or `↓` / `↑` - Move up/down in test list
- `Space` - Toggle test selection
- `Enter` - Focus on logs panel
- `q` / `Ctrl+C` - Quit

#### Test Execution
- `a` - Run all tests
- `r` - Run selected tests
- `R` - Run failed tests only
- `.` - Repeat last run

#### Filtering
- `/` - Open filter prompt
- `f` - Toggle failed-only view
- `Esc` - Clear filter

#### Test Flags
- `g` - Toggle -race flag
- `c` - Toggle -cover flag
- `b` - Toggle -bench flag
- `z` - Toggle -fuzz flag
- `w` - Toggle watch mode
- `t` - Set build tags

#### Other
- `?` - Toggle help
- `s` - Save logs (not implemented yet)
- `o` - Open in editor (not implemented yet)

## UI Layout Overview

```
┌─ TESTS (j/k move, l expand, h collapse, Space select, ? help) ─────────────┐
│  ↑ 0/128  [pkg] internal/foo                                               │
│    ✓ TestAdd             3ms     ⌁ flaky:2%                                │
│    ✗ TestDivZero         2ms     msg: want4 got5                           │
│  ▸ internal/bar                                                            │
│  ▸ internal/baz                                                            │
│  filter: /Div|Add/   chips: [failed][short]                                │
└────────────────────────────────────────────────────────────────────────────┘
┌─ FLAGS (t tags g:-race c cover b bench z fuzz w watch o open-editor) ─────┐
│  -run=^(TestDivZero|TestAdd)$   -tags=integration   -race=ON   -short=OFF  │
└────────────────────────────────────────────────────────────────────────────┘
┌─ LOGS (Enter/Tab focus, ↑↓ scroll, s save, C-f search) ───────────────────┐
│ === internal/foo · TestDivZero · FAIL (2ms)                               │
│ --- want: 4                                                               │
│ +++ got : 5                                                               │
│ cmp.Diff: (-want +got)                                                    │
│ -4                                                                        │
│ +5                                                                        │
└────────────────────────────────────────────────────────────────────────────┘
STATUS: ● Running pkg:foo (3/12) | PASS:10 FAIL:2 SKIP:0 | ⏱ 12.3s | R rerun failed  . repeat  q quit
```

### Visual Highlights

- Fixed color roles: PASS = green, FAIL = red, SKIP = yellow, RUNNING = blue
- Focused panel uses a thick border and darker background for lazygit-style clarity
- Columns are aligned to minimize eye travel; durations and counts are right-aligned
- Active flags are rendered as rounded chips so toggles remain visible at a glance
- The status bar always shows package progress, aggregated counts, and quick actions

## Project Structure (Clean Architecture)

```
gotui/
├── cmd/
│   └── gotui/         # Main CLI entry point
├── internal/
│   ├── adapter/       # Interface adapters
│   │   ├── primary/   # Primary adapters (UI)
│   │   │   └── tui/   # Terminal UI components
│   │   └── secondary/ # Secondary adapters (Infrastructure)
│   │       ├── config/     # Configuration management
│   │       ├── runner/     # Test execution
│   │       ├── pkgrepo/    # Package repository
│   │       ├── fswatch/    # File system watching
│   │       ├── editor/     # Editor integration
│   │       └── cover/      # Coverage analysis
│   ├── usecase/       # Business logic (use cases)
│   │   ├── interfaces.go   # Port interfaces
│   │   ├── run_tests.go    # Test execution use case
│   │   └── list_packages.go # Package listing use case
│   └── shared/        # Shared components
│       ├── eventbus/  # Event bus implementation
│       └── ring/      # Ring buffer for logs
└── pkg/
    ├── errors/        # Error handling utilities
    └── logger/        # Logging utilities
```

### Architecture Overview

This project follows **Clean Architecture** principles (Hexagonal Architecture):

- **Use Cases** (`internal/usecase/`): Core business logic, framework-independent
- **Adapters** (`internal/adapter/`): Interface between use cases and external world
  - **Primary** (driving): UI components that call use cases
  - **Secondary** (driven): Infrastructure components called by use cases
- **Shared** (`internal/shared/`): Cross-cutting concerns like event bus
- **Packages** (`pkg/`): Reusable utilities

## Development

### Requirements

- Go 1.22.1 or later
- Terminal with 256-color support

### Building

```bash
go build ./cmd/gotui
```

### Testing

```bash
go test ./...
```

### Linting

```bash
golangci-lint run ./...
```

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License

## Acknowledgments

Built with ❤️ using the excellent [Charm](https://charm.sh) TUI libraries.