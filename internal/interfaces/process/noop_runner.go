package process

import "context"

// NoopCmdRunner returns empty output; useful for wiring tests and headless main.
type NoopCmdRunner struct{}

func NewNoopCmdRunner() *NoopCmdRunner { return &NoopCmdRunner{} }

func (r *NoopCmdRunner) Run(_ context.Context, _ string, _ ...string) (string, error) {
	return "", nil
}
