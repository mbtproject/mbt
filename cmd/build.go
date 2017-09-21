package cmd

import (
	"errors"
	"os"

	"github.com/buddyspike/mbt/lib"
	"github.com/spf13/cobra"
)

func init() {
	buildCommand.AddCommand(buildBranch)
	buildCommand.AddCommand(buildPr)
	RootCmd.AddCommand(buildCommand)
}

var buildBranch = &cobra.Command{
	Use:   "branch <path> <branch>",
	Short: "builds the specific branch",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("Must specify the path to repo and branch")
		}
		path := args[0]
		branch := args[1]

		m, err := lib.ManifestByBranch(path, branch)
		if err != nil {
			return err
		}

		err = lib.Build(m, os.Args)
		return err
	},
}

var buildPr = &cobra.Command{
	Use:   "pr <path> <source> <dest>",
	Short: "builds the pr from a source branch to destination branch",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 3 {
			return errors.New("require path, source branch and destination branch")
		}

		path := args[0]
		source := args[1]
		dest := args[2]

		m, err := lib.ManifestByPr(path, source, dest)
		if err != nil {
			return err
		}

		err = lib.Build(m, os.Args)
		return err
	},
}

var buildCommand = &cobra.Command{
	Use:   "build",
	Short: "Builds the applications in specified path",
}
