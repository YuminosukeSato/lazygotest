package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"lazygotest/internal/app"
	"lazygotest/internal/config"
)

// main は CLI 入口。How: フラグを解釈して App を起動する。
// Why not: 依存を最小化するため、ここでは外部 UI 依存の詳細を持たない。
func main() {
	var (
		watch    = flag.Bool("watch", false, "enable file watch")
		cover    = flag.Bool("cover", false, "enable coverage")
		race     = flag.Bool("race", false, "enable -race")
		short    = flag.Bool("short", false, "enable -short")
		timeout  = flag.Duration("timeout", 2*time.Minute, "go test timeout")
		tags     = flag.String("tags", "", "build tags (comma separated)")
		editor   = flag.String("editor", "", "editor command (default $EDITOR or nvim)")
		eargs    = flag.String("editor-args", "+{line} {file}", "editor args template")
		parallel = flag.Int("p", 1, "package-level parallelism")
		debug    = flag.Bool("debug", false, "enable debug logs")
	)
	flag.Parse()

	cfg := config.Config{
		Watch:      *watch,
		Cover:      *cover,
		Race:       *race,
		Short:      *short,
		Timeout:    *timeout,
		Tags:       splitComma(*tags),
		Editor:     firstNonEmpty(*editor, os.Getenv("EDITOR"), "nvim"),
		EditorArgs: *eargs,
		Parallel:   *parallel,
		Debug:      *debug,
		Packages:   flag.Args(),
	}

	if err := app.Run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func splitComma(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
