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

	buildLocal.Flags().BoolVarP(&all, "all", "a", false, "all modules")

	buildCommand.AddCommand(buildBranch)
	buildCommand.AddCommand(buildPr)
	buildCommand.AddCommand(buildDiff)
	buildCommand.AddCommand(buildHead)
	buildCommand.AddCommand(buildCommit)
	buildCommand.AddCommand(buildLocal)
	RootCmd.AddCommand(buildCommand)
}

var buildHead = &cobra.Command{
	Use:   "head",
	Short: "Build all modules in the commit pointed at current head",
	Long: `Build all modules in the commit pointed at current head

`,
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		return system.BuildCurrentBranch(os.Stdin, os.Stdout, os.Stderr, buildStageCB)
	}),
}

var buildBranch = &cobra.Command{
	Use:   "branch <branch>",
	Short: "Build all modules at the tip of the specified branch",
	Long: `Build all modules at the tip of the specified branch

Build all modules at the tip of the specified branch.
If branch name is not specified, the command assumes 'master'.

`,
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		branch := "master"
		if len(args) > 0 {
			branch = args[0]
		}

		return system.BuildBranch(branch, os.Stdin, os.Stdout, os.Stderr, buildStageCB)
	}),
}

var buildPr = &cobra.Command{
	Use:   "pr --src <branch> --dst <branch>",
	Short: "Build the modules changed in dst branch relatively to src branch",
	Long: `Build the modules changed in dst branch relatively to src branch

This command works out the merge base for src and dst branches and 
builds all modules impacted by the diff between merge base and 
the tip of dst branch.	

In addition to the modules impacted by changes, this command also 
builds their dependents.

	`,
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		if src == "" {
			return errors.New("requires source")
		}

		if dst == "" {
			return errors.New("requires dest")
		}

		return system.BuildPr(src, dst, os.Stdin, os.Stdout, os.Stderr, buildStageCB)
	}),
}

var buildDiff = &cobra.Command{
	Use:   "diff --from <sha> --to <sha>",
	Short: "Build the modules changed between from and to commits",
	Long: `Build the modules chanaged between from and to commits

Works out the merge base for from and to commits and 
builds all modules which have been changed between merge base and 
to commit.

In addition to the modules impacted by changes, this command also 
builds their dependents.

Commit SHA must be the complete 40 character SHA1 string.
	`,
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		if from == "" {
			return errors.New("requires from commit")
		}

		if to == "" {
			return errors.New("requires to commit")
		}

		return system.BuildDiff(from, to, os.Stdin, os.Stdout, os.Stderr, buildStageCB)
	}),
}

var buildCommit = &cobra.Command{
	Use:   "commit <sha>",
	Short: "Build all modules in the specified commit",
	Long: `Build all modules in the specified commit
	`,
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("requires the commit sha")
		}

		commit := args[0]

		return system.BuildCommit(commit, os.Stdin, os.Stdout, os.Stderr, buildStageCB)
	}),
}

var buildLocal = &cobra.Command{
	Use:   "local [--all]",
	Short: "Build all modules in uncommitted changes in current workspace",
	Long: `Build all modules in uncommitted changes in current workspace

Local builds always uses a fixed version identifier - 'local'.
Specify the --all flag to build all modules in current workspace. 
	`,
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		if all {
			return system.BuildWorkspace(os.Stdin, os.Stdout, os.Stderr, buildStageCB)
		}

		return system.BuildWorkspaceChanges(os.Stdin, os.Stdout, os.Stderr, buildStageCB)
	}),
}

func buildStageCB(a *lib.Module, s lib.BuildStage) {
	switch s {
	case lib.BuildStageBeforeBuild:
		logrus.Infof("BUILD %s in %s for %s", a.Name(), a.Path(), a.Version())
	case lib.BuildStageSkipBuild:
		logrus.Infof("SKIP %s in %s for %s", a.Name(), a.Path(), a.Version())
	}
}

var buildCommand = &cobra.Command{
	Use:   "build",
	Short: "Main command for building the repository",
	Long: `Main command for building the repository 

Execute the build according to the specified sub command. 
See sub command help for more information.
`,
}
