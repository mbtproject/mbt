package lib

const (
	// ErrClassNone Not specified
	ErrClassNone = iota
	// ErrClassUser is a user error that can be corrected
	ErrClassUser
	// ErrClassInternal is an internal error potentially due to a bug
	ErrClassInternal
)
