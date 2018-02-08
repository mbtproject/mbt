package lib

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleError(t *testing.T) {
	e := newError(ErrClassInternal, "a")
	assert.EqualError(t, e, "a")
}

func TestSimpleErrorWithFormat(t *testing.T) {
	e := newErrorf(ErrClassInternal, "a%s", "b")
	assert.EqualError(t, e, "ab")
}

func TestWrappedErrorWithMessage(t *testing.T) {
	i := errors.New("a")
	e := wrapm(ErrClassInternal, i, "b")
	assert.EqualError(t, e, "b")
	assert.EqualError(t, e.InnerError(), "a")
}

func TestWrappedError(t *testing.T) {
	e := wrap(ErrClassInternal, errors.New("a"))
	assert.EqualError(t, e, "a")
}

func TestStack(t *testing.T) {
	e := wrap(ErrClassInternal, errors.New("b"))
	assert.Equal(t, "github.com/mbtproject/mbt/lib.TestStack", e.Stack()[1].Function)
}

func WrappingAnMbtError(t *testing.T) {
	a := wrap(ErrClassInternal, errors.New("a"))
	assert.Equal(t, a, wrap(ErrClassInternal, a))
	assert.Equal(t, a, wrap(ErrClassUser, a))
}
