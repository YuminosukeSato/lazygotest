.PHONY: all build test run clean lint fmt help

BIN_DIR := ./bin
BIN_NAME := lazygotest
GO_BUILD_FLAGS := -ldflags="-s -w"
GO_TEST_FLAGS := -v -race -cover

all: fmt lint test build

build:
	@echo "Building $(BIN_NAME)..."
	@mkdir -p $(BIN_DIR)
	go build $(GO_BUILD_FLAGS) -tags tview -o $(BIN_DIR)/$(BIN_NAME) ./cmd/main_tview.go

test:
	@echo "Running tests..."
	go test ./...

test-tview:
	@echo "Running tview tests..."
	go test -tags tview ./cmd -v -timeout 10s

run: build
	@echo "Running $(BIN_NAME)..."
	$(BIN_DIR)/$(BIN_NAME)

clean:
	@echo "Cleaning..."
	rm -rf $(BIN_DIR)
	go clean -cache -testcache

lint:
	@echo "Running golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install: https://golangci-lint.run/usage/install/"; \
	fi

fmt:
	@echo "Formatting code..."
	go fmt ./...
	gofmt -s -w .

help:
	@echo "Available targets:"
	@echo "  all        - Format, lint, test, and build (default)"
	@echo "  build      - Build the binary"
	@echo "  test       - Run all tests"
	@echo "  test-tview - Run tview-specific tests"
	@echo "  run        - Build and run the application"
	@echo "  clean      - Remove build artifacts and cache"
	@echo "  lint       - Run golangci-lint"
	@echo "  fmt        - Format code with go fmt"
	@echo "  help       - Show this help message"
