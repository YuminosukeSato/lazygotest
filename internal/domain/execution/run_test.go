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

func assertErr(msg string) error {
	return &testError{msg: msg}
}

type testError struct {
	msg string
}

func (e *testError) Error() string { return e.msg }
