package domain

import (
	"time"
)

// NOTE: TestEvent is already defined in domain.go
// This file extends the test event functionality with tree structure

// TestEventAction represents the action field of TestEvent
const (
	ActionRun    = "run"    // Test started running
	ActionPause  = "pause"  // Test paused
	ActionCont   = "cont"   // Test continued
	ActionPass   = "pass"   // Test passed
	ActionFail   = "fail"   // Test failed
	ActionSkip   = "skip"   // Test skipped
	ActionBench  = "bench"  // Benchmark printed output
	ActionOutput = "output" // Test printed output
	ActionStart  = "start"  // Test binary started running
)

// TestTree represents a hierarchical structure of test results
type TestTree struct {
	Package  string               // Package name
	Tests    map[string]*TestNode // Top-level tests
	Status   TestStatus           // Overall package status
	Duration time.Duration        // Total duration
	Output   []string             // Package-level output
}

// TestNode represents a single test or subtest in the tree
type TestNode struct {
	Name     string               // Test name (without parent prefix)
	FullName string               // Full test name (including parent path)
	Status   TestStatus           // Test status
	Duration time.Duration        // Test duration
	Output   []string             // Test output lines
	SubTests map[string]*TestNode // Nested subtests
	Parent   *TestNode            // Parent test (nil for top-level)
}

// IsSubTest returns true if this is a subtest (has a parent)
func (n *TestNode) IsSubTest() bool {
	return n.Parent != nil
}

// GetPath returns the full path from root to this node
func (n *TestNode) GetPath() []string {
	if n.Parent == nil {
		return []string{n.Name}
	}
	return append(n.Parent.GetPath(), n.Name)
}
