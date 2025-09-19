# lazygotest

A LazyGit-inspired TUI for running Go tests with an interactive, keyboard-driven interface.

## Features

- **Interactive Test Selection**: Navigate and select tests using keyboard shortcuts
- **Real-time Output**: Watch test results stream in as they run
- **Failed Test Management**: Quickly re-run only failed tests
- **Smart Filtering**: Filter tests by name or package
- **Test Flags**: Toggle common flags like `-race`, `-cover`, `-short`
- **Split View**: Dedicated panels for tests, flags, and logs
- **Keyboard-driven**: Efficient workflow without leaving the keyboard

## Installation

### Homebrew (macOS/Linux)

```bash
brew install s21066/tap/lazygotest
```

### Go Install

```bash
go get github.com/s21066/lazygotest/cmd/lazygotest@latest
```

### Direct Download

Download the binary for your platform from the [latest release](https://github.com/s21066/lazygotest/releases/latest).

### Build from Source

```bash
git clone https://github.com/s21066/lazygotest.git
cd lazygotest
go build -o lazygotest ./cmd/lazygotest
```

## Usage

### Quick Start

```bash
# Run in your Go project
lazygotest ./...

# Test specific packages
lazygotest ./internal/... ./pkg/...
```

### Command Line Options

```bash
lazygotest [flags] [packages]

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
  -version        Print version information
```

### Keyboard Shortcuts

#### Navigation
- `Tab` / `Shift+Tab` - Switch between panels
- `j` / `k` or `↓` / `↑` - Navigate test list
- `Space` - Toggle test selection
- `Enter` - Focus logs panel
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
- `s` - Save logs (planned)
- `o` - Open in editor (planned)

## UI Overview

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

## Architecture

Built with Clean Architecture principles:

```
lazygotest/
├── cmd/lazygotest/           # CLI entry point
├── internal/
│   ├── adapter/         # Interface adapters
│   │   ├── primary/     # UI components (TUI)
│   │   └── secondary/   # Infrastructure (runner, config, etc.)
│   ├── usecase/         # Business logic
│   └── shared/          # Cross-cutting concerns
└── pkg/                 # Reusable utilities
```

### Design Principles

- **Clean Architecture**: Separation of concerns with clear boundaries
- **Dependency Inversion**: Business logic doesn't depend on UI or infrastructure
- **Event-driven**: Components communicate via events
- **Testable**: Each layer can be tested independently

## Development

### Requirements

- Go 1.22.1+
- Terminal with 256-color support

### Running Tests

```bash
go test ./...
```

### Debug Mode

```bash
lazygotest -debug ./...
# Logs written to debug.log
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
