package cmd

import "github.com/buddyspike/mbt/lib"

func handle(e error) error {
	if me, ok := e.(*lib.MbtError); ok {
		return me.WithLocation()
	}

	return e
}
