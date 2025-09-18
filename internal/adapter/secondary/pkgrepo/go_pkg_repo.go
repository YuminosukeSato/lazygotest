package pkgrepo

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"

	"lazygotest/internal/domain"
	"lazygotest/pkg/errors"
	"lazygotest/pkg/logger"
)

// GoPackageInfo represents the JSON output from go list
type GoPackageInfo struct {
	Dir          string   `json:"Dir"`
	ImportPath   string   `json:"ImportPath"`
	Name         string   `json:"Name"`
	Target       string   `json:"Target"`
	GoFiles      []string `json:"GoFiles"`
	TestGoFiles  []string `json:"TestGoFiles"`
	XTestGoFiles []string `json:"XTestGoFiles"`
}

// GoPackageRepo discovers Go packages using go list
type GoPackageRepo struct{}

// NewGoPackageRepo creates a new package repository
func NewGoPackageRepo() *GoPackageRepo {
	return &GoPackageRepo{}
}

// ListPackages discovers all packages with tests in the module
func (r *GoPackageRepo) ListPackages(ctx context.Context) ([]*domain.Package, error) {
	logger.Debug("Discovering packages with tests")

	cmd := exec.CommandContext(ctx, "go", "list", "-json", "./...")
	output, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute go list")
	}

	packages := make([]*domain.Package, 0)
	decoder := json.NewDecoder(strings.NewReader(string(output)))

	for decoder.More() {
		var pkgInfo GoPackageInfo
		if err := decoder.Decode(&pkgInfo); err != nil {
			logger.Warn("Failed to decode package info", "error", err)
			continue
		}

		// Only include packages with test files
		if !r.hasTests(pkgInfo) {
			continue
		}

		pkg := &domain.Package{
			ID:    domain.PkgID(pkgInfo.ImportPath),
			Path:  pkgInfo.Dir,
			Name:  pkgInfo.Name,
			Tests: []domain.TestCase{},
		}

		packages = append(packages, pkg)
		logger.Debug("Found package with tests", "package", pkgInfo.ImportPath)
	}

	logger.Info("Discovered packages with tests", "count", len(packages))
	return packages, nil
}

// GetPackage retrieves information about a specific package
func (r *GoPackageRepo) GetPackage(ctx context.Context, pkgPath string) (*domain.Package, error) {
	cmd := exec.CommandContext(ctx, "go", "list", "-json", pkgPath)
	output, err := cmd.Output()
	if err != nil {
		return nil, errors.NotFound(pkgPath)
	}

	var pkgInfo GoPackageInfo
	if err := json.Unmarshal(output, &pkgInfo); err != nil {
		return nil, errors.Wrap(err, "failed to parse package JSON")
	}

	return &domain.Package{
		ID:    domain.PkgID(pkgInfo.ImportPath),
		Path:  pkgInfo.Dir,
		Name:  pkgInfo.Name,
		Tests: []domain.TestCase{},
	}, nil
}

// hasTests checks if a package has test files
func (r *GoPackageRepo) hasTests(pkg GoPackageInfo) bool {
	return len(pkg.TestGoFiles) > 0 || len(pkg.XTestGoFiles) > 0
}

// FilterPackages filters packages based on a pattern
func (r *GoPackageRepo) FilterPackages(packages []*domain.Package, pattern string) []*domain.Package {
	if pattern == "" {
		return packages
	}

	pattern = strings.ToLower(pattern)
	filtered := make([]*domain.Package, 0)

	for _, pkg := range packages {
		if strings.Contains(strings.ToLower(string(pkg.ID)), pattern) ||
			strings.Contains(strings.ToLower(pkg.Name), pattern) {
			filtered = append(filtered, pkg)
		}
	}

	return filtered
}
