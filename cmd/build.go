package cmd

import (
	"errors"
	"mbt/lib"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(buildCommand)
}

var buildCommand = &cobra.Command{
	Use:   "build",
	Short: "Builds the applications in specified path",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("must specify the path to source")
		}
		path := args[0]
		lib.ResolveChanges(path)
		return nil
	},
}
