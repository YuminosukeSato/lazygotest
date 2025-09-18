package logger

import (
	"log/slog"
	"os"
	"sync"
)

var (
	instance *slog.Logger
	once     sync.Once
	logFile  *os.File
	mu       sync.RWMutex
)

// Init initializes the global logger with the given file path
func Init(filepath string) error {
	mu.Lock()
	defer mu.Unlock()

	var err error
	once.Do(func() {
		if filepath != "" {
			logFile, err = os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				return
			}

			opts := &slog.HandlerOptions{
				Level:     slog.LevelDebug,
				AddSource: true,
			}
			handler := slog.NewTextHandler(logFile, opts)
			instance = slog.New(handler)
		} else {
			// Use default stdout logger if no file specified
			instance = slog.Default()
		}

		slog.SetDefault(instance)
	})

	return err
}

// Close closes the log file if it was opened
func Close() error {
	mu.Lock()
	defer mu.Unlock()

	if logFile != nil {
		err := logFile.Close()
		logFile = nil
		return err
	}
	return nil
}

// Get returns the current logger instance
func Get() *slog.Logger {
	mu.RLock()
	defer mu.RUnlock()

	if instance == nil {
		return slog.Default()
	}
	return instance
}

// Debug logs a debug level message
func Debug(msg string, args ...any) {
	Get().Debug(msg, args...)
}

// Info logs an info level message
func Info(msg string, args ...any) {
	Get().Info(msg, args...)
}

// Warn logs a warning level message
func Warn(msg string, args ...any) {
	Get().Warn(msg, args...)
}

// Error logs an error level message
func Error(msg string, args ...any) {
	Get().Error(msg, args...)
}

// With returns a logger with additional context
func With(args ...any) *slog.Logger {
	return Get().With(args...)
}
