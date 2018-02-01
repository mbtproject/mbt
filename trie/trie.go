package trie

// Trie is a representation of Prefix Trie data structure.
type Trie struct {
	root *node
}

// Match is a result of looking up a value in a Trie.
type Match struct {
	// Indicates whether this is a successful match or not.
	Success bool
	// Indicates whether the input matches a complete entry or not.
	IsCompleteMatch bool
	// For successful results (i.e. Success = true), this field has
	// the same value as string being searched for.
	// For unsuccessful results (i.e. Success = false), it contains
	// the closest matching prefix.
	NearestPrefix string
}

type node struct {
	prefix          string
	isCompleteMatch bool
	children        map[rune]*node
}

// NewTrie creates a new instance of Trie.
func NewTrie() *Trie {
	return &Trie{
		root: &node{
			prefix:          "",
			isCompleteMatch: true,
			children:        make(map[rune]*node),
		},
	}
}

// Add creates a new entry in Trie.
func (t *Trie) Add(value string) *Trie {
	addOne(t.root, []rune(value))
	return t
}

// Find searches the specified value in Trie.
func (t *Trie) Find(value string) *Match {
	return findCore([]rune(value), t.root)
}

func addOne(to *node, from []rune) {
	if len(from) == 0 {
		to.isCompleteMatch = true
		return
	}

	head := from[0]
	next, ok := to.children[head]
	if !ok {
		next = &node{
			prefix:   to.prefix + string(head),
			children: make(map[rune]*node),
		}
		to.children[head] = next
	}

	addOne(next, from[1:])
}

func findCore(target []rune, in *node) *Match {
	if len(target) == 0 {
		return &Match{
			Success:         true,
			IsCompleteMatch: in.isCompleteMatch,
			NearestPrefix:   in.prefix,
		}
	}

	head := target[0]
	if n, ok := in.children[head]; ok {
		return findCore(target[1:], n)
	}

	return &Match{
		Success:       false,
		NearestPrefix: in.prefix,
	}
}
