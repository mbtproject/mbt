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
