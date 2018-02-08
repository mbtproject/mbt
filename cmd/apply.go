package cmd

import (
	"errors"
	"io"
	"os"

	"github.com/mbtproject/mbt/lib"
	"github.com/spf13/cobra"
)

var (
	out string
)

func init() {
	applyCmd.PersistentFlags().StringVar(&to, "to", "", "template to apply")
	applyCmd.PersistentFlags().StringVar(&out, "out", "", "output path")
	applyCmd.AddCommand(applyBranchCmd)
	applyCmd.AddCommand(applyCommitCmd)
	applyCmd.AddCommand(applyHeadCmd)
	applyCmd.AddCommand(applyLocal)
	RootCmd.AddCommand(applyCmd)
}

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Main command for applying the repository manifest over a template",
	Long: `Main command for applying the repository manifest over a template 

Repository manifest is a data structure created by inspecting .mbt.yml files.
It contains the information about the modules stored within the repository therefore,
can be used for generating artifacts such as deployment scripts.

Apply command transforms the specified go template with the manifest. 

Template must be committed to the repository.
	`,
}

var applyBranchCmd = &cobra.Command{
	Use:   "branch <branch>",
	Short: "Applies the manifest of specified branch over a template",
	Long: `Applies the manifest of specified branch over a template 

Calculated manifest and the template is based on the tip of the specified branch.
	`,
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		branch := "master"
		if len(args) > 0 {
			branch = args[0]
		}

		return applyCore(func(to string, output io.Writer) error {
			return lib.ApplyBranch(in, to, branch, output)
		})
	}),
}

var applyCommitCmd = &cobra.Command{
	Use:   "commit <sha>",
	Short: "Applies the manifest of specified commit over a template",
	Long: `Applies the manifest of specified commit over a template

Calculated manifest and the template is based on the specified commit.

Commit SHA must be the complete 40 character SHA1 string.
	`,
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("requires the commit sha")
		}

		commit := args[0]

		return applyCore(func(to string, output io.Writer) error {
			return lib.ApplyCommit(in, commit, to, output)
		})
	}),
}

var applyHeadCmd = &cobra.Command{
	Use:   "head",
	Short: "Applies the manifest of current head over a template",
	Long: `Applies the manifest of current head over a template

Calculated manifest and the template is based on the current head.
	`,
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		return applyCore(func(to string, output io.Writer) error {
			return lib.ApplyHead(in, to, output)
		})
	}),
}

var applyLocal = &cobra.Command{
	Use:   "local",
	Short: "Applies the manifest of local directory state over a template",
	Long: `Applies the manifest of local directory state over a template

Calculated manifest and the template is based on the content of local directory.
This command is useful for testing pending changes in workspace.
	`,
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		return applyCore(func(to string, output io.Writer) error {
			return lib.ApplyLocal(in, to, output)
		})
	}),
}

type applyFunc func(to string, output io.Writer) error

func applyCore(f applyFunc) error {
	if to == "" {
		return errors.New("requires the path to template, specify --to argument")
	}

	output, err := getOutput(out)
	if err != nil {
		return err
	}

	return f(to, output)
}

func getOutput(out string) (io.Writer, error) {
	if out == "" {
		return os.Stdout, nil
	}
	return os.Create(out)
}
