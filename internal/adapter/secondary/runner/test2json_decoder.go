package runner

import (
	"bufio"
	"context"
	"encoding/json"
	"io"

	"lazygotest/internal/domain"
	"lazygotest/pkg/errors"
	"lazygotest/pkg/logger"
)

// Test2JsonDecoder decodes test2json output
type Test2JsonDecoder struct {
	reader *bufio.Scanner
}

// NewTest2JsonDecoder creates a new decoder for test2json output
func NewTest2JsonDecoder(r io.Reader) *Test2JsonDecoder {
	return &Test2JsonDecoder{
		reader: bufio.NewScanner(r),
	}
}

// Decode reads and decodes test events from the input stream
func (d *Test2JsonDecoder) Decode(ctx context.Context) (<-chan domain.TestEvent, <-chan error) {
	events := make(chan domain.TestEvent, 100)
	errs := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errs)

		lineNum := 0
		for d.reader.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}

			lineNum++
			line := d.reader.Bytes()

			var event domain.TestEvent
			if err := json.Unmarshal(line, &event); err != nil {
				logger.Warn("Failed to decode JSON line", "line", lineNum, "content", string(line))
				// Send raw output as a fallback
				events <- domain.TestEvent{
					Action: "output",
					Output: string(line) + "\n",
				}
				continue
			}

			logger.Debug("Decoded event",
				"action", event.Action, "package", event.Package, "test", event.Test)

			select {
			case events <- event:
			case <-ctx.Done():
				return
			}
		}

		if err := d.reader.Err(); err != nil {
			errs <- errors.Wrap(err, "failed to decode test2json stream")
		}
	}()

	return events, errs
}

// ParseTestName extracts the test name from various test event formats
func ParseTestName(event domain.TestEvent) string {
	if event.Test != "" {
		return event.Test
	}
	// For package-level events
	return ""
}

// IsTestEvent checks if the event is related to a specific test
func IsTestEvent(event domain.TestEvent) bool {
	return event.Test != ""
}

// IsPackageEvent checks if the event is package-level
func IsPackageEvent(event domain.TestEvent) bool {
	return event.Test == "" && event.Package != ""
}
