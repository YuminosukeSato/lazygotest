package process

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type RunStatus string

const (
	RunStatusPass RunStatus = "pass"
	RunStatusFail RunStatus = "fail"
	RunStatusRun  RunStatus = "run"
)

type TestEvent struct {
	Action  string  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test,omitempty"`
	Output  string  `json:"Output,omitempty"`
	Elapsed float64 `json:"Elapsed,omitempty"`
}

type TestResult struct {
	Status RunStatus
	Events []TestEvent
}

type GoTestRunner struct {
	runner CmdRunner
}

func NewGoTestRunner(runner CmdRunner) GoTestRunner {
	return GoTestRunner{runner: runner}
}

func (g GoTestRunner) Run(ctx context.Context, importPath, testName string) (TestResult, error) {
	args := []string{"go", "test", "-json"}
	if testName != "" {
		args = append(args, "-run", fmt.Sprintf("^%s$", testName))
	}
	args = append(args, importPath)
	out, err := g.runner.Run(ctx, "", args...)
	if err != nil {
		return TestResult{}, err
	}
	dec := json.NewDecoder(strings.NewReader(out))
	events := make([]TestEvent, 0, 16)
	status := RunStatusPass
	for dec.More() {
		var ev TestEvent
		if err := dec.Decode(&ev); err != nil {
			return TestResult{}, fmt.Errorf("decode go test json: %w", err)
		}
		ev.Output = strings.ReplaceAll(ev.Output, "\\n", "\n")
		events = append(events, ev)
		status = nextStatus(status, ev.Action)
	}
	return TestResult{
		Status: status,
		Events: events,
	}, nil
}

func nextStatus(current RunStatus, action string) RunStatus {
	switch action {
	case "fail":
		return RunStatusFail
	case "pass":
		if current != RunStatusFail {
			return RunStatusPass
		}
	case "run":
		if current == RunStatusPass || current == RunStatusRun {
			return RunStatusRun
		}
	}
	return current
}
