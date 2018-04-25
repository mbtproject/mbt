/*
Copyright 2018 MBT Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
