package lib

import (
	"fmt"
	"path"
	"runtime"
)

// MbtError is a generic wrapper for errors occur in various mbt components.
type MbtError struct {
	message      string
	innerError   error
	file         string
	line         int
	showLocation bool
}

func (e *MbtError) Error() string {
	s := ""
	if e.showLocation {
		s = fmt.Sprintf("mbt (%s,%v): %s", e.file, e.line, e.message)
	} else {
		s = fmt.Sprintf("mbt: %s", e.message)
	}
	if e.message != "" && e.innerError != nil {
		s = fmt.Sprintf("%s - ", s)
	}
	if e.innerError != nil {
		s = fmt.Sprintf("%s%s", s, e.innerError)
	}
	return s
}

// Line returns the line number where e was created.
func (e *MbtError) Line() int {
	return e.line
}

// File returns the filename where e was created.
func (e *MbtError) File() string {
	return e.file
}

// WithLocation enables the error location info in the output.
func (e *MbtError) WithLocation() *MbtError {
	e.showLocation = true
	return e
}

func newError(message string) *MbtError {
	return newMbtError(nil, message)
}

func newErrorf(message string, args ...interface{}) *MbtError {
	return newMbtError(nil, message, args...)
}

func wrap(innerError error) *MbtError {
	return newMbtError(innerError, "")
}

func wrapm(innerError error, message string, args ...interface{}) *MbtError {
	return newMbtError(innerError, message, args...)
}

func failf(innerError error, message string, args ...interface{}) {
	panic(newMbtError(innerError, message, args...))
}

func newMbtError(innerError error, message string, args ...interface{}) *MbtError {
	m := fmt.Sprintf(message, args...)
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	} else {
		file = path.Base(file)
	}
	return &MbtError{message: m, innerError: innerError, file: file, line: line}
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
