package process_test

import (
	"context"
	"strings"
	"testing"

	"github.com/s21066/lazygotest/internal/domain/catalog"
	"github.com/s21066/lazygotest/internal/interfaces/process"
)

type fakeRunner struct {
	outputs []string
	err     error
	calls   []struct {
		dir  string
		args []string
	}
}

func (f *fakeRunner) Run(_ context.Context, dir string, args ...string) (string, error) {
	call := struct {
		dir  string
		args []string
	}{dir: dir, args: append([]string(nil), args...)}
	f.calls = append(f.calls, call)
	if len(f.outputs) == 0 {
		return "", f.err
	}
	out := f.outputs[0]
	f.outputs = f.outputs[1:]
	return out, f.err
}

func TestGoCmdSourceCollectsModulesAndTests(t *testing.T) {
	t.Parallel()
	goListOut := strings.Join([]string{
		`{"ImportPath":"github.com/me/a","Dir":"/repo/a","TestGoFiles":["a_test.go"],"Module":{"Path":"mod","Dir":"/repo"}}`,
		`{"ImportPath":"github.com/me/b","Dir":"/repo/b","TestGoFiles":["b_test.go"],"Module":{"Path":"mod","Dir":"/repo"}}`,
	}, "\n")
	goTestListA := "TestFoo\nBenchmarkFoo\n"
	goTestListB := "FuzzBar\nExampleBar\n"

	runner := &fakeRunner{
		outputs: []string{goListOut, goTestListA, goTestListB},
	}
	src := process.NewGoCmdSource("/repo", runner)

	mods, err := src.Modules(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mods) != 1 {
		t.Fatalf("expected single module, got %d", len(mods))
	}
	pkgs := mods[0].Packages
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if got := pkgs[0].Tests[0].Kind; got != catalog.TestKindUnit {
		t.Fatalf("expected unit kind, got %s", got)
	}
	if got := pkgs[0].Tests[1].Kind; got != catalog.TestKindBenchmark {
		t.Fatalf("expected benchmark kind, got %s", got)
	}
	if got := pkgs[1].Tests[0].Kind; got != catalog.TestKindFuzz {
		t.Fatalf("expected fuzz kind, got %s", got)
	}
	if got := pkgs[1].Tests[1].Kind; got != catalog.TestKindExample {
		t.Fatalf("expected example kind, got %s", got)
	}
}
