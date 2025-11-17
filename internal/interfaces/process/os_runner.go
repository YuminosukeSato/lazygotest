package process

import (
	"context"
	"os/exec"
	"strings"
)

// OSRunner executes commands using exec.CommandContext.
type OSRunner struct{}

func NewOSRunner() *OSRunner { return &OSRunner{} }

func (r *OSRunner) Run(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.CombinedOutput()
	return strings.TrimRight(string(out), "\n"), err
}
