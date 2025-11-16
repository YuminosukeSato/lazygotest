package catalog_test

import (
	"testing"

	"github.com/s21066/lazygotest/internal/domain/catalog"
)

func TestBuildSnapshotSuccess(t *testing.T) {
	t.Parallel()
	snap, err := catalog.BuildSnapshot([]catalog.ModuleInput{
		{
			Name:    "mod",
			RootDir: "/repo",
			Packages: []catalog.PackageInput{
				{
					ImportPath: "github.com/me/a",
					Dir:        "/repo/a",
					Tests: []catalog.TestInput{
						{Name: "TestFoo", File: "foo_test.go", Line: 12, Kind: catalog.TestKindUnit},
						{Name: "BenchmarkFoo", File: "foo_test.go", Line: 42, Kind: catalog.TestKindBenchmark},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	mods := snap.Modules()
	if len(mods) != 1 {
		t.Fatalf("expected single module, got %d", len(mods))
	}
	if got := len(mods[0].Packages()); got != 1 {
		t.Fatalf("expected single package, got %d", got)
	}
}
