package presentation_test

import (
	"testing"

	"github.com/YuminosukeSato/lazygotest/internal/domain/catalog"
	"github.com/YuminosukeSato/lazygotest/internal/domain/execution"
	"github.com/YuminosukeSato/lazygotest/internal/presentation"
)

func TestProjectBuildsThreePanes(t *testing.T) {
	t.Parallel()
	tc, _ := catalog.NewTestCase("TestFoo", "foo_test.go", 10, catalog.TestKindUnit)
	pkg, _ := catalog.NewPackage("github.com/me/a", "/repo/a", []catalog.TestCase{tc})
	mod, _ := catalog.NewModule("mod", "/repo", []catalog.Package{pkg})
	snap := catalog.NewSnapshot([]catalog.Module{mod})
	run := execution.NewRun("github.com/me/a", "TestFoo")
	run.AddDuration(0.3)
	run.Complete(nil)

	vm := presentation.Project(snap, []execution.Run{run})

	if len(vm.Packages) != 1 {
		t.Fatalf("expected 1 package node, got %d", len(vm.Packages))
	}
	if vm.Packages[0].Label != "github.com/me/a (1 tests)" {
		t.Fatalf("unexpected package label %q", vm.Packages[0].Label)
	}
	if len(vm.Tests) != 1 || vm.Tests[0].Name != "TestFoo" {
		t.Fatalf("expected single test row for TestFoo")
	}
	if len(vm.Runs) != 1 {
		t.Fatalf("expected single run row")
	}
	if vm.Runs[0].Duration != "0.3s" {
		t.Fatalf("expected duration 0.3s, got %s", vm.Runs[0].Duration)
	}
}
