package main

import (
	"errors"
	"fmt"
	"slices"
)

// Simple generic directed graph implementation using an adjacence list
type Graph[T comparable] struct {
	adjacencyList map[T][]T
}

func NewGraph[T comparable]() *Graph[T] {
	return &Graph[T]{
		adjacencyList: make(map[T][]T),
	}
}

// Adds an edge between two vertices (that are created if not existent)
func (g *Graph[T]) AddEdge(from, to T) {
	g.adjacencyList[from] = append(g.adjacencyList[from], to)
	// Add "to" vertex to graph as well
	g.AddVertex(to)
}

// Add a vertex to the graph. Returns an error if already exists
func (g *Graph[T]) AddVertex(vertex T) error {
	if _, ok := g.adjacencyList[vertex]; !ok {
		g.adjacencyList[vertex] = make([]T, 0)
		return nil
	} else {
		return errors.New("vertex already exists")
	}
}

// Returns all vertices in the graph
func (g *Graph[T]) GetAllVertices() []T {
	keys := make([]T, len(g.adjacencyList))
	i := 0
	for k := range g.adjacencyList {
		keys[i] = k
		i++
	}
	return keys
}

// Returns the successors of a given vertex
func (g *Graph[T]) GetSuccessors(vertex T) []T {
	return g.adjacencyList[vertex]
}

// Returns the predecessors of a given vertex
func (g *Graph[T]) GetPredecessors(vertex T) []T {
	predecessors := make([]T, 0)
	for k := range g.adjacencyList {
		if slices.Contains(g.adjacencyList[k], vertex) {
			predecessors = append(predecessors, k)
		}
	}
	return predecessors
}

// Prints the graph in random order
func (g *Graph[T]) Print() {
	for vertex, neighbors := range g.adjacencyList {
		fmt.Printf("%v -> %v\n", vertex, neighbors)
	}
}

// Returns root vertices, i.e. vertices without any incoming edges
func (g *Graph[T]) GetRootNodes() []T {
	rootnodes := make([]T, 0)
	for vertex := range g.adjacencyList {
		predecessors := g.GetPredecessors(vertex)
		if len(predecessors) == 0 {
			rootnodes = append(rootnodes, vertex)
		}
	}
	return rootnodes
}
