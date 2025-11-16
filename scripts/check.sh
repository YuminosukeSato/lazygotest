#!/usr/bin/env bash
set -euo pipefail

# Reason: run tests and lint with workspace-local caches to avoid permission/network issues.
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

export GOCACHE="${REPO_ROOT}/.cache/go-build"
export GOLANGCI_LINT_CACHE="${REPO_ROOT}/.cache/golangci-lint"
export GOTOOLCHAIN=local

cd "${REPO_ROOT}"

profile="${PROFILE:-unit}"
run_args=()
case "${profile}" in
  race)   run_args+=("-race");;
  cover)  run_args+=("-cover");;
  short)  run_args+=("-short");;
  *)      profile="unit";;
esac

echo "== go test (${profile}) =="
go test ${run_args[@]+"${run_args[@]}"} ./...

echo "== golangci-lint =="
golangci-lint run ./...
