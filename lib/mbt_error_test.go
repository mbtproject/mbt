package lib

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleError(t *testing.T) {
	e := newError("a", "b")
	assert.EqualError(t, e, "mbt a: b")
}

func TestSimpleErrorWithFormat(t *testing.T) {
	e := newErrorf("a", "b%s", "c")
	assert.EqualError(t, e, "mbt a: bc")
}

func TestWrappedErrorWithMessage(t *testing.T) {
	e := wrapm("a", errors.New("c"), "b")
	assert.EqualError(t, e, "mbt a: b - c")
}

func TestWrappedError(t *testing.T) {
	e := wrap("a", errors.New("c"))
	assert.EqualError(t, e, "mbt a: c")
}
