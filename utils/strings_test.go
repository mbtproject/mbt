package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSubsequence(t *testing.T) {
	assert.True(t, IsSubsequence("barry", "barry", false))
	assert.True(t, IsSubsequence("barry", "arr", false))
	assert.True(t, IsSubsequence("barry", "ary", false))
	assert.True(t, IsSubsequence("barray allen", "aal", false))
	assert.True(t, IsSubsequence("abc", "", false))
	assert.True(t, IsSubsequence("", "", false))

	assert.False(t, IsSubsequence("barray", "yr", false))
	assert.False(t, IsSubsequence("barry", "barray a", false))
	assert.False(t, IsSubsequence("barry", "bR", false))
	assert.False(t, IsSubsequence("", "abc", false))

	assert.True(t, IsSubsequence("barry", "barry", true))
	assert.True(t, IsSubsequence("barry", "arr", true))
	assert.True(t, IsSubsequence("barry", "ary", true))
	assert.True(t, IsSubsequence("barray allen", "aal", true))
	assert.True(t, IsSubsequence("abc", "", true))
	assert.True(t, IsSubsequence("", "", true))
	assert.True(t, IsSubsequence("barry", "bR", true))

	assert.False(t, IsSubsequence("barray", "yr", true))
	assert.False(t, IsSubsequence("barry", "barray a", true))
	assert.False(t, IsSubsequence("", "abc", true))
}
