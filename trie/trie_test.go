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

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyTrie(t *testing.T) {
	i := NewTrie()

	m := i.Match("abc")

	assert.False(t, m.Success)
	assert.False(t, m.IsCompleteMatch)
	assert.Equal(t, "", m.NearestPrefix)
	assert.Nil(t, m.Value)
}
func TestFindingEmptyString(t *testing.T) {
	i := NewTrie()
	i.Add("abc", 42)

	m := i.Match("")

	assert.True(t, m.Success)
	assert.True(t, m.IsCompleteMatch)
	assert.Equal(t, "", m.NearestPrefix)
	assert.Nil(t, m.Value)
}

func TestAddingEmptyString(t *testing.T) {
	i := NewTrie()
	i.Add("", 42)
	m := i.Match("")

	assert.True(t, m.Success)
	assert.True(t, m.IsCompleteMatch)
	assert.Equal(t, "", m.NearestPrefix)
	assert.Equal(t, 42, m.Value)

	m = i.Match("a")
	assert.False(t, m.Success)
	assert.False(t, m.IsCompleteMatch)
	assert.Equal(t, "", m.NearestPrefix)
	assert.Nil(t, m.Value)
}

func TestSimpleMatch(t *testing.T) {
	i := NewTrie()
	i.Add("a", 42)

	m := i.Match("a")

	assert.True(t, m.Success)
	assert.True(t, m.IsCompleteMatch)
	assert.Equal(t, "a", m.NearestPrefix)
	assert.Equal(t, 42, m.Value)
}

func TestAddingSameKeyTwice(t *testing.T) {
	i := NewTrie()
	i.Add("abc", 42)
	i.Add("abc", 43)

	m := i.Match("abc")

	assert.True(t, m.Success)
	assert.True(t, m.IsCompleteMatch)
	assert.Equal(t, "abc", m.NearestPrefix)
	// Last one should win
	assert.Equal(t, 43, m.Value)
}

func TestPrefix(t *testing.T) {
	i := NewTrie()
	i.Add("aa", 42)
	i.Add("aabb", 43)

	m := i.Match("aa")

	assert.True(t, m.Success)
	assert.True(t, m.IsCompleteMatch)
	assert.Equal(t, "aa", m.NearestPrefix)
	assert.Equal(t, 42, m.Value)

	m = i.Match("aab")

	assert.True(t, m.Success)
	assert.False(t, m.IsCompleteMatch)
	assert.Equal(t, "aab", m.NearestPrefix)
	assert.Nil(t, m.Value)

	m = i.Match("aabb")

	assert.True(t, m.Success)
	assert.True(t, m.IsCompleteMatch)
	assert.Equal(t, "aabb", m.NearestPrefix)
	assert.Equal(t, 43, m.Value)

	m = i.Match("ab")

	assert.False(t, m.Success)
	assert.False(t, m.IsCompleteMatch)
	assert.Equal(t, "a", m.NearestPrefix)
	assert.Nil(t, m.Value)
}

func TestUnsuccessfulMatch(t *testing.T) {
	i := NewTrie()

	i.Add("abc", 42)

	m := i.Match("def")

	assert.False(t, m.Success)
	assert.False(t, m.IsCompleteMatch)
	assert.Equal(t, "", m.NearestPrefix)
	assert.Nil(t, m.Value)
}

func TestUnicodeEntries(t *testing.T) {
	i := NewTrie()
	i.Add("😄🍓💩👠", 42)

	m := i.Match("😄🍓")

	assert.True(t, m.Success)
	assert.False(t, m.IsCompleteMatch)
	assert.Equal(t, "😄🍓", m.NearestPrefix)
	assert.Nil(t, m.Value)

	m = i.Match("😄🍓💩👠")

	assert.True(t, m.Success)
	assert.True(t, m.IsCompleteMatch)
	assert.Equal(t, "😄🍓💩👠", m.NearestPrefix)
	assert.Equal(t, 42, m.Value)

	m = i.Match("🍓💩")

	assert.False(t, m.Success)
	assert.False(t, m.IsCompleteMatch)
	assert.Equal(t, "", m.NearestPrefix)
	assert.Nil(t, m.Value)
}

func TestFind(t *testing.T) {
	i := NewTrie()
	i.Add("aba", 42)
	i.Add("abb", 43)

	v, ok := i.Find("aba")
	assert.True(t, ok)
	assert.Equal(t, 42, v)

	v, ok = i.Find("abb")
	assert.True(t, ok)
	assert.Equal(t, 43, v)

	v, ok = i.Find("ab")
	assert.False(t, ok)
	assert.Nil(t, v)

	v, ok = i.Find("")
	assert.True(t, ok)
	assert.Nil(t, v)

	v, ok = i.Find("def")
	assert.False(t, ok)
	assert.Nil(t, v)

	i.Add("", 42)
	v, ok = i.Find("")
	assert.True(t, ok)
	assert.Equal(t, 42, v)
}

func TestContainsPrefix(t *testing.T) {
	i := NewTrie()
	i.Add("aba", 42)

	assert.True(t, i.ContainsPrefix("aba"))
	assert.True(t, i.ContainsPrefix("ab"))
	assert.True(t, i.ContainsPrefix(""))
	assert.False(t, i.ContainsPrefix("abc"))
	assert.False(t, i.ContainsPrefix("def"))
}

func TestContainsProperPrefix(t *testing.T) {
	i := NewTrie()
	i.Add("aba", 42)

	assert.True(t, i.ContainsProperPrefix("ab"))
	assert.False(t, i.ContainsProperPrefix(""))
	assert.False(t, i.ContainsProperPrefix("aba"))
	assert.False(t, i.ContainsProperPrefix("def"))
	assert.False(t, i.ContainsProperPrefix("ba"))
}
