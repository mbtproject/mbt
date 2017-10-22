package cmd

import "github.com/mbtproject/mbt/lib"

func handle(e error) error {
	if me, ok := e.(*lib.MbtError); ok {
		return me.WithLocation()
	}

	return e
}
