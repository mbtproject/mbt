package trie

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyTrie(t *testing.T) {
	i := NewTrie()

	m := i.Find("abc")

	assert.False(t, m.Success)
	assert.False(t, m.IsCompleteMatch)
	assert.Equal(t, "", m.NearestPrefix)
}
func TestFindingEmptyString(t *testing.T) {
	i := NewTrie()
	i.Add("abc")

	m := i.Find("")

	assert.True(t, m.Success)
	assert.True(t, m.IsCompleteMatch)
	assert.Equal(t, "", m.NearestPrefix)
}

func TestSimpleMatch(t *testing.T) {
	i := NewTrie()
	i.Add("a")

	m := i.Find("a")

	assert.True(t, m.Success)
	assert.True(t, m.IsCompleteMatch)
	assert.Equal(t, "a", m.NearestPrefix)
}

func TestAddingSameKeyTwice(t *testing.T) {
	i := NewTrie()
	i.Add("abc")
	i.Add("abc")

	m := i.Find("abc")

	assert.True(t, m.Success)
	assert.True(t, m.IsCompleteMatch)
	assert.Equal(t, "abc", m.NearestPrefix)
}

func TestPrefix(t *testing.T) {
	i := NewTrie()
	i.Add("aa")
	i.Add("aabb")

	m := i.Find("aa")

	assert.True(t, m.Success)
	assert.True(t, m.IsCompleteMatch)
	assert.Equal(t, "aa", m.NearestPrefix)

	m = i.Find("aab")

	assert.True(t, m.Success)
	assert.False(t, m.IsCompleteMatch)
	assert.Equal(t, "aab", m.NearestPrefix)

	m = i.Find("aabb")

	assert.True(t, m.Success)
	assert.True(t, m.IsCompleteMatch)
	assert.Equal(t, "aabb", m.NearestPrefix)

	m = i.Find("ab")

	assert.False(t, m.Success)
	assert.False(t, m.IsCompleteMatch)
	assert.Equal(t, "a", m.NearestPrefix)
}

func TestUnsuccessfulMatch(t *testing.T) {
	i := NewTrie()

	i.Add("abc")

	m := i.Find("def")

	assert.False(t, m.Success)
	assert.False(t, m.IsCompleteMatch)
	assert.Equal(t, "", m.NearestPrefix)
}

func TestUnicodeEntries(t *testing.T) {
	i := NewTrie()
	i.Add("😄🍓💩👠")

	m := i.Find("😄🍓")

	assert.True(t, m.Success)
	assert.False(t, m.IsCompleteMatch)
	assert.Equal(t, "😄🍓", m.NearestPrefix)

	m = i.Find("😄🍓💩👠")

	assert.True(t, m.Success)
	assert.True(t, m.IsCompleteMatch)
	assert.Equal(t, "😄🍓💩👠", m.NearestPrefix)

	m = i.Find("🍓💩")

	assert.False(t, m.Success)
	assert.False(t, m.IsCompleteMatch)
	assert.Equal(t, "", m.NearestPrefix)
}
