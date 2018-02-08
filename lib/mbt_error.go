package lib

import (
	"fmt"
	"runtime"
	"strings"
)

const (
	// ErrClassNone Not specified
	ErrClassNone = iota
	// ErrClassUser is a user error that can be corrected
	ErrClassUser
	// ErrClassInternal is an internal error potentially due to a bug
	ErrClassInternal
)

// MbtError is a generic wrapper for errors occur in various mbt components.
type MbtError struct {
	message          string
	innerError       error
	class            int
	stack            []runtime.Frame
	showExtendedInfo bool
}

func (e *MbtError) Error() string {
	s := e.message
	if s == "" && e.innerError != nil {
		s = e.innerError.Error()
	}

	if e.showExtendedInfo {
		stack := make([]string, 0, len(e.stack))
		for _, f := range e.stack {
			stack = append(stack, fmt.Sprintf("%s %s(%v)", f.Function, f.File, f.Line))
		}
		return fmt.Sprintf(`%s
inner error: %s
call stack:
%s
`, s, e.innerError.Error(), strings.Join(stack, "\n"))
	}

	return s
}

// InnerError returns the inner error wrapped in this error if there's one.
func (e *MbtError) InnerError() error {
	return e.innerError
}

// Class returns the class of this error.
// See ErrClassXxx constants for possible values.
func (e *MbtError) Class() int {
	return e.class
}

// Stack returns the callstack (up to 32 frames) indicating where the
// error occurred
func (e *MbtError) Stack() []runtime.Frame {
	return e.stack
}

// WithExtendedInfo returns a new instance of MbtError which prints the
// additional details such as callstack.
func (e *MbtError) WithExtendedInfo() *MbtError {
	return &MbtError{
		class:            e.class,
		innerError:       e.innerError,
		message:          e.message,
		showExtendedInfo: true,
		stack:            e.stack,
	}
}

func newError(klass int, message string) *MbtError {
	return newMbtError(klass, nil, message)
}

func newErrorf(klass int, message string, args ...interface{}) *MbtError {
	return newMbtError(klass, nil, message, args...)
}

func wrap(klass int, innerError error) *MbtError {
	if wrapped, ok := innerError.(*MbtError); ok {
		return wrapped
	}

	return newMbtError(klass, innerError, "")
}

func wrapm(klass int, innerError error, message string, args ...interface{}) *MbtError {
	return newMbtError(klass, innerError, message, args...)
}

func failf(klass int, innerError error, message string, args ...interface{}) {
	panic(wrapm(klass, innerError, message, args...))
}

func newMbtError(klass int, innerError error, message string, args ...interface{}) *MbtError {
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
	return &MbtError{class: klass, message: m, innerError: innerError, stack: frames}
}

func handlePanic(err *error) {
	if r := recover(); r != nil {
		if _, ok := r.(*MbtError); ok {
			*err = r.(error)
		} else {
			panic(r)
		}
	}
}
