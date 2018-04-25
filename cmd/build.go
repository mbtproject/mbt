/*
Copyright 2018 MBT Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"errors"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/mbtproject/mbt/lib"
	"github.com/spf13/cobra"
)

func init() {
	buildPr.Flags().StringVar(&src, "src", "", "Source branch")
	buildPr.Flags().StringVar(&dst, "dst", "", "Destination branch")

	buildDiff.Flags().StringVar(&from, "from", "", "From commit")
	buildDiff.Flags().StringVar(&to, "to", "", "To commit")

	buildLocal.Flags().BoolVarP(&all, "all", "a", false, "All modules")
	buildLocal.Flags().StringVarP(&name, "name", "n", "", "Build modules with a name that matches this value. Multiple names can be specified as a comma separated string.")
	buildLocal.Flags().BoolVarP(&fuzzy, "fuzzy", "f", false, "Use fuzzy match when filtering")

	buildCommit.Flags().BoolVarP(&content, "content", "c", false, "Build the modules impacted by the content of the commit")
	buildCommit.Flags().StringVarP(&name, "name", "n", "", "Build modules with a name that matches this value. Multiple names can be specified as a comma separated string.")
	buildCommit.Flags().BoolVarP(&fuzzy, "fuzzy", "f", false, "Use fuzzy match when filtering")

	buildBranch.Flags().StringVarP(&name, "name", "n", "", "Build modules with a name that matches this value. Multiple names can be specified as a comma separated string.")
	buildBranch.Flags().BoolVarP(&fuzzy, "fuzzy", "f", false, "Use fuzzy match when filtering")

	buildHead.Flags().StringVarP(&name, "name", "n", "", "Build modules with a name that matches this value. Multiple names can be specified as a comma separated string.")
	buildHead.Flags().BoolVarP(&fuzzy, "fuzzy", "f", false, "Use fuzzy match when filtering")

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
		return summarise(system.BuildCurrentBranch(&lib.FilterOptions{Name: name, Fuzzy: fuzzy}, os.Stdin, os.Stdout, os.Stderr, buildStageCB))
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

		return summarise(system.BuildBranch(branch, &lib.FilterOptions{Name: name, Fuzzy: fuzzy}, os.Stdin, os.Stdout, os.Stderr, buildStageCB))
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

		return summarise(system.BuildPr(src, dst, os.Stdin, os.Stdout, os.Stderr, buildStageCB))
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

		return summarise(system.BuildDiff(from, to, os.Stdin, os.Stdout, os.Stderr, buildStageCB))
	}),
}

var buildCommit = &cobra.Command{
	Use:   "commit <sha>",
	Short: "Build all modules in the specified commit",
	Long: `Build all modules in the specified commit

	If --content flag is specified, this command will build just the modules
	impacted by the specified commit.
	`,
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("requires the commit sha")
		}

		commit := args[0]

		if content {
			return summarise(system.BuildCommitContent(commit, os.Stdin, os.Stdout, os.Stderr, buildStageCB))
		}
		return summarise(system.BuildCommit(commit, &lib.FilterOptions{Name: name, Fuzzy: fuzzy}, os.Stdin, os.Stdout, os.Stderr, buildStageCB))
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
		if all || name != "" {
			return summarise(system.BuildWorkspace(&lib.FilterOptions{Name: name, Fuzzy: fuzzy}, os.Stdin, os.Stdout, os.Stderr, buildStageCB))
		}

		return summarise(system.BuildWorkspaceChanges(os.Stdin, os.Stdout, os.Stderr, buildStageCB))
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

func summarise(summary *lib.BuildSummary, err error) error {
	if err == nil {
		logrus.Infof("Modules: %v Built: %v Skipped: %v",
			len(summary.Manifest.Modules),
			len(summary.Completed),
			len(summary.Skipped))

		logrus.Infof("Build finished for commit %v", summary.Manifest.Sha)
	}
	return err
}

var buildCommand = &cobra.Command{
	Use:   "build",
	Short: "Main command for building the repository",
	Long: `Main command for building the repository

Execute the build according to the specified sub command.
See sub command help for more information.
`,
}
