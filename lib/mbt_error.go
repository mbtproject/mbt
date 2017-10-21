package lib

import "fmt"

// MbtError is a generic wrapper for errors occur in various mbt components.
type MbtError struct {
	component  string
	message    string
	innerError error
}

func (e *MbtError) Error() string {
	s := fmt.Sprintf("mbt %s: %s", e.component, e.message)
	if e.message != "" && e.innerError != nil {
		s = fmt.Sprintf("%s - ", s)
	}
	if e.innerError != nil {
		s = fmt.Sprintf("%s%s", s, e.innerError)
	}
	return s
}

func newError(component, message string) *MbtError {
	return wrapm(component, nil, message)
}

func newErrorf(component, message string, args ...interface{}) *MbtError {
	return wrapm(component, nil, message, args...)
}

func wrap(component string, innerError error) *MbtError {
	return wrapm(component, innerError, "")
}

func wrapm(component string, innerError error, message string, args ...interface{}) *MbtError {
	m := fmt.Sprintf(message, args...)
	return &MbtError{component: component, message: m, innerError: innerError}
}

func failf(component string, innerError error, message string, args ...interface{}) {
	panic(wrapm(component, innerError, message, args...))
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
