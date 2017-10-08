package cmd

import (
	"errors"

	"github.com/buddyspike/mbt/lib"
	"github.com/spf13/cobra"
)

var (
	out string
)

func init() {
	ApplyCmd.PersistentFlags().StringVar(&to, "to", "", "template to apply")
	ApplyCmd.PersistentFlags().StringVar(&out, "out", "", "output path")
	ApplyCmd.AddCommand(ApplyBranchCmd)
	ApplyCmd.AddCommand(ApplyCommitCmd)
	RootCmd.AddCommand(ApplyCmd)
}

var ApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Applies a manifest over a template",
}

var ApplyBranchCmd = &cobra.Command{
	Use:   "branch <branch>",
	Short: "Applies the manifest of specified branch over a template",
	RunE: func(cmd *cobra.Command, args []string) error {
		branch := "master"
		if len(args) > 0 {
			branch = args[0]
		}

		if to == "" {
			return errors.New("requires the path to template")
		}

		return lib.ApplyBranch(in, to, branch, out)
	},
}

var ApplyCommitCmd = &cobra.Command{
	Use:   "commit <sha>",
	Short: "Applies the manifest of specified commit over a template",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("requires the commit sha")
		}

		commit := args[0]

		return lib.ApplyCommit(in, commit, to, out)
	},
}
