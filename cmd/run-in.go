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

	"github.com/sirupsen/logrus"

	"github.com/mbtproject/mbt/lib"
	"github.com/spf13/cobra"
)

func init() {
	runIn.PersistentFlags().StringVarP(&command, "command", "m", "", "Command to execute")

	runInPr.Flags().StringVar(&src, "src", "", "Source branch")
	runInPr.Flags().StringVar(&dst, "dst", "", "Destination branch")

	runInDiff.Flags().StringVar(&from, "from", "", "From commit")
	runInDiff.Flags().StringVar(&to, "to", "", "To commit")

	runInLocal.Flags().BoolVarP(&all, "all", "a", false, "All modules")
	runInLocal.Flags().StringVarP(&name, "name", "n", "", "Build modules with a name that matches this value. Multiple names can be specified as a comma separated string.")
	runInLocal.Flags().BoolVarP(&fuzzy, "fuzzy", "f", false, "Use fuzzy match when filtering")

	runInCommit.Flags().BoolVarP(&content, "content", "c", false, "Build the modules impacted by the content of the commit")
	runInCommit.Flags().StringVarP(&name, "name", "n", "", "Build modules with a name that matches this value. Multiple names can be specified as a comma separated string.")
	runInCommit.Flags().BoolVarP(&fuzzy, "fuzzy", "f", false, "Use fuzzy match when filtering")

	runInBranch.Flags().StringVarP(&name, "name", "n", "", "Build modules with a name that matches this value. Multiple names can be specified as a comma separated string.")
	runInBranch.Flags().BoolVarP(&fuzzy, "fuzzy", "f", false, "Use fuzzy match when filtering")

	runInHead.Flags().StringVarP(&name, "name", "n", "", "Build modules with a name that matches this value. Multiple names can be specified as a comma separated string.")
	runInHead.Flags().BoolVarP(&fuzzy, "fuzzy", "f", false, "Use fuzzy match when filtering")

	runIn.AddCommand(runInBranch)
	runIn.AddCommand(runInPr)
	runIn.AddCommand(runInDiff)
	runIn.AddCommand(runInHead)
	runIn.AddCommand(runInCommit)
	runIn.AddCommand(runInLocal)
	RootCmd.AddCommand(runIn)
}

var runInHead = &cobra.Command{
	Use:   "head",
	Short: "Run command in all modules in the commit pointed at current head",
	Long: `Run command in modules in the commit pointed at current head

`,
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		return summariseRun(system.RunInCurrentBranch(command, &lib.FilterOptions{Name: name, Fuzzy: fuzzy}, lib.CmdOptionsWithStdIO(runCmdStageCB)))
	}),
}

var runInBranch = &cobra.Command{
	Use:   "branch <branch>",
	Short: "Run command in all modules at the tip of the specified branch",
	Long: `Build all modules at the tip of the specified branch

Build all modules at the tip of the specified branch.
If branch name is not specified, the command assumes 'master'.

`,
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		branch := "master"
		if len(args) > 0 {
			branch = args[0]
		}

		return summariseRun(system.RunInBranch(command, branch, &lib.FilterOptions{Name: name, Fuzzy: fuzzy}, lib.CmdOptionsWithStdIO(runCmdStageCB)))
	}),
}

var runInPr = &cobra.Command{
	Use:   "pr --src <branch> --dst <branch>",
	Short: "Run command in modules changed in dst branch relatively to src branch",
	Long: `Run command in modules changed in dst branch relatively to src branch

This command works out the merge base for src and dst branches and
runs the specified command in all modules impacted by the diff between merge base and
the tip of dst branch.

In addition to the modules impacted by changes, command is also
executed in their dependents (recursively).
	`,
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		if src == "" {
			return errors.New("requires source")
		}

		if dst == "" {
			return errors.New("requires dest")
		}

		return summariseRun(system.RunInPr(command, src, dst, lib.CmdOptionsWithStdIO(runCmdStageCB)))
	}),
}

var runInDiff = &cobra.Command{
	Use:   "diff --from <sha> --to <sha>",
	Short: "Run command in modules changed between from and to commits",
	Long: `Run command in modules chanaged between from and to commits

Works out the merge base for from and to commits and
runs the command in modules which have been changed between merge base and
to commit.

In addition to the modules impacted by changes, command is also
executed in their dependents.

Commit SHA must be the complete 40 character SHA1 string.
	`,
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		if from == "" {
			return errors.New("requires from commit")
		}

		if to == "" {
			return errors.New("requires to commit")
		}

		return summariseRun(system.RunInDiff(command, from, to, lib.CmdOptionsWithStdIO(runCmdStageCB)))
	}),
}

var runInCommit = &cobra.Command{
	Use:   "commit <sha>",
	Short: "Run command in all modules in the specified commit",
	Long: `Run command in all modules in the specified commit

	If --content flag is specified, command is run in just the modules
	impacted by the specified commit.
	`,
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("requires the commit sha")
		}

		commit := args[0]

		if content {
			return summariseRun(system.RunInCommitContent(command, commit, lib.CmdOptionsWithStdIO(runCmdStageCB)))
		}
		return summariseRun(system.RunInCommit(command, commit, &lib.FilterOptions{Name: name, Fuzzy: fuzzy}, lib.CmdOptionsWithStdIO(runCmdStageCB)))
	}),
}

var runInLocal = &cobra.Command{
	Use:   "local [--all]",
	Short: "Run command in all modules in uncommitted changes in current workspace",
	Long: `Run command in all modules in uncommitted changes in current workspace

Specify the --all flag to run command in all modules in current workspace.
	`,
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		if all || name != "" {
			return summariseRun(system.RunInWorkspace(command, &lib.FilterOptions{Name: name, Fuzzy: fuzzy}, lib.CmdOptionsWithStdIO(runCmdStageCB)))
		}

		return summariseRun(system.RunInWorkspaceChanges(command, lib.CmdOptionsWithStdIO(runCmdStageCB)))
	}),
}

func runCmdStageCB(a *lib.Module, s lib.CmdStage) {
	switch s {
	case lib.CmdStageBeforeBuild:
		logrus.Infof("RUN command %s in module %s (path: %s version: %s)", command, a.Name(), a.Path(), a.Version())
	case lib.CmdStageSkipBuild:
		logrus.Infof("SKIP %s in %s for %s", a.Name(), a.Path(), a.Version())
	}
}

func summariseRun(summary *lib.RunResult, err error) error {
	if err == nil {
		logrus.Infof("Modules: %v Success: %v Failed: %v Skipped: %v",
			len(summary.Manifest.Modules),
			len(summary.Completed),
			len(summary.Failures),
			len(summary.Skipped))

		logrus.Infof("Build finished for commit %v", summary.Manifest.Sha)
	}
	return err
}

var runIn = &cobra.Command{
	Use:   "run-in",
	Short: "Main command for running user defined commands",
	Long: `Main command for running user defined commands

Execute a user defined command according to the specified sub command.
See sub command help for more information.
`,
}
