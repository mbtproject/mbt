package cmd

import (
	"errors"
	"fmt"

	"github.com/buddyspike/mbt/lib"
	"github.com/spf13/cobra"
)

func init() {
	DescribeCmd.AddCommand(DescribeEnvironment)
	RootCmd.AddCommand(DescribeCmd)
}

var DescribeCmd = &cobra.Command{
	Use:   "describe",
	Short: "describe the repo",
}

var DescribeEnvironment = &cobra.Command{
	Use:   "branch <path> <branch>",
	Short: "describe the environment for the given branch",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("requires path and branch")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		branch := args[1]
		m, err := lib.ManifestByBranch(path, branch)
		if err != nil {
			return err
		}

		for _, a := range m.Applications {
			fmt.Printf("%s %s\n", a.Application.Name, a.Version)
		}

		return nil
	},
}
