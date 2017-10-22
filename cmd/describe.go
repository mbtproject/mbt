package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mbtproject/mbt/lib"
	"github.com/spf13/cobra"
)

func init() {
	describePrCmd.Flags().StringVar(&src, "src", "", "source branch")
	describePrCmd.Flags().StringVar(&dst, "dst", "", "destination branch")

	describeCmd.AddCommand(describeCommitCmd)
	describeCmd.AddCommand(describeBranchCmd)
	describeCmd.AddCommand(describePrCmd)
	RootCmd.AddCommand(describeCmd)
}

var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describes the manifest of a repo",
}

var describeBranchCmd = &cobra.Command{
	Use:   "branch <branch>",
	Short: "Describes the manifest for the given branch",
	RunE: func(cmd *cobra.Command, args []string) error {
		branch := "master"
		if len(args) > 0 {
			branch = args[0]
		}
		m, err := lib.ManifestByBranch(in, branch)
		if err != nil {
			return handle(err)
		}

		output(m)
		return nil
	},
}

var describePrCmd = &cobra.Command{
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
			return handle(err)
		}

		output(m)

		return nil
	},
}

var describeCommitCmd = &cobra.Command{
	Use:   "commit <sha>",
	Short: "Describes the manifest for a given commit",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("requires the commit sha")
		}

		commit := args[0]

		m, err := lib.ManifestBySha(in, commit)
		if err != nil {
			return handle(err)
		}

		output(m)

		return nil
	},
}

const columnWidth = 30

func formatRow(args ...interface{}) string {
	padded := make([]interface{}, len(args))
	for i, a := range args {
		requiredPadding := columnWidth - len(a.(string))
		if requiredPadding > 0 {
			padded[i] = fmt.Sprintf("%s%s", a, strings.Join(make([]string, requiredPadding), " "))
		} else {
			padded[i] = a
		}
	}
	return fmt.Sprintf("%s\t\t%s\t\t%s\n", padded...)
}

func output(m *lib.Manifest) {
	fmt.Print(formatRow("Name", "Path", "Version"))
	for _, a := range m.Applications {
		fmt.Printf(formatRow(a.Name(), a.Path(), a.Version()))
	}
}
