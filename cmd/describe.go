package cmd

import (
	"errors"
	"fmt"

	"github.com/buddyspike/mbt/lib"
	"github.com/spf13/cobra"
)

func init() {
	DescribePrCmd.Flags().StringVar(&src, "src", "", "source branch")
	DescribePrCmd.Flags().StringVar(&dst, "dst", "", "destination branch")

	DescribeCmd.AddCommand(DescribeCommitCmd)
	DescribeCmd.AddCommand(DescribeBranchCmd)
	DescribeCmd.AddCommand(DescribePrCmd)
	RootCmd.AddCommand(DescribeCmd)
}

var DescribeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describes the manifest of a repo",
}

var DescribeBranchCmd = &cobra.Command{
	Use:   "branch <branch>",
	Short: "Describes the manifest for the given branch",
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
	Use:   "pr --src <branch> --dst <branch>",
	Short: "Describes the manifest for a given pr",
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

		output(m)

		return nil
	},
}

var DescribeCommitCmd = &cobra.Command{
	Use:   "commit <sha>",
	Short: "Describes the manifest for a given commit",
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

func output(m *lib.Manifest) {
	for _, a := range m.Applications {
		fmt.Printf("%s %s\n", a.Name, a.Version)
	}
}
