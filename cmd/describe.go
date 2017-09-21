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
	Use:   "environment",
	Short: "describe the environment at the given rev",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires the path to repo")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var path, rev string
		path = args[0]
		if len(args) > 1 {
			rev = args[1]
		}
		m, err := lib.ManifestByBranch(path, rev)
		if err != nil {
			return err
		}

		for _, a := range m.Applications {
			fmt.Printf("%s %s\n", a.Application.Name, a.Version)
		}

		return nil
	},
}
