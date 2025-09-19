package usecase

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/YuminosukeSato/lazygotest/internal/adapter/secondary/runner"
	"github.com/YuminosukeSato/lazygotest/internal/domain"
	"github.com/YuminosukeSato/lazygotest/internal/shared/eventbus"
	"github.com/YuminosukeSato/lazygotest/pkg/logger"
)

// RunTestsUseCase handles test execution
type RunTestsUseCase struct {
	runner    TestRunner
	publisher EventPublisher
}

// NewRunTestsUseCase creates a new RunTestsUseCase
func NewRunTestsUseCase(runner TestRunner, publisher EventPublisher) *RunTestsUseCase {
	return &RunTestsUseCase{
		runner:    runner,
		publisher: publisher,
	}
}

// ExecutePackage runs all tests in a package
func (uc *RunTestsUseCase) ExecutePackage(ctx context.Context, pkgID domain.PkgID) error {
	opts := runner.RunOptions{
		Packages: []string{string(pkgID)},
		Verbose:  true,
	}

	return uc.execute(ctx, opts)
}

// ExecuteTest runs a specific test
func (uc *RunTestsUseCase) ExecuteTest(ctx context.Context, testID domain.TestID) error {
	opts := runner.RunOptions{
		Packages: []string{string(testID.Pkg)},
		RunRegex: "^" + testID.Name + "$",
		Verbose:  true,
	}

	return uc.execute(ctx, opts)
}

// ExecuteAll runs all tests
func (uc *RunTestsUseCase) ExecuteAll(ctx context.Context) error {
	opts := runner.RunOptions{
		Packages: []string{"./..."},
		Verbose:  true,
	}

	return uc.execute(ctx, opts)
}

// ExecuteMultipleTests runs multiple specific tests
func (uc *RunTestsUseCase) ExecuteMultipleTests(ctx context.Context, testIDs []domain.TestID) error {
	if len(testIDs) == 0 {
		return nil
	}

	// Group tests by package
	testsByPackage := make(map[domain.PkgID][]string)
	for _, testID := range testIDs {
		// Escape special regex characters in test names
		escapedName := regexp.QuoteMeta(testID.Name)
		testsByPackage[domain.PkgID(testID.Pkg)] = append(testsByPackage[domain.PkgID(testID.Pkg)], escapedName)
	}

	// For now, if tests are from same package, run them together
	// TODO: Support running tests from multiple packages
	for pkgID, testNames := range testsByPackage {
		// Create regex pattern with OR condition
		runRegex := "^(" + strings.Join(testNames, "|") + ")$"

		opts := runner.RunOptions{
			Packages: []string{string(pkgID)},
			RunRegex: runRegex,
			Verbose:  true,
		}

		// Execute the tests
		if err := uc.execute(ctx, opts); err != nil {
			return err
		}
	}

	return nil
}

// ExecuteWithOptions runs tests with custom options
func (uc *RunTestsUseCase) ExecuteWithOptions(ctx context.Context, opts runner.RunOptions) error {
	return uc.execute(ctx, opts)
}

func (uc *RunTestsUseCase) execute(ctx context.Context, opts runner.RunOptions) error {
	logger.Info("Running tests", "options", opts)

	// Publish test started event
	uc.publisher.Publish(ctx, eventbus.TopicTestStarted, &TestStartedEvent{
		StartedAt: time.Now(),
		Options:   opts,
	})

	// Run tests and stream events
	events, errs := uc.runner.Run(ctx, opts)

	// Process events
	go uc.processEvents(ctx, events, errs)

	return nil
}

func (uc *RunTestsUseCase) processEvents(ctx context.Context, events <-chan domain.TestEvent, errs <-chan error) {
	summary := &domain.TestSummary{
		StartedAt: time.Now(),
	}

	for {
		select {
		case event, ok := <-events:
			if !ok {
				// Stream closed, tests completed
				summary.CompletedAt = time.Now()
				summary.Duration = summary.CompletedAt.Sub(summary.StartedAt)
				uc.publisher.Publish(ctx, eventbus.TopicTestCompleted, summary)
				return
			}

			// Publish each test event
			uc.publisher.Publish(ctx, eventbus.TopicTestEvent, event)

			// Update summary based on event
			uc.updateSummary(summary, event)

			// Publish failure events
			if event.Action == "fail" {
				uc.publisher.PublishAsync(ctx, eventbus.TopicTestFailed, event)
			}

		case err := <-errs:
			if err != nil {
				logger.Error("Test execution error", "error", err)
				uc.publisher.Publish(ctx, eventbus.TopicError, err)
			}

		case <-ctx.Done():
			logger.Debug("Test execution cancelled")
			return
		}
	}
}

func (uc *RunTestsUseCase) updateSummary(summary *domain.TestSummary, event domain.TestEvent) {
	switch event.Action {
	case "pass":
		if event.Test != "" {
			summary.Passed++
			summary.TotalTests++
		} else {
			summary.TotalPackages++
		}
	case "fail":
		if event.Test != "" {
			summary.Failed++
			summary.TotalTests++
		}
	case "skip":
		if event.Test != "" {
			summary.Skipped++
			summary.TotalTests++
		}
	}
}

// TestStartedEvent is published when tests start
type TestStartedEvent struct {
	StartedAt time.Time
	Options   runner.RunOptions
}
