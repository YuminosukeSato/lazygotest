#!/usr/bin/env bash
set -euo pipefail

# 使用法: fzf でパッケージ/テストを選んで go test -json を実行
# 依存: fzf, sed, paste, tee (jq があれば失敗のみ再実行も可)

PKG=$(go list ./... | fzf) || exit 1

REGEX=$(go test -list . "$PKG" \
  | sed -n 's/^\(Test\|Example\|Fuzz\)\w\+$/\0/p' \
  | fzf -m | paste -sd'|' -)
[ -z "${REGEX:-}" ] && exit 0

go test -v -json -run "^((${REGEX}))$" "$PKG" | tee /tmp/last-go-test.json

