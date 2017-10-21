package lib

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleError(t *testing.T) {
	e := newError("a", "b", nil)
	assert.EqualError(t, e, "a: b")
}

func TestSimpleErrorWithFormat(t *testing.T) {
	e := newError("a", "b%s", nil, "c")
	assert.EqualError(t, e, "a: bc")
}

func TestInnerError(t *testing.T) {
	e := newError("a", "b", errors.New("c"))
	assert.EqualError(t, e, "a: b - c")
}
