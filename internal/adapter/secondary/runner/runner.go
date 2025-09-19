package runner

import (
	"context"
	"os/exec"
	"strconv"
	"strings"

	"github.com/YuminosukeSato/lazygotest/internal/domain"
	"github.com/YuminosukeSato/lazygotest/pkg/errors"
	"github.com/YuminosukeSato/lazygotest/pkg/logger"
)

// RunOptions configures test execution
type RunOptions struct {
	Packages     []string
	RunRegex     string
	Tags         string
	Race         bool
	Cover        bool
	Verbose      bool
	Parallel     int
	Timeout      string
	CoverProfile string
}

// TestRunner executes go test commands
type TestRunner struct{}

// NewTestRunner creates a new test runner
func NewTestRunner() *TestRunner {
	return &TestRunner{}
}

// Run executes tests with the given options and streams events
func (r *TestRunner) Run(ctx context.Context, opts RunOptions) (<-chan domain.TestEvent, <-chan error) {
	events := make(chan domain.TestEvent, 100)
	errs := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errs)

		args := r.buildArgs(opts)
		logger.Info("Running go test", "args", strings.Join(args, " "))

		cmd := exec.CommandContext(ctx, "go", args...)

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			errs <- errors.Wrap(err, "failed to get pipe")
			return
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			errs <- errors.Wrap(err, "failed to get pipe")
			return
		}

		if err := cmd.Start(); err != nil {
			errs <- errors.Wrap(err, "failed to start go test command")
			return
		}

		// Decode stdout (test2json output)
		decoder := NewTest2JsonDecoder(stdout)
		decodedEvents, decodeErrs := decoder.Decode(ctx)

		// Also capture stderr for build errors
		stderrDecoder := NewTest2JsonDecoder(stderr)
		stderrEvents, stderrErrs := stderrDecoder.Decode(ctx)

		// Merge streams
		done := make(chan struct{})
		go func() {
			defer close(done)
			for {
				select {
				case event, ok := <-decodedEvents:
					if !ok {
						decodedEvents = nil
						if stderrEvents == nil {
							return
						}
						continue
					}
					select {
					case events <- event:
					case <-ctx.Done():
						return
					}
				case event, ok := <-stderrEvents:
					if !ok {
						stderrEvents = nil
						if decodedEvents == nil {
							return
						}
						continue
					}
					select {
					case events <- event:
					case <-ctx.Done():
						return
					}
				case err := <-decodeErrs:
					if err != nil {
						errs <- err
					}
				case err := <-stderrErrs:
					if err != nil {
						errs <- err
					}
				case <-ctx.Done():
					return
				}
			}
		}()

		<-done

		if err := cmd.Wait(); err != nil {
			// Non-zero exit code is expected for failing tests
			if _, ok := err.(*exec.ExitError); !ok {
				errs <- errors.Wrap(err, "go test command failed unexpectedly")
			}
		}
	}()

	return events, errs
}

// buildArgs constructs the go test command arguments
func (r *TestRunner) buildArgs(opts RunOptions) []string {
	args := []string{"test", "-json"}

	if opts.Verbose {
		args = append(args, "-v")
	}

	if opts.Race {
		args = append(args, "-race")
	}

	if opts.Cover {
		args = append(args, "-cover")
		if opts.CoverProfile != "" {
			args = append(args, "-coverprofile="+opts.CoverProfile)
		}
	}

	if opts.RunRegex != "" {
		args = append(args, "-run", opts.RunRegex)
	}

	if opts.Tags != "" {
		args = append(args, "-tags", opts.Tags)
	}

	if opts.Parallel > 0 {
		args = append(args, "-parallel="+strconv.Itoa(opts.Parallel))
	}

	if opts.Timeout != "" {
		args = append(args, "-timeout", opts.Timeout)
	}

	// Add packages
	if len(opts.Packages) > 0 {
		args = append(args, opts.Packages...)
	} else {
		args = append(args, "./...")
	}

	return args
}

// ListTests runs go test -list to discover test names
func (r *TestRunner) ListTests(ctx context.Context, pkg string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "go", "test", "-list", ".", pkg)
	output, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(err, "failed to list tests")
	}

	lines := strings.Split(string(output), "\n")
	var tests []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Test") || strings.HasPrefix(line, "Benchmark") {
			tests = append(tests, line)
		}
	}

	return tests, nil
}
