package presentation

import (
	"fmt"

	"github.com/s21066/lazygotest/internal/domain/catalog"
	"github.com/s21066/lazygotest/internal/domain/execution"
)

type PackageNode struct {
	Label string
	Path  string
}

type TestRow struct {
	Name string
	Kind catalog.TestKind
}

type RunRow struct {
	Label    string
	Status   execution.Status
	Duration string
}

type ViewModel struct {
	Packages []PackageNode
	Tests    []TestRow
	Runs     []RunRow
}

// Project builds a UI-friendly snapshot from catalog and run history.
func Project(snapshot catalog.Snapshot, runs []execution.Run) ViewModel {
	pkgNodes := make([]PackageNode, 0, len(snapshot.Modules()))
	testRows := make([]TestRow, 0, 32)
	for _, mod := range snapshot.Modules() {
		for _, pkg := range mod.Packages() {
			label := fmt.Sprintf("%s (%d tests)", pkg.ImportPath(), len(pkg.Tests()))
			pkgNodes = append(pkgNodes, PackageNode{Label: label, Path: pkg.ImportPath()})
			for _, tc := range pkg.Tests() {
				testRows = append(testRows, TestRow{Name: tc.Name, Kind: tc.Kind})
			}
		}
	}

	runRows := make([]RunRow, 0, len(runs))
	for _, run := range runs {
		label := fmt.Sprintf("%s %s", run.PackagePath(), run.Target())
		runRows = append(runRows, RunRow{
			Label:    label,
			Status:   run.Status(),
			Duration: run.FormattedDuration(),
		})
	}

	return ViewModel{
		Packages: pkgNodes,
		Tests:    testRows,
		Runs:     runRows,
	}
}
