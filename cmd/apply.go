package cmd

import (
	"errors"

	"github.com/buddyspike/mbt/lib"
	"github.com/spf13/cobra"
)

func init() {
	ApplyCmd.AddCommand(ApplyBranchCmd)
	RootCmd.AddCommand(ApplyCmd)
}

var ApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "applies a manifest over a template",
}

var ApplyBranchCmd = &cobra.Command{
	Use:   "branch <path> <branch> <template> [out]",
	Short: "applies the manifest of specified branch over a given template",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 3 {
			return errors.New("required path, branch and template")
		}
		dir := args[0]
		branch := args[1]
		template := args[2]
		out := ""

		if len(args) > 3 {
			out = args[3]
		}

		return lib.ApplyBranch(dir, template, branch, out)
	},
}
