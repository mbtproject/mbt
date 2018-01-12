# graph

General purpose graph functions.

[![Build Status](https://travis-ci.org/buddyspike/graph.svg?branch=master)](https://travis-ci.org/buddyspike/graph)
[![Go Report Card](https://goreportcard.com/badge/github.com/buddyspike/graph)](https://goreportcard.com/report/github.com/buddyspike/graph)
[![Coverage Status](https://coveralls.io/repos/github/buddyspike/graph/badge.svg?branch=master)](https://coveralls.io/github/buddyspike/graph?branch=master)

## Install
```
go get github.com/buddyspike/graph
```

## API

### TopSort
Topological sorting of a directed acyclic graph.

```go
package main

import (
	"fmt"

	"github.com/buddyspike/topsort"
)

// Example type that represent a vertex in the graph
type node struct {
	name     string
	children []*node
}

// Create a type that implements graph.NodeProvider interface
type nodeProvider struct {}

func (n *nodeProvider) ID(vertex interface{}) interface{} {
	return vertex.(*node).name
}

func (n *nodeProvider) ChildCount(vertex interface{}) int {
	return len(vertex.(*node).children)
}

func (n *nodeProvider) Child(vertex interface{}, index int) (interface{}, error) {
	return vertex.(*node).children[i]
}

func newNode(name string) *node {
	return &node{
		name:     name,
		children: []*node{},
	}
}

func main() {
	// Create our nodes and associate them.
	a := newNode("a")
	b := newNode("b")
	a.children = []*node{b}

	// Create a list with items
	g := new(list.List)
	g.PushBack(a)
	g.PushBack(b) // This is not strictly needed since we have a link from a to b

	// Perform the sort.
	s, err := graph.TopSort(g, &nodeProvider{})

	if err != nil {
		panic(err)
	}

	// Print each item in the returned array.
	// The results should be printed in dependency order.
	for e := s.Front(); e != nil; e = e.Next() {
		println((e.Value.(*node)).name)
	}
	// Output
	// b
	// a
}
```

