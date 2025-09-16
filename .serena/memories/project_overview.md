Project: lazygotest
Purpose: Lazygit-like TUI to discover, filter, run, and re-run Go tests with live logs and coverage.
Tech Stack: Go 1.22; Charmstack (Bubble Tea, Bubbles, Lip Gloss) planned; fsnotify; go tool cover.
Status: MVP skeleton committed in local workspace: initial TUI placeholder, runner stub, types, config, and fzf script.
Structure:
- cmd/lazytest: CLI entry
- internal/app: Bubble Tea model and Run
- internal/ui: keymap and styles (later)
- internal/core: go test JSON parsing and runner
- internal/config: config struct
- pkg/types: shared types
- scripts/lazytest-fzf.sh: quick fzf-based workflow
Guidelines: Japanese-only comms; no git ops; format and lint after changes; ask y/n before execution; use Serena; prefer Gemini for quick research.
Entrypoints:
- App: go run ./cmd/lazytest
- Script: bash scripts/lazytest-fzf.sh
Testing/Formatting:
- Format: go fmt ./...
- Lint: prefer golangci-lint (future), fallback go vet ./... after deps installed
- Tests (tool runs tests): go test -json/-v flags
OS: Darwin (zsh)
Notes: Network downloads needed later for Charm deps.