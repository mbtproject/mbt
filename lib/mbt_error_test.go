package lib

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleError(t *testing.T) {
	e := newError("a")
	assert.EqualError(t, e, "mbt: a")
}

func TestSimpleErrorWithFormat(t *testing.T) {
	e := newErrorf("a%s", "b")
	assert.EqualError(t, e, "mbt: ab")
}

func TestWrappedErrorWithMessage(t *testing.T) {
	e := wrapm(errors.New("a"), "b")
	assert.EqualError(t, e, "mbt: b - a")
}

func TestWrappedError(t *testing.T) {
	e := wrap(errors.New("a"))
	assert.EqualError(t, e, "mbt: a")
}

func TestLine(t *testing.T) {
	e := newError("a")
	assert.Equal(t, 31, e.Line())
}

func TestFile(t *testing.T) {
	e := newError("a")
	assert.Equal(t, "mbt_error_test.go", e.File())
}

func TestLocation(t *testing.T) {
	e := newError("a").WithLocation()
	assert.EqualError(t, e, "mbt (mbt_error_test.go,41): a")
}
