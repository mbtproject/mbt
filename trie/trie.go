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
	// For complete matches, we also return the value stored in that node.
	Value interface{}
}

type node struct {
	prefix          string
	isCompleteMatch bool
	children        map[rune]*node
	value           interface{}
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
func (t *Trie) Add(key string, value interface{}) *Trie {
	addOne(t.root, []rune(key), value)
	return t
}

// Match searches the specified key in Trie.
func (t *Trie) Match(key string) *Match {
	return findCore([]rune(key), t.root)
}

// ContainsPrefix returns true if trie contains the specified key
// with or without a value.
func (t *Trie) ContainsPrefix(key string) bool {
	m := t.Match(key)
	return m.Success
}

// ContainsProperPrefix returns true if specified key is a
// proper prefix of an entry in the trie. A proper prefix of a key is
// a string that is not an exact match. For example if you store
// key "aba", both "a" and "ab" are proper prefixes but "aba" is not.
// In this implementation, empty string is not considered as a
// proper prefix unless a call to .Add is made with an empty string.
func (t *Trie) ContainsProperPrefix(key string) bool {
	m := t.Match(key)
	return m.Success && !m.IsCompleteMatch
}

// Find returns the value specified with a given key.
// If the key is not found or a partial match, it returns
// an additional boolean value indicating the failure.
// Empty string will always return true but value may be null
// unless a call to .Add is made with an empty string and a non nil value.
func (t *Trie) Find(key string) (interface{}, bool) {
	m := t.Match(key)
	return m.Value, m.IsCompleteMatch
}

func addOne(to *node, from []rune, value interface{}) {
	if len(from) == 0 {
		to.isCompleteMatch = true
		to.value = value
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

	addOne(next, from[1:], value)
}

func findCore(target []rune, in *node) *Match {
	if len(target) == 0 {
		return &Match{
			Success:         true,
			IsCompleteMatch: in.isCompleteMatch,
			NearestPrefix:   in.prefix,
			Value:           in.value,
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
