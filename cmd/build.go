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
	buildCommand.AddCommand(buildHead)
	RootCmd.AddCommand(buildCommand)
}

var buildHead = &cobra.Command{
	Use:   "head",
	Short: "Builds the local directory",
	Long: `Builds the local directory

Builds all modules as of the head of the branch that is in the specified folder.

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := lib.ManifestByHead(in)
		if err != nil {
			return handle(err)
		}

		return build(m)
	},
}

var buildBranch = &cobra.Command{
	Use:   "branch <branch>",
	Short: "Builds the specified branch",
	Long: `Builds the specified branch

Builds all modules as of the tip of specified branch.
If branch name is not specified assumes 'master'.

`,
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
	Short: "Builds the diff between the src and dst branches",
	Long: `Builds the diff between the src and dst branches

Works out the merge base for src and dst branches and 
builds all modules which have been changed between merge base and 
the tip of dst branch.	

	`,
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
	Short: "Builds the diff between from and to commits",
	Long: `Builds the diff between from and to commits

Works out the merge base for from and to commits and 
builds all modules which have been changed between merge base and 
to commit.

Commit SHA must be the complete 40 character SHA1 string.
	`,
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
	Short: "Main command for building the repository",
	Long: `Main command for building the repository 

Executes the build according to the specified sub command. 
See sub command help for more information.
`,
}
