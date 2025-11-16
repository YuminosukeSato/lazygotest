package catalog_test

import (
	"testing"

	"github.com/s21066/lazygotest/internal/domain/catalog"
)

func TestBuildSnapshotRejectsDuplicateModules(t *testing.T) {
	t.Parallel()
	_, err := catalog.BuildSnapshot([]catalog.ModuleInput{
		{Name: "mod", RootDir: "/repo", Packages: []catalog.PackageInput{}},
		{Name: "mod", RootDir: "/repo2", Packages: []catalog.PackageInput{}},
	})
	if err == nil {
		t.Fatalf("expected error for duplicate module")
	}
}

func TestPackageSorting(t *testing.T) {
	t.Parallel()
	tc, err := catalog.NewTestCase("TestA", "a_test.go", 10, catalog.TestKindUnit)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	pkgA, err := catalog.NewPackage("github.com/me/a", "/repo/a", []catalog.TestCase{tc})
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	pkgB, err := catalog.NewPackage("github.com/me/b", "/repo/b", []catalog.TestCase{tc})
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	mod, err := catalog.NewModule("mod", "/repo", []catalog.Package{pkgB, pkgA})
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	pkgs := mod.Packages()
	if pkgs[0].ImportPath() != "github.com/me/a" {
		t.Fatalf("expected packages sorted, got %v", pkgs)
	}
}
