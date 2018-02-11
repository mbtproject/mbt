package cmd

import (
	"fmt"

	"github.com/mbtproject/mbt/e"
	"github.com/mbtproject/mbt/lib"
	"github.com/spf13/cobra"
)

type handlerFunc func(command *cobra.Command, args []string) error

func buildHandler(handler handlerFunc) handlerFunc {
	return func(command *cobra.Command, args []string) error {
		err := handler(command, args)
		if err == nil {
			return nil
		}

		if ee, ok := err.(*e.E); ok {
			if ee.Class() == lib.ErrClassInternal {
				return fmt.Errorf(`An unexpected error occurred. See below for more details.
For support, create a new issue at https://github.com/mbtproject/mbt/issues

%v`, ee.WithExtendedInfo())
			} else if debug {
				return ee.WithExtendedInfo()
			}
		}

		return err
	}
}
