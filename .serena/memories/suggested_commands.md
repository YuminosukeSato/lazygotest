Common commands
- go run ./cmd/lazytest                       # launch TUI (after deps)
- bash scripts/lazytest-fzf.sh                # quick fzf-based test runner
- go fmt ./...                                # format
- go vet ./...                                # basic lint (after deps)
- go test -v -json ./... | tee out.json       # raw run for debugging
- rg -n "^func Test" -g "*_test.go"          # list tests (ripgrep)
- chmod +x scripts/lazytest-fzf.sh            # make the script executable
- uv --version                                # (if Python tooling used)
