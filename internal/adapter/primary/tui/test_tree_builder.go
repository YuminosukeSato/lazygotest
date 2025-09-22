package tui

import (
	"strings"
	"time"

	"github.com/YuminosukeSato/lazygotest/internal/domain"
)

// BuildTestTree builds or updates a TestTree from a TestEvent
func (m *Model) BuildTestTree(event domain.TestEvent) {
	pkgID := event.Package
	if pkgID == "" {
		return
	}

	// Get or create the test tree for this package
	tree, exists := m.testTrees[pkgID]
	if !exists {
		tree = &domain.TestTree{
			Package: event.Package,
			Tests:   make(map[string]*domain.TestNode),
		}
		m.testTrees[pkgID] = tree
	}

	// If this is a test event (not package-level)
	if event.Test != "" {
		// Parse test name to handle subtests
		testParts := strings.Split(event.Test, "/")
		
		var currentNode *domain.TestNode
		var parentNode *domain.TestNode
		
		// Navigate/create the tree structure
		for i, part := range testParts {
			if i == 0 {
				// Top-level test
				currentNode = tree.Tests[part]
				if currentNode == nil {
					currentNode = &domain.TestNode{
						Name:     part,
						FullName: part,
						SubTests: make(map[string]*domain.TestNode),
					}
					tree.Tests[part] = currentNode
				}
			} else {
				// Subtest
				if parentNode != nil && parentNode.SubTests != nil {
					currentNode = parentNode.SubTests[part]
					if currentNode == nil {
						fullName := strings.Join(testParts[:i+1], "/")
						currentNode = &domain.TestNode{
							Name:     part,
							FullName: fullName,
							Parent:   parentNode,
							SubTests: make(map[string]*domain.TestNode),
						}
						parentNode.SubTests[part] = currentNode
					}
				}
			}
			parentNode = currentNode
		}

		// Update the current node based on the event action
		if currentNode != nil {
			switch event.Action {
			case domain.ActionRun:
				currentNode.Status = domain.StatusRunning
			case domain.ActionPass:
				currentNode.Status = domain.StatusPassed
				if event.Elapsed > 0 {
					currentNode.Duration = time.Duration(event.Elapsed * float64(time.Second))
				}
			case domain.ActionFail:
				currentNode.Status = domain.StatusFailed
				if event.Elapsed > 0 {
					currentNode.Duration = time.Duration(event.Elapsed * float64(time.Second))
				}
			case domain.ActionSkip:
				currentNode.Status = domain.StatusSkipped
			case domain.ActionOutput:
				if currentNode.Output == nil {
					currentNode.Output = []string{}
				}
				currentNode.Output = append(currentNode.Output, event.Output)
			}
		}
	} else {
		// Package-level event
		switch event.Action {
		case domain.ActionPass:
			tree.Status = domain.StatusPassed
			if event.Elapsed > 0 {
				tree.Duration = time.Duration(event.Elapsed * float64(time.Second))
			}
		case domain.ActionFail:
			tree.Status = domain.StatusFailed
			if event.Elapsed > 0 {
				tree.Duration = time.Duration(event.Elapsed * float64(time.Second))
			}
		case domain.ActionOutput:
			if tree.Output == nil {
				tree.Output = []string{}
			}
			tree.Output = append(tree.Output, event.Output)
		}
	}
}