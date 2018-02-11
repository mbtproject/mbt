package lib

import "github.com/mbtproject/mbt/e"

const (
	// ErrClassNone Not specified
	ErrClassNone = iota
	// ErrClassUser is a user error that can be corrected
	ErrClassUser
	// ErrClassInternal is an internal error potentially due to a bug
	ErrClassInternal
)

func handlePanic(err *error) {
	if r := recover(); r != nil {
		if _, ok := r.(*e.E); ok {
			*err = r.(error)
		} else {
			panic(r)
		}
	}
}
