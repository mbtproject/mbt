package cmd

import (
	"errors"
	"os"

	"gopkg.in/sirupsen/logrus.v1"

	"github.com/mbtproject/mbt/lib"
	"github.com/spf13/cobra"
)

func init() {
	buildPr.Flags().StringVar(&src, "src", "", "source branch")
	buildPr.Flags().StringVar(&dst, "dst", "", "destination branch")

	buildDiff.Flags().StringVar(&from, "from", "", "from commit")
	buildDiff.Flags().StringVar(&to, "to", "", "to commit")

	buildCommand.AddCommand(buildBranch)
	buildCommand.AddCommand(buildPr)
	buildCommand.AddCommand(buildDiff)
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
			return handle(err)
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
			return handle(err)
		}

		return build(m)
	},
}

var buildDiff = &cobra.Command{
	Use:   "diff --from <sha> --to <sha>",
	Short: "Builds the diff between src and dst commit shas",
	RunE: func(cmd *cobra.Command, args []string) error {
		if from == "" {
			return errors.New("requires from commit")
		}

		if to == "" {
			return errors.New("requires to commit")
		}

		m, err := lib.ManifestByDiff(in, from, to)
		if err != nil {
			return handle(err)
		}

		return build(m)
	},
}

func build(m *lib.Manifest) error {
	return lib.Build(m, os.Stdin, os.Stdout, os.Stderr, func(a *lib.Module, s lib.BuildStage) {
		switch s {
		case lib.BuildStageBeforeBuild:
			logrus.Infof("BUILD %s in %s for %s", a.Name(), a.Path(), a.Version())
		case lib.BuildStageSkipBuild:
			logrus.Infof("SKIP %s in %s for %s", a.Name(), a.Path(), a.Version())
		}
	})
}

var buildCommand = &cobra.Command{
	Use:   "build",
	Short: "Builds the modules in specified path",
}
