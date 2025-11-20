# lazygotest

Command-line helpers to catalog and run Go tests in a TDD-friendly way.

## Features
- Build a test catalog from `go list -json ./...` and `go test -list` without running tests.
- Classify unit, benchmark, example, and fuzz tests based on names.
- Stream `go test -json` output and keep structured events.
- Offline-friendly: uses workspace caches (`GOCACHE`, `GOLANGCI_LINT_CACHE`).

## Quick Start

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
make clean       # Remove build artifacts and cache
make lint        # Run golangci-lint
make fmt         # Format code
make all         # Format, lint, test, and build
```

## Usage
1. Run `./scripts/check.sh` (defaults to PROFILE=unit).  
   - `PROFILE=race` adds `-race`, `PROFILE=cover` adds `-cover`, `PROFILE=short` adds `-short`, `RUN=TestName` adds `-run ^TestName$`.
2. Start the headless UI stub: `go run ./cmd`. (It wires adapters; real tview primitives can be swapped easily.)
3. Keymap (vim-like, simple English): `h/j/k/l` move, `gg`/`G` top/bottom, `Enter` run selection, `r` rerun last, `/` reset filter, `Tab`/`Shift-Tab` move focus, `?` help (placeholder), `q` quit (caller handles exit).
4. Adapters: `internal/ui/tview_adapter.go` provides minimal implementations of Tree/List/History/Log panes; replace with tview primitives in production.
5. tview build: `go run -tags tview ./cmd` uses real tview primitives (vendor 同梱). `/` opens a filter modal, `?` opens a help overlay.

## Development
- Follow TDD: add a failing test, implement the minimum, then refactor.
- Keep PRs small (≤20 lines) and avoid `git add .`; stage files explicitly.
- No external network required once module cache is prepared.
