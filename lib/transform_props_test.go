package lib

import (
	"testing"

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
	assert.Equal(t, ErrClassInternal, (err.(*MbtError)).Class())
	assert.Nil(t, o)
}
