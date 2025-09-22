package tui

import (
	"context"
	"runtime"
	"sync"

	"github.com/YuminosukeSato/lazygotest/internal/adapter/secondary/runner"
	"github.com/YuminosukeSato/lazygotest/internal/domain"
	"github.com/YuminosukeSato/lazygotest/internal/usecase"
	"github.com/YuminosukeSato/lazygotest/pkg/logger"
	"golang.org/x/sync/errgroup"
)

// ParallelRunner manages concurrent test execution
type ParallelRunner struct {
	maxWorkers int
	runTestsUC *usecase.RunTestsUseCase
	mutex      sync.Mutex
	results    map[string]*domain.TestSummary // package ID -> summary
}

// NewParallelRunner creates a new parallel test runner
func NewParallelRunner(runTestsUC *usecase.RunTestsUseCase) *ParallelRunner {
	// Use number of CPU cores for max workers, minimum 2
	maxWorkers := runtime.NumCPU()
	if maxWorkers < 2 {
		maxWorkers = 2
	}

	return &ParallelRunner{
		maxWorkers: maxWorkers,
		runTestsUC: runTestsUC,
		results:    make(map[string]*domain.TestSummary),
	}
}

// RunPackagesParallel runs tests for multiple packages in parallel
func (pr *ParallelRunner) RunPackagesParallel(ctx context.Context, packages []*domain.Package, race bool, cover bool) error {
	// Clear previous results
	pr.mutex.Lock()
	pr.results = make(map[string]*domain.TestSummary)
	pr.mutex.Unlock()

	// Create errgroup with limited concurrency
	g, ctx := errgroup.WithContext(ctx)
	
	// Create a semaphore channel to limit concurrent workers
	sem := make(chan struct{}, pr.maxWorkers)
	
	for _, pkg := range packages {
		// Capture package in closure
		pkg := pkg
		
		g.Go(func() error {
			// Acquire semaphore
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }() // Release semaphore
			case <-ctx.Done():
				return ctx.Err()
			}
			
			logger.Debug("Starting parallel test for package", "package", pkg.Name)
			
			// Run tests for this package using RunOptions
			opts := runner.RunOptions{
				Packages: []string{string(pkg.ID)},
				Verbose:  true,
				Race:     race,
				Cover:    cover,
			}
			
			err := pr.runTestsUC.ExecuteWithOptions(ctx, opts)
			var summary *domain.TestSummary
			if err != nil {
				logger.Error("Failed to run tests for package", "package", pkg.Name, "error", err)
				// Don't return error, continue with other packages
			}
			
			// Store result
			if summary != nil {
				pr.mutex.Lock()
				pr.results[string(pkg.ID)] = summary
				pr.mutex.Unlock()
			}
			
			logger.Debug("Completed parallel test for package", "package", pkg.Name)
			return nil
		})
	}
	
	// Wait for all goroutines to complete
	err := g.Wait()
	if err != nil {
		logger.Error("Error during parallel execution", "error", err)
		return err
	}
	
	return nil
}

// GetResults returns the test results for all packages
func (pr *ParallelRunner) GetResults() map[string]*domain.TestSummary {
	pr.mutex.Lock()
	defer pr.mutex.Unlock()
	
	// Return a copy to avoid race conditions
	results := make(map[string]*domain.TestSummary)
	for k, v := range pr.results {
		results[k] = v
	}
	return results
}

// GetTotalSummary aggregates all test results into a single summary
func (pr *ParallelRunner) GetTotalSummary() *domain.TestSummary {
	pr.mutex.Lock()
	defer pr.mutex.Unlock()
	
	total := &domain.TestSummary{
		Total:   0,
		Passed:  0,
		Failed:  0,
		Skipped: 0,
	}
	
	for _, summary := range pr.results {
		total.Total += summary.Total
		total.Passed += summary.Passed
		total.Failed += summary.Failed
		total.Skipped += summary.Skipped
		total.Duration += summary.Duration
	}
	
	return total
}

// GetMaxWorkers returns the maximum number of concurrent workers
func (pr *ParallelRunner) GetMaxWorkers() int {
	return pr.maxWorkers
}

// SetMaxWorkers updates the maximum number of concurrent workers
func (pr *ParallelRunner) SetMaxWorkers(max int) {
	if max < 1 {
		max = 1
	}
	pr.maxWorkers = max
}