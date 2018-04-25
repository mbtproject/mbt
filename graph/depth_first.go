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

import (
	"container/list"
)

type tState int

const (
	stateNew = iota
	stateOpen
	stateClosed
)

// This is our non-recursive depth-first traversal
// implementation.
type depthFirst struct {
	State        map[interface{}]tState
	Lifo         *list.List
	NodeProvider NodeProvider
}

func (d *depthFirst) Run() ([]interface{}, error) {
	r := make([]interface{}, 0)

	for d.Lifo.Len() > 0 {
		head := d.Lifo.Front()
		d.Lifo.Remove(head)

		n := head.Value
		key := d.NodeProvider.ID(n)

		s := d.State[key]
		if s == stateClosed {
			continue
		}

		if s == stateOpen {
			d.State[key] = stateClosed
			r = append(r, n)
			continue
		}

		childCount := d.NodeProvider.ChildCount(n)

		if childCount == 0 {
			d.State[key] = stateClosed
			r = append(r, n)
		} else {
			d.State[key] = stateOpen
			d.Lifo.PushFront(n)

			for i := 0; i < childCount; i++ {
				c, err := d.NodeProvider.Child(n, i)
				if err != nil {
					return nil, err
				}
				cid := d.NodeProvider.ID(c)
				s := d.State[cid]
				if s == stateOpen {
					path := d.currentPath(cid)
					path = append(path, c)
					return nil, &CycleError{Path: path}
				}
				d.Lifo.PushFront(c)
			}
		}
	}

	return r, nil
}

func (d *depthFirst) currentPath(startID interface{}) []interface{} {
	result := make([]interface{}, 0)
	tail := d.Lifo.Back()

	// Skip all items until the specified start node
	for ; tail != nil; tail = tail.Prev() {
		if d.NodeProvider.ID(tail.Value) == startID {
			break
		}
	}

	for ; tail != nil; tail = tail.Prev() {
		v := tail.Value
		key := d.NodeProvider.ID(v)
		if state, ok := d.State[key]; ok {
			if state == stateOpen {
				result = append(result, v)
			}
		}
	}

	return result
}

func newDepthFirst(nodeProvider NodeProvider, startNode interface{}, state map[interface{}]tState) *depthFirst {
	lifo := list.New()
	lifo.PushBack(startNode)

	return &depthFirst{
		State:        state,
		Lifo:         lifo,
		NodeProvider: nodeProvider,
	}
}
