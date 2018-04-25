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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type node struct {
	name     string
	children []*node
}

type testNodeProvider struct {
	childError error
}

func (n *testNodeProvider) ID(vertex interface{}) interface{} {
	return vertex.(*node).name
}

func (n *testNodeProvider) ChildCount(vertex interface{}) int {
	return len(vertex.(*node).children)
}

func (n *testNodeProvider) Child(vertex interface{}, index int) (interface{}, error) {
	if n.childError != nil {
		return nil, n.childError
	}

	return vertex.(*node).children[index], nil
}

func newNode(name string) *node {
	return &node{
		name:     name,
		children: []*node{},
	}
}

func TestNoDependency(t *testing.T) {
	n := newNode("a")

	s, _ := TopSort(&testNodeProvider{}, n)

	assert.Equal(t, 1, len(s))
	assert.Equal(t, n, s[0])
}

func TestSingleDependency(t *testing.T) {
	a := newNode("a")
	b := newNode("b")
	a.children = []*node{b}

	s, _ := TopSort(&testNodeProvider{}, a, b)

	assert.Equal(t, 2, len(s))
	assert.Equal(t, b, s[0])
	assert.Equal(t, a, s[1])
}

func TestDiamondDependency(t *testing.T) {
	a := newNode("a")
	b := newNode("b")
	c := newNode("c")
	d := newNode("d")

	b.children = []*node{d}
	c.children = []*node{d}
	a.children = []*node{b, c}

	s, _ := TopSort(&testNodeProvider{}, a, b, c, d)

	assert.Equal(t, 4, len(s))
	assert.Equal(t, d, s[0])
	assert.Equal(t, c, s[1])
	assert.Equal(t, b, s[2])
	assert.Equal(t, a, s[3])
}

func TestDirectCircularDependency(t *testing.T) {
	a := newNode("a")
	a.children = []*node{a}

	s, err := TopSort(&testNodeProvider{}, a)
	cErr := err.(*CycleError)

	assert.EqualError(t, cErr, "not a dag")
	assert.Equal(t, a, cErr.Path[0])
	assert.Equal(t, a, cErr.Path[1])
	assert.Nil(t, s)
}

func TestIndirectCircularDependency(t *testing.T) {
	a := newNode("a")
	b := newNode("b")
	b.children = []*node{a}
	a.children = []*node{b}

	s, err := TopSort(&testNodeProvider{}, a, b)
	cErr := err.(*CycleError)

	assert.EqualError(t, cErr, "not a dag")
	assert.Len(t, cErr.Path, 3)
	assert.Equal(t, a, cErr.Path[0])
	assert.Equal(t, b, cErr.Path[1])
	assert.Equal(t, a, cErr.Path[2])
	assert.Nil(t, s)
}

func TestIndirectCircularDependencyForDuplicatedRoots(t *testing.T) {
	a := newNode("a")
	b := newNode("b")
	b.children = []*node{a}
	a.children = []*node{b}

	s, err := TopSort(&testNodeProvider{}, a, b, a, b)
	cErr := err.(*CycleError)

	assert.EqualError(t, cErr, "not a dag")
	assert.Len(t, cErr.Path, 3)
	assert.Equal(t, a, cErr.Path[0])
	assert.Equal(t, b, cErr.Path[1])
	assert.Equal(t, a, cErr.Path[2])
	assert.Nil(t, s)
}

func TestCommonLinksFromDisjointNodes(t *testing.T) {
	a := newNode("a")
	b := newNode("b")
	c := newNode("c")
	a.children = []*node{c}
	b.children = []*node{c}

	s, _ := TopSort(&testNodeProvider{}, a, b)

	assert.Equal(t, 3, len(s))
	assert.Equal(t, c, s[0])
	assert.Equal(t, a, s[1])
	assert.Equal(t, b, s[2])
}

func TestIdentity(t *testing.T) {
	a1 := newNode("a")
	a2 := newNode("a")

	s, _ := TopSort(&testNodeProvider{}, a1, a2)

	assert.Equal(t, 1, len(s))
	assert.Equal(t, a1, s[0])
}

func TestNilInput(t *testing.T) {
	_, err := TopSort(nil)
	assert.EqualError(t, err, "nodeProvider should be a valid reference")
}

func TestEmptyInput(t *testing.T) {
	s, err := TopSort(&testNodeProvider{})
	assert.NoError(t, err)
	assert.Len(t, s, 0)
}

func TestGetChildrenError(t *testing.T) {
	a := newNode("a")
	b := newNode("b")
	a.children = []*node{b}
	_, err := TopSort(&testNodeProvider{childError: errors.New("foo")}, a)
	assert.EqualError(t, err, "foo")
}

func TestComplexGraph(t *testing.T) {
	/*
		a -> [b, c, d]
		b -> [c]
		c -> [e]
		d -> [c, g]
		e -> []
		f -> [c]
		g -> [f]
	*/
	a := newNode("a")
	b := newNode("b")
	c := newNode("c")
	d := newNode("d")
	e := newNode("e")
	f := newNode("f")
	g := newNode("g")

	a.children = []*node{b, c, d}
	b.children = []*node{c}
	c.children = []*node{e}
	d.children = []*node{c, g}
	f.children = []*node{c}
	g.children = []*node{f}

	s, err := TopSort(&testNodeProvider{}, a)

	assert.Nil(t, err)
	assert.Equal(t, 7, len(s))
	assert.Equal(t, e, s[0])
	assert.Equal(t, c, s[1])
	assert.Equal(t, f, s[2])
	assert.Equal(t, g, s[3])
	assert.Equal(t, d, s[4])
	assert.Equal(t, b, s[5])
	assert.Equal(t, a, s[6])
}

func TestComplexCycle(t *testing.T) {
	/*
		a -> [b, c, d]
		b -> [c]
		c -> [e]
		d -> [c, g]
		e -> [f]
		f -> [c]
		g -> [f]
	*/
	a := newNode("a")
	b := newNode("b")
	c := newNode("c")
	d := newNode("d")
	e := newNode("e")
	f := newNode("f")
	g := newNode("g")

	a.children = []*node{b, c, d}
	b.children = []*node{c}
	c.children = []*node{e}
	d.children = []*node{c, g}
	f.children = []*node{c}
	g.children = []*node{f}
	e.children = []*node{f}

	s, err := TopSort(&testNodeProvider{}, a)
	cErr := err.(*CycleError)

	assert.Nil(t, s)
	assert.EqualError(t, cErr, "not a dag")
	assert.Len(t, cErr.Path, 4)
	assert.Equal(t, f, cErr.Path[0])
	assert.Equal(t, c, cErr.Path[1])
	assert.Equal(t, e, cErr.Path[2])
	assert.Equal(t, f, cErr.Path[3])
}

func TestGetVertices(t *testing.T) {
	a := newNode("a")
	b := newNode("b")
	a.children = []*node{b}

	s, _ := GetVertices(&testNodeProvider{}, a, b)
	assert.Equal(t, b, s[0])
	assert.Equal(t, a, s[1])
}
