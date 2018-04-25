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

package e

import (
	"errors"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	ErrClassInternal = iota
	ErrClassUser
)

func TestSimpleError(t *testing.T) {
	e := NewError(ErrClassInternal, "a")
	assert.EqualError(t, e, "a")
}

func TestSimpleErrorWithFormat(t *testing.T) {
	e := NewErrorf(ErrClassInternal, "a%s", "b")
	assert.EqualError(t, e, "ab")
}

func TestWrappedErrorWithMessage(t *testing.T) {
	i := errors.New("a")
	e := Wrapf(ErrClassInternal, i, "b")
	assert.EqualError(t, e, "b")
	assert.EqualError(t, e.InnerError(), "a")
}

func TestWrappedError(t *testing.T) {
	e := Wrap(ErrClassInternal, errors.New("a"))
	assert.EqualError(t, e, "a")
}

func TestStack(t *testing.T) {
	ptr, _, _, ok := runtime.Caller(0)
	assert.True(t, ok)

	frames := runtime.CallersFrames([]uintptr{ptr})
	f, _ := frames.Next()
	e := Wrap(ErrClassInternal, errors.New("b"))
	assert.Equal(t, f.Function, e.Stack()[1].Function)
}

func TestExtendedInfo(t *testing.T) {
	err := NewError(ErrClassInternal, "blah")
	assert.Contains(t, err.WithExtendedInfo().Error(), "call stack")
}

func WrappingAnE(t *testing.T) {
	a := Wrap(ErrClassInternal, errors.New("a"))
	assert.Equal(t, a, Wrap(ErrClassInternal, a))
	assert.Equal(t, a, Wrap(ErrClassUser, a))
}
