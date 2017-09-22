package cmd

import (
	"errors"
	"fmt"

	"github.com/buddyspike/mbt/lib"
	"github.com/spf13/cobra"
)

func init() {
	DescribePrCmd.Flags().StringVar(&source, "source", "", "source branch")
	DescribePrCmd.Flags().StringVar(&dest, "dest", "", "destination branch")

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
	Use:   "branch <branch>",
	Short: "describe the environment for the given branch",
	RunE: func(cmd *cobra.Command, args []string) error {
		branch := "master"
		if len(args) > 0 {
			branch = args[0]
		}
		m, err := lib.ManifestByBranch(in, branch)
		if err != nil {
			return err
		}

		output(m)
		return nil
	},
}

var DescribePrCmd = &cobra.Command{
	Use:   "pr <source> <dest>",
	Short: "describe the environment unique for a given pr",
	RunE: func(cmd *cobra.Command, args []string) error {
		if source == "" {
			return errors.New("requires source")
		}

		if dest == "" {
			return errors.New("requires dest")
		}

		m, err := lib.ManifestByPr(in, source, dest)
		if err != nil {
			return err
		}

		output(m)

		return nil
	},
}

var DescribeCommitCmd = &cobra.Command{
	Use:   "commit <sha>",
	Short: "describe the environment for a given commit",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("requires the commit sha")
		}

		commit := args[0]

		m, err := lib.ManifestBySha(in, commit)
		if err != nil {
			return err
		}

		output(m)

		return nil
	},
}
