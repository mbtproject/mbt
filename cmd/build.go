package cmd

import (
	"errors"
	"strings"

	"github.com/buddyspike/mbt/lib"
	"github.com/spf13/cobra"
)

var pipedArgs []string

func init() {
	buildCommand.PersistentFlags().StringArrayVarP(&pipedArgs, "args", "a", []string{}, "arguments to be passed into build scripts")
	buildCommand.AddCommand(buildBranch)
	buildCommand.AddCommand(buildPr)
	RootCmd.AddCommand(buildCommand)
}

func preparePipedArgs() []string {
	a := []string{}
	for _, i := range pipedArgs {
		if strings.Contains(i, "=") {
			k := strings.Split(i, "=")
			a = append(a, k...)
		} else {
			a = append(a, i)
		}
	}
	return a
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

		err = lib.Build(m, preparePipedArgs())
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

		err = lib.Build(m, preparePipedArgs())
		return err
	},
}

var buildCommand = &cobra.Command{
	Use:   "build",
	Short: "Builds the applications in specified path",
}
