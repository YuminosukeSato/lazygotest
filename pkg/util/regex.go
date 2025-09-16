package util

import (
	"regexp"
	"strings"
)

// BuildRunRegex は go test -run 用の正規表現を生成する。
// How: 名前を QuoteMeta した上で | で結合し ^$ で全体一致させる。
// Why not: 巨大な集合最適化は後続対応（pkg 分割など）。
func BuildRunRegex(names []string) string {
	if len(names) == 0 {
		return ""
	}
	esc := make([]string, len(names))
	for i, n := range names {
		esc[i] = regexp.QuoteMeta(n)
	}
	return "^(" + strings.Join(esc, "|") + ")$"
}
