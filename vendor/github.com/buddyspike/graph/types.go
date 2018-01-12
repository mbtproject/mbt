package graph

// NodeProvider is the interface between the vertices stored in the graph
// and various graph functions.
type NodeProvider interface {
	// GetId returns an identifier that can be used to uniquely identity
	// the vertex. That identifier is used internally to determine if
	// two nodes are same.
	ID(vertex interface{}) interface{}

	// ChildCount returns the number of children this vertex has.
	ChildCount(vertex interface{}) int

	// Child returns the child vertex at index in vertex.
	Child(vertex interface{}, index int) (interface{}, error)
}
