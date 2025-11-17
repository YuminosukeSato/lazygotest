package process

import (
	"context"
	"strings"
)

// InMemoryRunner satisfies CmdRunner for tests/headless main; it records args and returns preset output.
type InMemoryRunner struct {
	Output string
	Err    error
	Calls  [][]string
}

func (r *InMemoryRunner) Run(_ context.Context, _ string, args ...string) (string, error) {
	r.Calls = append(r.Calls, append([]string(nil), args...))
	return r.Output, r.Err
}

// WithLines helps build JSON/event streams.
func (r *InMemoryRunner) WithLines(lines ...string) *InMemoryRunner {
	r.Output = strings.Join(lines, "\n")
	return r
}
