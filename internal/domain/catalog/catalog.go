package catalog

import (
	"errors"
	"sort"
)

var (
	ErrEmptyModuleName   = errors.New("module name must not be empty")
	ErrEmptyPackagePath  = errors.New("package import path must not be empty")
	ErrDuplicateModule   = errors.New("module already registered")
	ErrDuplicatePackage  = errors.New("package already registered within module")
	ErrInvalidTestName   = errors.New("test name must not be empty")
	ErrInvalidTestKind   = errors.New("test kind is not allowed")
	ErrInvalidModuleRoot = errors.New("module root directory must not be empty")
	ErrInvalidPackageDir = errors.New("package directory must not be empty")
)

type TestKind string

const (
	TestKindUnit      TestKind = "unit"
	TestKindBenchmark TestKind = "benchmark"
	TestKindExample   TestKind = "example"
	TestKindFuzz      TestKind = "fuzz"
)

type Snapshot struct {
	modules []Module
}

func NewSnapshot(modules []Module) Snapshot {
	cp := make([]Module, len(modules))
	copy(cp, modules)
	return Snapshot{modules: cp}
}

func (s Snapshot) Modules() []Module {
	cp := make([]Module, len(s.modules))
	copy(cp, s.modules)
	return cp
}

type Module struct {
	name     string
	rootDir  string
	packages []Package
}

func NewModule(name, rootDir string, pkgs []Package) (Module, error) {
	switch {
	case name == "":
		return Module{}, ErrEmptyModuleName
	case rootDir == "":
		return Module{}, ErrInvalidModuleRoot
	}

	seen := make(map[string]struct{}, len(pkgs))
	out := make([]Package, len(pkgs))
	for i, pkg := range pkgs {
		if pkg.importPath == "" {
			return Module{}, ErrEmptyPackagePath
		}
		if _, ok := seen[pkg.importPath]; ok {
			return Module{}, ErrDuplicatePackage
		}
		seen[pkg.importPath] = struct{}{}
		out[i] = pkg
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].importPath < out[j].importPath
	})
	return Module{
		name:     name,
		rootDir:  rootDir,
		packages: out,
	}, nil
}

func (m Module) Name() string    { return m.name }
func (m Module) RootDir() string { return m.rootDir }
func (m Module) Packages() []Package {
	cp := make([]Package, len(m.packages))
	copy(cp, m.packages)
	return cp
}

type Package struct {
	importPath string
	dir        string
	tests      []TestCase
}

func NewPackage(importPath, dir string, tests []TestCase) (Package, error) {
	switch {
	case importPath == "":
		return Package{}, ErrEmptyPackagePath
	case dir == "":
		return Package{}, ErrInvalidPackageDir
	}
	out := make([]TestCase, len(tests))
	copy(out, tests)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Kind != out[j].Kind {
			return out[i].Kind < out[j].Kind
		}
		return out[i].Name < out[j].Name
	})
	return Package{
		importPath: importPath,
		dir:        dir,
		tests:      out,
	}, nil
}

func (p Package) ImportPath() string { return p.importPath }
func (p Package) Dir() string        { return p.dir }
func (p Package) Tests() []TestCase {
	cp := make([]TestCase, len(p.tests))
	copy(cp, p.tests)
	return cp
}

type TestCase struct {
	Name string
	File string
	Line int
	Kind TestKind
}

func NewTestCase(name, file string, line int, kind TestKind) (TestCase, error) {
	if name == "" {
		return TestCase{}, ErrInvalidTestName
	}
	switch kind {
	case TestKindUnit, TestKindBenchmark, TestKindExample, TestKindFuzz:
	default:
		return TestCase{}, ErrInvalidTestKind
	}
	return TestCase{
		Name: name,
		File: file,
		Line: line,
		Kind: kind,
	}, nil
}
