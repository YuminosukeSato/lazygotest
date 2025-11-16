package process

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/s21066/lazygotest/internal/domain/catalog"
)

// CmdRunner は外部プロセス実行を抽象化する。
type CmdRunner interface {
	Run(ctx context.Context, dir string, args ...string) (string, error)
}

// GoCmdSource は go コマンドからテスト情報を収集する Source 実装。
type GoCmdSource struct {
	runner CmdRunner
	root   string
}

func NewGoCmdSource(root string, runner CmdRunner) GoCmdSource {
	return GoCmdSource{runner: runner, root: root}
}

type goListPackage struct {
	ImportPath   string   `json:"ImportPath"`
	Dir          string   `json:"Dir"`
	TestGoFiles  []string `json:"TestGoFiles"`
	XTestGoFiles []string `json:"XTestGoFiles"`
	Module       *struct {
		Path string `json:"Path"`
		Dir  string `json:"Dir"`
	} `json:"Module"`
}

func (s GoCmdSource) Modules(ctx context.Context) ([]catalog.ModuleInput, error) {
	out, err := s.runner.Run(ctx, s.root, "go", "list", "-json", "./...")
	if err != nil {
		return nil, fmt.Errorf("go list: %w", err)
	}
	dec := json.NewDecoder(strings.NewReader(out))
	modPkgs := make(map[string][]catalog.PackageInput)
	modRoots := make(map[string]string)

	for dec.More() {
		var pkg goListPackage
		if err := dec.Decode(&pkg); err != nil {
			return nil, fmt.Errorf("decode go list: %w", err)
		}
		modulePath := ""
		moduleDir := s.root
		if pkg.Module != nil {
			modulePath = pkg.Module.Path
			moduleDir = pkg.Module.Dir
		}
		tests, err := s.collectTests(ctx, pkg.ImportPath)
		if err != nil {
			return nil, fmt.Errorf("go test -list %s: %w", pkg.ImportPath, err)
		}
		modPkgs[modulePath] = append(modPkgs[modulePath], catalog.PackageInput{
			ImportPath: pkg.ImportPath,
			Dir:        pkg.Dir,
			Tests:      tests,
		})
		modRoots[modulePath] = moduleDir
	}

	modules := make([]catalog.ModuleInput, 0, len(modPkgs))
	for path, pkgs := range modPkgs {
		modules = append(modules, catalog.ModuleInput{
			Name:     path,
			RootDir:  modRoots[path],
			Packages: pkgs,
		})
	}
	return modules, nil
}

func (s GoCmdSource) collectTests(ctx context.Context, importPath string) ([]catalog.TestInput, error) {
	out, err := s.runner.Run(ctx, s.root, "go", "test", "-list", ".", importPath)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(out, "\n")
	tests := make([]catalog.TestInput, 0, len(lines))
	for _, ln := range lines {
		name := strings.TrimSpace(ln)
		if name == "" {
			continue
		}
		tests = append(tests, catalog.TestInput{
			Name: name,
			Kind: classifyTest(name),
		})
	}
	return tests, nil
}

func classifyTest(name string) catalog.TestKind {
	switch {
	case strings.HasPrefix(name, "Benchmark"):
		return catalog.TestKindBenchmark
	case strings.HasPrefix(name, "Example"):
		return catalog.TestKindExample
	case strings.HasPrefix(name, "Fuzz"):
		return catalog.TestKindFuzz
	default:
		return catalog.TestKindUnit
	}
}
