package graph

import (
	"container/list"
	"errors"
)

type tState int

const (
	stateNew = iota
	stateOpen
	stateClosed
)

// TopSort performs a topological sort of the provided graph.
// Returns an array containing the sorted graph or an
// error if the provided graph is not a directed acyclic graph (DAG).
// Input graph should be represented as adjacency list.
func TopSort(graph *list.List, nodeProvider NodeProvider) (*list.List, error) {
	if graph == nil {
		return nil, errors.New("graph should be a valid reference")
	}

	if nodeProvider == nil {
		return nil, errors.New("nodeProvider should be a valid reference")
	}

	state := make(map[interface{}]tState)

	lifo := list.New()
	for e := graph.Front(); e != nil; e = e.Next() {
		lifo.PushBack(e.Value)
	}

	r := new(list.List)

	for lifo.Len() > 0 {
		head := lifo.Front()
		lifo.Remove(head)

		n := head.Value
		key := nodeProvider.ID(n)

		s := state[key]
		if s == stateClosed {
			continue
		}

		if s == stateOpen {
			state[key] = stateClosed
			r.PushBack(n)
			continue
		}

		childCount := nodeProvider.ChildCount(n)

		if childCount == 0 {
			state[key] = stateClosed
			r.PushBack(n)
		} else {
			state[key] = stateOpen
			lifo.PushFront(n)

			for i := 0; i < childCount; i++ {
				c, err := nodeProvider.Child(n, i)
				if err != nil {
					return nil, err
				}
				cid := nodeProvider.ID(c)
				s := state[cid]
				if s == stateOpen {
					return nil, errors.New("not a dag")
				}
				lifo.PushFront(c)
			}
		}
	}

	return r, nil
}

// GetVertices returns the list of vetices found in the input graph.
// Input graph should be represented as adjacency list.
func GetVertices(graph *list.List, nodeProvider NodeProvider) (*list.List, error) {
	return TopSort(graph, nodeProvider)
}
