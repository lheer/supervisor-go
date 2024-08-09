package main

import (
	"slices"
	"testing"
)

func TestGraph(t *testing.T) {
	// Test creation
	graph := NewGraph[int]()
	graph.AddEdge(1, 2)
	graph.AddEdge(1, 3)
	graph.AddEdge(2, 4)
	err := graph.AddVertex(5)
	if err != nil {
		t.Errorf("Got error while adding vertex: %s", err)
	}
	err = graph.AddVertex(5)
	if err == nil {
		t.Errorf("Expected error while adding existing vertex, got nil")
	}

	// Test successors
	successors := graph.GetSuccessors(1)
	if len(successors) != 2 {
		t.Errorf("Expected one successor, got %v", successors)
	}
	if !slices.Contains(successors, 2) || !slices.Contains(successors, 3) {
		t.Errorf("Got wrong successors for vertex 1: %v", successors)
	}

	successors = graph.GetSuccessors(5)
	if len(successors) != 0 {
		t.Errorf("Expected 0 successors for 5, got %v", successors)
	}

	// Test predecessors
	predecessors := graph.GetPredecessors(2)
	if !slices.Contains(predecessors, 1) || len(predecessors) != 1 {
		t.Errorf("Expected predecessors of 2 to be 1, got %v", predecessors)
	}

	// Test rootnodes
	rootnodes := graph.GetRootNodes()
	if len(rootnodes) != 2 || !slices.Contains(rootnodes, 1) || !slices.Contains(rootnodes, 5) {
		t.Errorf("Expected root nodes of graph to be [1, 5], got %v", rootnodes)
	}
}
