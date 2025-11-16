package catalog

import "fmt"

type ModuleInput struct {
	Name     string
	RootDir  string
	Packages []PackageInput
}

type PackageInput struct {
	ImportPath string
	Dir        string
	Tests      []TestInput
}

type TestInput struct {
	Name string
	File string
	Line int
	Kind TestKind
}

func BuildSnapshot(inputs []ModuleInput) (Snapshot, error) {
	modules := make([]Module, 0, len(inputs))
	seen := make(map[string]struct{}, len(inputs))
	for _, in := range inputs {
		if in.Name == "" {
			return Snapshot{}, ErrEmptyModuleName
		}
		if _, ok := seen[in.Name]; ok {
			return Snapshot{}, fmt.Errorf("%w: %s", ErrDuplicateModule, in.Name)
		}
		seen[in.Name] = struct{}{}
		pkgs := make([]Package, 0, len(in.Packages))
		for _, pkgIn := range in.Packages {
			tests := make([]TestCase, 0, len(pkgIn.Tests))
			for _, testIn := range pkgIn.Tests {
				tc, err := NewTestCase(testIn.Name, testIn.File, testIn.Line, testIn.Kind)
				if err != nil {
					return Snapshot{}, err
				}
				tests = append(tests, tc)
			}
			pkg, err := NewPackage(pkgIn.ImportPath, pkgIn.Dir, tests)
			if err != nil {
				return Snapshot{}, err
			}
			pkgs = append(pkgs, pkg)
		}
		mod, err := NewModule(in.Name, in.RootDir, pkgs)
		if err != nil {
			return Snapshot{}, err
		}
		modules = append(modules, mod)
	}
	return NewSnapshot(modules), nil
}
