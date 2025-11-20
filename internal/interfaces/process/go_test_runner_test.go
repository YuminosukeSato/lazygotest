package process_test

import (
	"context"
	"strings"
	"testing"

	"github.com/YuminosukeSato/lazygotest/internal/interfaces/process"
)

func TestGoTestRunnerParsesJSONStream(t *testing.T) {
	t.Parallel()
	stream := strings.Join([]string{
		`{"Time":"2024-01-01T00:00:00Z","Action":"run","Package":"github.com/me/a","Test":"TestFoo"}`,
		`{"Time":"2024-01-01T00:00:00Z","Action":"output","Package":"github.com/me/a","Test":"TestFoo","Output":"hello\\n"}`,
		`{"Time":"2024-01-01T00:00:01Z","Action":"pass","Package":"github.com/me/a","Test":"TestFoo","Elapsed":0.1}`,
		`{"Time":"2024-01-01T00:00:01Z","Action":"pass","Package":"github.com/me/a","Elapsed":0.2}`,
	}, "\n")
	runner := &fakeRunner{outputs: []string{stream}}
	rt := process.NewGoTestRunner(runner)

	result, err := rt.Run(context.Background(), "github.com/me/a", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != process.RunStatusPass {
		t.Fatalf("expected pass, got %s", result.Status)
	}
	if len(result.Events) != 4 {
		t.Fatalf("expected 4 events, got %d", len(result.Events))
	}
	if got := result.Events[1].Output; got != "hello\n" {
		t.Fatalf("expected output preserved, got %q", got)
	}
	if diff := result.Duration - 0.3; diff < -1e-9 || diff > 1e-9 {
		t.Fatalf("expected duration 0.3, got %f", result.Duration)
	}
}

func TestGoTestRunnerPropagatesFailure(t *testing.T) {
	t.Parallel()
	stream := `{"Action":"fail","Package":"pkg"}`
	runner := &fakeRunner{outputs: []string{stream}}
	rt := process.NewGoTestRunner(runner)

	result, err := rt.Run(context.Background(), "pkg", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != process.RunStatusFail {
		t.Fatalf("expected fail, got %s", result.Status)
	}
}
