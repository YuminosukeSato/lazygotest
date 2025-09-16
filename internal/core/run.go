package core

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os/exec"
)

// Runner は go test プロセスを起動し、JSON イベントをストリームで返す。
// How: `go test -json` を実行し、1 行ごとに JSON をデコードして ch に流す。
// Why not: MVP では単一プロセス・逐次実行のみ。並列や pkg 分割は後続対応。
type Runner struct{}

type RunFlags struct {
	Race    bool
	Cover   bool
	Short   bool
	Tags    []string
	Timeout string // 例: 2m
}

func (r *Runner) Run(ctx context.Context, pkgs []string, runRegexp string, flags RunFlags, out chan<- TestEvent) error {
	args := []string{"test", "-json", "-v"}
	if runRegexp != "" {
		args = append(args, "-run", runRegexp)
	}
	if flags.Race {
		args = append(args, "-race")
	}
	if flags.Cover {
		args = append(args, "-cover", "-coverprofile", "/tmp/lazytest.cover")
	}
	if flags.Short {
		args = append(args, "-short")
	}
	if flags.Timeout != "" {
		args = append(args, "-timeout", flags.Timeout)
	}
	if len(flags.Tags) > 0 {
		args = append(args, "-tags", joinComma(flags.Tags))
	}
	args = append(args, pkgs...)

	cmd := exec.CommandContext(ctx, "go", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	// stderr はそのまま捨てずに output として流す（簡易）
	go forwardLinesAsOutput(stderr, out)

	// 標準出力の JSON を読み出す
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		var ev TestEvent
		if err := json.Unmarshal(scanner.Bytes(), &ev); err != nil {
			// 壊れた行は output として扱う
			out <- TestEvent{Action: "output", Output: string(scanner.Bytes())}
			continue
		}
		out <- ev
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return cmd.Wait()
}

func forwardLinesAsOutput(r io.Reader, out chan<- TestEvent) {
	s := bufio.NewScanner(r)
	for s.Scan() {
		out <- TestEvent{Action: "output", Output: s.Text()}
	}
}

func joinComma(ss []string) string {
	if len(ss) == 0 {
		return ""
	}
	out := ss[0]
	for i := 1; i < len(ss); i++ {
		out += "," + ss[i]
	}
	return out
}
