#!/usr/bin/env bash
set -euo pipefail

# Reason: run tests and lint with workspace-local caches to avoid permission/network issues.
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

export GOCACHE="${REPO_ROOT}/.cache/go-build"
export GOLANGCI_LINT_CACHE="${REPO_ROOT}/.cache/golangci-lint"
export GOTOOLCHAIN=local

cd "${REPO_ROOT}"

echo "== go test =="
go test ./...

echo "== golangci-lint =="
golangci-lint run ./...
