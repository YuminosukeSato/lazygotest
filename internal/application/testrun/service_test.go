package testrun_test

import (
	"context"
	"errors"
	"testing"

	"github.com/s21066/lazygotest/internal/application/testrun"
	"github.com/s21066/lazygotest/internal/domain/execution"
	"github.com/s21066/lazygotest/internal/interfaces/process"
)

type fakeRunner struct {
	result process.TestResult
	err    error
	calls  []struct {
		importPath string
		testName   string
	}
}

func (f *fakeRunner) Run(_ context.Context, importPath, testName string) (process.TestResult, error) {
	f.calls = append(f.calls, struct {
		importPath string
		testName   string
	}{importPath: importPath, testName: testName})
	return f.result, f.err
}

func TestServiceSuccessPopulatesLogsAndStatus(t *testing.T) {
	t.Parallel()
	runner := &fakeRunner{
		result: process.TestResult{
			Status: process.RunStatusPass,
			Events: []process.TestEvent{
				{Action: "run", Test: "TestFoo"},
				{Action: "output", Output: "hello\n"},
				{Action: "pass", Test: "TestFoo"},
			},
		},
	}
	svc := testrun.NewService(runner)

	run, err := svc.Run(context.Background(), "github.com/me/a", "TestFoo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if run.Status() != execution.StatusSuccess {
		t.Fatalf("expected success, got %s", run.Status())
	}
	if got := len(run.Logs()); got != 1 {
		t.Fatalf("expected 1 log line, got %d", got)
	}
	if got := runner.calls[0].testName; got != "TestFoo" {
		t.Fatalf("expected test name forwarded, got %s", got)
	}
}

func TestServiceFailureMarksRunFailed(t *testing.T) {
	t.Parallel()
	runner := &fakeRunner{
		result: process.TestResult{
			Status: process.RunStatusFail,
			Events: []process.TestEvent{
				{Action: "fail", Test: "TestBar", Output: "boom\n"},
			},
		},
	}
	svc := testrun.NewService(runner)

	run, err := svc.Run(context.Background(), "github.com/me/a", "TestBar")
	if err == nil {
		t.Fatalf("expected error but got nil")
	}
	if run.Status() != execution.StatusFailed {
		t.Fatalf("expected failed, got %s", run.Status())
	}
}

func TestServicePropagatesRunnerError(t *testing.T) {
	t.Parallel()
	wantErr := errors.New("exec failed")
	runner := &fakeRunner{err: wantErr}
	svc := testrun.NewService(runner)

	_, err := svc.Run(context.Background(), "github.com/me/a", "TestBar")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}
