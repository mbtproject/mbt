package cmd

import (
	"errors"
	"strings"

	"github.com/buddyspike/mbt/lib"
	"github.com/spf13/cobra"
)

var pipedArgs []string

func init() {
	buildPr.Flags().StringVar(&src, "src", "", "source branch")
	buildPr.Flags().StringVar(&dst, "dst", "", "destination branch")

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
	Use:   "branch <branch>",
	Short: "Builds the specific branch",
	RunE: func(cmd *cobra.Command, args []string) error {
		branch := "master"
		if len(args) > 0 {
			branch = args[0]
		}

		m, err := lib.ManifestByBranch(in, branch)
		if err != nil {
			return err
		}

		err = lib.Build(m, preparePipedArgs())
		return err
	},
}

var buildPr = &cobra.Command{
	Use:   "pr --src <branch> --dst <branch>",
	Short: "Builds a pr from src branch to dst branch",
	RunE: func(cmd *cobra.Command, args []string) error {
		if src == "" {
			return errors.New("requires source")
		}

		if dst == "" {
			return errors.New("requires dest")
		}

		m, err := lib.ManifestByPr(in, src, dst)
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
