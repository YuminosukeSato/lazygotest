package testrun

import (
	"context"
	"errors"
	"fmt"

	"github.com/s21066/lazygotest/internal/domain/execution"
	"github.com/s21066/lazygotest/internal/interfaces/process"
)

// Runner abstracts go test -json executor.
type Runner interface {
	Run(ctx context.Context, importPath, testName string) (process.TestResult, error)
}

type Service struct {
	runner Runner
}

func NewService(runner Runner) Service {
	return Service{runner: runner}
}

func (s Service) Run(ctx context.Context, importPath, testName string) (execution.Run, error) {
	res, err := s.runner.Run(ctx, importPath, testName)
	if err != nil {
		return execution.Run{}, err
	}
	run := execution.NewRun(importPath, testName)
	for _, ev := range res.Events {
		if ev.Output != "" {
			run.AppendLog(ev.Output)
		}
	}
	run.AddDuration(res.Duration)
	switch res.Status {
	case process.RunStatusPass:
		run.Complete(nil)
		return run, nil
	case process.RunStatusFail:
		run.Complete(errors.New("test failed"))
		return run, fmt.Errorf("test failed: %s %s", importPath, testName)
	default:
		return run, fmt.Errorf("unknown status: %s", res.Status)
	}
}
