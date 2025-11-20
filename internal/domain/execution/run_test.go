package execution_test

import (
	"testing"

	"github.com/s21066/lazygotest/internal/domain/execution"
)

func TestRunStartsAsRunning(t *testing.T) {
	t.Parallel()
	run := execution.NewRun("pkg", "TestFoo")
	if run.Status() != execution.StatusRunning {
		t.Fatalf("expected running, got %s", run.Status())
	}
}

func TestRunCompleteSuccess(t *testing.T) {
	t.Parallel()
	run := execution.NewRun("pkg", "TestFoo")
	run.Complete(nil)
	if run.Status() != execution.StatusSuccess {
		t.Fatalf("expected success, got %s", run.Status())
	}
}

func TestRunCompleteFailure(t *testing.T) {
	t.Parallel()
	run := execution.NewRun("pkg", "TestFoo")
	run.Complete(assertErr("boom"))
	if run.Status() != execution.StatusFailed {
		t.Fatalf("expected failed, got %s", run.Status())
	}
}

func TestRunDuration(t *testing.T) {
	t.Parallel()
	run := execution.NewRun("pkg", "TestFoo")
	run.AddDuration(0.1)
	run.AddDuration(-1) // ignored
	if got := run.Duration(); got != 0.1 {
		t.Fatalf("expected duration 0.1, got %f", got)
	}
}

func TestFormattedDuration(t *testing.T) {
	t.Parallel()
	run := execution.NewRun("pkg", "TestFoo")
	if run.FormattedDuration() != "0ms" {
		t.Fatalf("expected 0ms")
	}
	run.AddDuration(0.05)
	if run.FormattedDuration() != "50ms" {
		t.Fatalf("expected 50ms, got %s", run.FormattedDuration())
	}
	run.AddDuration(0.25)
	if run.FormattedDuration() != "0.3s" {
		t.Fatalf("expected 0.3s, got %s", run.FormattedDuration())
	}
	run.AddDuration(1.2)
	if run.FormattedDuration() != "1.50s" {
		t.Fatalf("expected 1.50s, got %s", run.FormattedDuration())
	}
}

func assertErr(msg string) error {
	return &testError{msg: msg}
}

type testError struct {
	msg string
}

func (e *testError) Error() string { return e.msg }
