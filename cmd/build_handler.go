package cmd

import (
	"fmt"

	"github.com/mbtproject/mbt/lib"
	"github.com/spf13/cobra"
)

type handlerFunc func(command *cobra.Command, args []string) error

func buildHandler(handler handlerFunc) handlerFunc {
	return func(command *cobra.Command, args []string) error {
		e := handler(command, args)
		if e == nil {
			return nil
		}

		if mbtError, ok := e.(*lib.MbtError); ok {
			if mbtError.Class() == lib.ErrClassInternal {
				return fmt.Errorf(`An unexpected error occurred. See below for more details.
For support, create a new issue at https://github.com/mbtproject/mbt/issues

%v`, mbtError.WithExtendedInfo())
			} else if debug {
				return mbtError.WithExtendedInfo()
			}
		}

		return e
	}
}
