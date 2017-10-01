package cmd

import (
	"errors"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/buddyspike/mbt/lib"
	"github.com/spf13/cobra"
)

func init() {
	buildPr.Flags().StringVar(&src, "src", "", "source branch")
	buildPr.Flags().StringVar(&dst, "dst", "", "destination branch")

	buildCommand.AddCommand(buildBranch)
	buildCommand.AddCommand(buildPr)
	RootCmd.AddCommand(buildCommand)
}

var buildBranch = &cobra.Command{
	Use:   "branch <branch>",
	Short: "Builds the specific branch",
	RunE: func(cmd *cobra.Command, args []string) error {
		branch := "master"
		if len(args) > 0 {
			branch = args[0]
		}

		m, err := lib.ManifestByBranch(in, branch)
		if err != nil {
			return err
		}

		return build(m)
	},
}

var buildPr = &cobra.Command{
	Use:   "pr --src <branch> --dst <branch>",
	Short: "Builds a pr from src branch to dst branch",
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

		return build(m)
	},
}

func build(m *lib.Manifest) error {
	return lib.Build(m, os.Stdin, os.Stdout, os.Stderr, func(a *lib.Application, s lib.BuildStage) {
		if s == lib.BUILD_STAGE_SKIP_BUILD {
			logrus.Info("Skipping build for %s", a.Name)
		}
	})
}

var buildCommand = &cobra.Command{
	Use:   "build",
	Short: "Builds the applications in specified path",
}
