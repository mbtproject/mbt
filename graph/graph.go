/*
Copyright 2018 MBT Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package graph

import "errors"

// NodeProvider is the interface between the vertices stored in the graph
// and various graph functions.
// This interface enables the consumers of graph functions to adopt their
// data structures for graph related operations without converting to
// a strict format beforehand.
type NodeProvider interface {
	// ID returns an identifier that can be used to uniquely identify
	// the vertex. This identifier is used internally to determine if
	// two nodes are same.
	ID(vertex interface{}) interface{}

	// ChildCount returns the number of children this vertex has.
	ChildCount(vertex interface{}) int

	// Child returns the child vertex at index in vertex.
	Child(vertex interface{}, index int) (interface{}, error)
}

// CycleError occurs when a cyclic reference is detected in a directed
// acyclic graph.
type CycleError struct {
	Path []interface{}
}

func (e *CycleError) Error() string {
	return "not a dag"
}

// TopSort performs a topological sort of the provided graph.
// Returns an array containing the sorted graph or an
// error if the provided graph is not a directed acyclic graph (DAG).
func TopSort(nodeProvider NodeProvider, graph ...interface{}) ([]interface{}, error) {
	if nodeProvider == nil {
		return nil, errors.New("nodeProvider should be a valid reference")
	}

	traversalState := make(map[interface{}]tState)
	results := make([]interface{}, 0)

	for _, node := range graph {
		nodes, err := newDepthFirst(nodeProvider, node, traversalState).Run()
		if err != nil {
			return nil, err
		}
		results = append(results, nodes...)
	}

	return results, nil
}

// GetVertices returns the list of vetices found in the input graph.
// Input graph should be represented as adjacency list.
func GetVertices(nodeProvider NodeProvider, graph ...interface{}) ([]interface{}, error) {
	return TopSort(nodeProvider, graph...)
}
