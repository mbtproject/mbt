package graph

import (
	"container/list"
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

func makeGraph(items ...interface{}) *list.List {
	g := new(list.List)
	for _, i := range items {
		g.PushBack(i)
	}
	return g
}

func TestNoDependency(t *testing.T) {
	n := newNode("a")

	s, _ := TopSort(makeGraph(n), &testNodeProvider{})

	assert.Equal(t, 1, s.Len())
	assert.Equal(t, n, s.Front().Value)
}

func TestSingleDependency(t *testing.T) {
	a := newNode("a")
	b := newNode("b")
	a.children = []*node{b}

	s, _ := TopSort(makeGraph(a, b), &testNodeProvider{})

	assert.Equal(t, 2, s.Len())
	assert.Equal(t, b, s.Front().Value)
	assert.Equal(t, a, s.Front().Next().Value)
}

func TestDiamondDependency(t *testing.T) {
	a := newNode("a")
	b := newNode("b")
	c := newNode("c")
	d := newNode("d")

	b.children = []*node{d}
	c.children = []*node{d}
	a.children = []*node{b, c}

	s, _ := TopSort(makeGraph(a, b, c, d), &testNodeProvider{})

	assert.Equal(t, 4, s.Len())
	assert.Equal(t, d, s.Front().Value)
	assert.Equal(t, c, s.Front().Next().Value)
	assert.Equal(t, b, s.Front().Next().Next().Value)
	assert.Equal(t, a, s.Front().Next().Next().Next().Value)
}

func TestDirectCircularDependency(t *testing.T) {
	a := newNode("a")
	a.children = []*node{a}

	s, err := TopSort(makeGraph(a), &testNodeProvider{})

	assert.EqualError(t, err, "not a dag")
	assert.Nil(t, s)
}

func TestIndirectCircularDependency(t *testing.T) {
	a := newNode("a")
	b := newNode("b")
	b.children = []*node{a}
	a.children = []*node{b}

	s, err := TopSort(makeGraph(a, b), &testNodeProvider{})

	assert.EqualError(t, err, "not a dag")
	assert.Nil(t, s)
}

func TestCommonLinksFromDisjointNodes(t *testing.T) {
	a := newNode("a")
	b := newNode("b")
	c := newNode("c")
	a.children = []*node{c}
	b.children = []*node{c}

	s, _ := TopSort(makeGraph(a, b), &testNodeProvider{})

	assert.Equal(t, 3, s.Len())
	assert.Equal(t, c, s.Front().Value)
	assert.Equal(t, a, s.Front().Next().Value)
	assert.Equal(t, b, s.Front().Next().Next().Value)
}

func TestIdentity(t *testing.T) {
	a1 := newNode("a")
	a2 := newNode("a")

	s, _ := TopSort(makeGraph(a1, a2), &testNodeProvider{})

	assert.Equal(t, 1, s.Len())
	assert.Equal(t, a1, s.Front().Value)
}

func TestNilInput(t *testing.T) {
	_, err := TopSort(nil, nil)
	assert.EqualError(t, err, "graph should be a valid reference")
	_, err = TopSort(makeGraph(), nil)
	assert.EqualError(t, err, "nodeProvider should be a valid reference")
}

func TestGetChildrenError(t *testing.T) {
	a := newNode("a")
	b := newNode("b")
	a.children = []*node{b}
	_, err := TopSort(makeGraph(a), &testNodeProvider{childError: errors.New("foo")})
	assert.EqualError(t, err, "foo")
}

func TestGetVertices(t *testing.T) {
	a := newNode("a")
	b := newNode("b")
	a.children = []*node{b}

	s, _ := GetVertices(makeGraph(a, b), &testNodeProvider{})
	assert.Equal(t, b, s.Front().Value)
	assert.Equal(t, a, s.Front().Next().Value)
}
