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

package lib

import (
	"testing"

	"github.com/mbtproject/mbt/e"
	"github.com/stretchr/testify/assert"
)

func TestFlatMap(t *testing.T) {
	i := make(map[string]interface{})
	i["a"] = 10
	i["b"] = "foo"

	o, err := transformProps(i)
	check(t, err)

	assert.Equal(t, 10, o["a"])
	assert.Equal(t, "foo", o["b"])
}

func TestNestedMap(t *testing.T) {
	i := make(map[string]interface{})
	n := make(map[interface{}]interface{})
	n["a"] = "foo"
	i["a"] = n

	o, err := transformProps(i)
	check(t, err)

	assert.Equal(t, "foo", o["a"].(map[string]interface{})["a"])
}

func TestNonStringKey(t *testing.T) {
	i := make(map[string]interface{})
	n := make(map[interface{}]interface{})
	n[42] = "foo"
	i["a"] = n

	o, err := transformProps(i)

	assert.EqualError(t, err, "Key is not a string 42")
	assert.Equal(t, ErrClassInternal, (err.(*e.E)).Class())
	assert.Nil(t, o)
}
