package cmd

import (
	"errors"

	"github.com/mbtproject/mbt/lib"
	"gopkg.in/spf13/cobra.v0"
)

var (
	out string
)

func init() {
	applyCmd.PersistentFlags().StringVar(&to, "to", "", "template to apply")
	applyCmd.PersistentFlags().StringVar(&out, "out", "", "output path")
	applyCmd.AddCommand(applyBranchCmd)
	applyCmd.AddCommand(applyCommitCmd)
	RootCmd.AddCommand(applyCmd)
}

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Applies a manifest over a template",
}

var applyBranchCmd = &cobra.Command{
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

		return handle(lib.ApplyBranch(in, to, branch, out))
	},
}

var applyCommitCmd = &cobra.Command{
	Use:   "commit <sha>",
	Short: "Applies the manifest of specified commit over a template",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("requires the commit sha")
		}

		commit := args[0]

		return handle(lib.ApplyCommit(in, commit, to, out))
	},
}
