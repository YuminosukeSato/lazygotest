package domain

import (
	"time"
)

// PkgID represents a unique identifier for a package
type PkgID string

// Test represents a single test case
type Test struct {
	Package  string
	Name     string
	Duration time.Duration
	Status   TestStatus
	Output   string
}

// TestCase represents a test case with additional metadata
type TestCase struct {
	ID       TestID
	Package  string
	Name     string
	Duration time.Duration
	Status   TestStatus
	Output   string
	Logs     []string
	LastFail *FailInfo
}

// TestStatus represents the status of a test
type TestStatus int

const (
	TestStatusPending TestStatus = iota
	TestStatusRunning
	TestStatusPassed
	TestStatusFailed
	TestStatusSkipped
)

// Legacy status constants for compatibility
const (
	StatusRunning  = TestStatusRunning
	StatusPassed   = TestStatusPassed
	StatusFailed   = TestStatusFailed
	StatusSkipped  = TestStatusSkipped
)

// FailInfo represents failure information for a test
type FailInfo struct {
	FullLog string
	Error   string
}

// Package represents a Go package
type Package struct {
	ID    PkgID
	Path  string
	Name  string
	Tests []TestCase
}

// TestResult represents the result of running a test
type TestResult struct {
	Test     Test
	Output   string
	Duration time.Duration
	Error    error
}

// TestEvent represents an event during test execution
type TestEvent struct {
	Time    time.Time
	Action  string
	Package string
	Test    string
	Output  string
	Elapsed float64
}

// TestID represents a unique identifier for a test
type TestID struct {
	Pkg  string
	Name string
}

// TestSummary represents a summary of test results
type TestSummary struct {
	Total         int
	Passed        int
	Failed        int
	Skipped       int
	TotalTests    int
	TotalPackages int
	StartedAt     time.Time
	CompletedAt   time.Time
	Duration      time.Duration
}