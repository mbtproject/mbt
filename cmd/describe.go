package cmd

import (
	"errors"
	"fmt"

	"github.com/buddyspike/mbt/lib"
	"github.com/spf13/cobra"
)

func init() {
	DescribeCmd.AddCommand(DescribeCommitCmd)
	DescribeCmd.AddCommand(DescribeBranchCmd)
	DescribeCmd.AddCommand(DescribePrCmd)
	RootCmd.AddCommand(DescribeCmd)
}

func output(m *lib.Manifest) {
	for _, a := range m.Applications {
		fmt.Printf("%s %s\n", a.Application.Name, a.Version)
	}
}

var DescribeCmd = &cobra.Command{
	Use:   "describe",
	Short: "describe the repo",
}

var DescribeBranchCmd = &cobra.Command{
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

		output(m)
		return nil
	},
}

var DescribePrCmd = &cobra.Command{
	Use:   "pr <path> <source> <dest>",
	Short: "describe the environment unique for a given pr",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 3 {
			return errors.New("requires path and source and dest")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		source := args[1]
		dest := args[2]

		m, err := lib.ManifestByPr(path, source, dest)
		if err != nil {
			return err
		}

		output(m)

		return nil
	},
}

var DescribeCommitCmd = &cobra.Command{
	Use:   "commit <path> <sha>",
	Short: "describe the environment for a given commit",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("requires path and commit")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		commit := args[1]

		m, err := lib.ManifestBySha(path, commit)
		if err != nil {
			return err
		}

		output(m)

		return nil
	},
}
