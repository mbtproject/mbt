package lib

import "fmt"

// MbtError is a generic wrapper for errors occur in various mbt components.
type MbtError struct {
	component  string
	message    string
	innerError error
}

func (e *MbtError) Error() string {
	s := fmt.Sprintf("%s: %s", e.component, e.message)
	if e.innerError != nil {
		s = fmt.Sprintf("%s - %s", s, e.innerError.Error())
	}
	return s
}

func newError(component, message string, innerError error, args ...interface{}) *MbtError {
	m := fmt.Sprintf(message, args...)
	return &MbtError{component: component, message: m, innerError: innerError}
}

func panicMbt(component, message string, innerError error, args ...interface{}) {
	panic(newError(component, message, innerError, args...))
}
