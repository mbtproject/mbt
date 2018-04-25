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
	"fmt"
	"runtime"
	"strings"
)

// E is a container for errors.
// In addition to the standard error message, it also contains the
// stack trace and a few other important attributes.
type E struct {
	message          string
	innerError       error
	class            int
	stack            []runtime.Frame
	showExtendedInfo bool
}

func (e *E) Error() string {
	s := e.message
	if s == "" && e.innerError != nil {
		s = e.innerError.Error()
	}

	if e.showExtendedInfo {
		innerErr := ""
		if e.innerError != nil {
			innerErr = e.innerError.Error()
		}

		stack := make([]string, 0, len(e.stack))
		for _, f := range e.stack {
			stack = append(stack, fmt.Sprintf("%s %s(%v)", f.Function, f.File, f.Line))
		}
		return fmt.Sprintf(`%s
inner error: %s
call stack:
%s
`, s, innerErr, strings.Join(stack, "\n"))
	}

	return s
}

// InnerError returns the inner error wrapped in this error if there's one.
func (e *E) InnerError() error {
	return e.innerError
}

// Class returns the class of this error.
func (e *E) Class() int {
	return e.class
}

// Stack returns the callstack (up to 32 frames) indicating where the
// error occurred
func (e *E) Stack() []runtime.Frame {
	return e.stack
}

// WithExtendedInfo returns a new instance of E which prints the
// additional details such as callstack.
func (e *E) WithExtendedInfo() *E {
	return &E{
		class:            e.class,
		innerError:       e.innerError,
		message:          e.message,
		showExtendedInfo: true,
		stack:            e.stack,
	}
}

// NewError creates a new E
func NewError(klass int, message string) *E {
	return newE(klass, nil, message)
}

// NewErrorf creates a new E with an interpolated message
func NewErrorf(klass int, message string, args ...interface{}) *E {
	return newE(klass, nil, message, args...)
}

// Wrap creates a new E that wraps another error
func Wrap(klass int, innerError error) *E {
	if wrapped, ok := innerError.(*E); ok {
		return wrapped
	}

	return newE(klass, innerError, "")
}

// Wrapf creates a new wrapped E with an interpolated message
func Wrapf(klass int, innerError error, message string, args ...interface{}) *E {
	return newE(klass, innerError, message, args...)
}

// Failf panics with a wrapped E containing an interpolated message
func Failf(klass int, innerError error, message string, args ...interface{}) {
	panic(Wrapf(klass, innerError, message, args...))
}

func newE(klass int, innerError error, message string, args ...interface{}) *E {
	m := fmt.Sprintf(message, args...)
	pc := make([]uintptr, 32 /*limit the number of frames to 32 max*/)
	n := runtime.Callers(2, pc)
	frames := make([]runtime.Frame, 0, n)
	rframes := runtime.CallersFrames(pc[:n])
	for {
		f, more := rframes.Next()
		frames = append(frames, f)
		if !more {
			break
		}
	}
	return &E{class: klass, message: m, innerError: innerError, stack: frames}
}
