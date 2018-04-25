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

package intercept

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Target interface {
	F1() int
	F2(int) int
	F3(int, int) int
}

type TestTarget struct {
}

func (t *TestTarget) F1() int {
	return 42
}

func (t *TestTarget) F2(i int) int {
	return i
}

func (t *TestTarget) F3(i int, j int) int {
	return i + j
}

func (t *TestTarget) F4() (int, error) {
	return 42, nil
}

func (t *TestTarget) F5() (*TestTarget, error) {
	return nil, errors.New("doh")
}

func TestSingleReturn(t *testing.T) {
	target := &TestTarget{}
	i := NewInterceptor(target)
	assert.Equal(t, 42, i.Call("F1")[0].(int))
}

func TestSpecificReturn(t *testing.T) {
	target := &TestTarget{}
	i := NewInterceptor(target)
	i.Config("F1").Return(24)
	assert.Equal(t, 24, i.Call("F1")[0].(int))
}

func TestSpecificDo(t *testing.T) {
	target := &TestTarget{}
	i := NewInterceptor(target)
	i.Config("F1").Do(func(args ...interface{}) []interface{} {
		return []interface{}{24}
	})

	assert.Equal(t, 24, i.Call("F1")[0].(int))
}

func TestSingleInput(t *testing.T) {
	target := &TestTarget{}
	i := NewInterceptor(target)
	assert.Equal(t, 42, i.Call("F2", 42)[0].(int))
}

func TestMultipleInput(t *testing.T) {
	target := &TestTarget{}
	i := NewInterceptor(target)
	assert.Equal(t, 42, i.Call("F3", 21, 21)[0].(int))
}

func TestInterfaceDispatch(t *testing.T) {
	var target Target
	target = &TestTarget{}
	i := NewInterceptor(target)
	assert.Equal(t, 42, i.Call("F1")[0].(int))
}

func TestMultipleConfigurationsOfSameMethod(t *testing.T) {
	target := &TestTarget{}
	i := NewInterceptor(target)
	i.Config("F1").Return(24)
	i.Config("F1").Return(32)
	assert.Equal(t, 32, i.Call("F1")[0].(int))
}

func TestNullConfigCalls(t *testing.T) {
	target := &TestTarget{}
	i := NewInterceptor(target)
	i.Config("F1")
	assert.Equal(t, 42, i.Call("F1")[0].(int))
}

func TestInvalidMethod(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			assert.Fail(t, "should panic")
		}
	}()

	target := &TestTarget{}
	i := NewInterceptor(target)
	i.Config("Foo").Return(42)
	i.Call("Foo")
}

func TestErrors(t *testing.T) {
	target := &TestTarget{}
	i := NewInterceptor(target)
	assert.Nil(t, i.Call("F5")[0].(*TestTarget))
}
