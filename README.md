# lazygotest

Command-line helpers to catalog and run Go tests in a TDD-friendly way.  

## Features
- Build a test catalog from `go list -json ./...` and `go test -list` without running tests.
- Classify unit, benchmark, example, and fuzz tests based on names.
- Stream `go test -json` output and keep structured events.
- Offline-friendly: uses workspace caches (`GOCACHE`, `GOLANGCI_LINT_CACHE`).

## Usage
1. Run `./scripts/check.sh` (defaults to PROFILE=unit).  
   - `PROFILE=race` adds `-race`, `PROFILE=cover` adds `-cover`, `PROFILE=short` adds `-short`.
2. Integrate `internal/interfaces/process` adapters into your TUI or CLI.

## Development
- Follow TDD: add a failing test, implement the minimum, then refactor.
- Keep PRs small (≤20 lines) and avoid `git add .`; stage files explicitly.
- No external network required once module cache is prepared.
