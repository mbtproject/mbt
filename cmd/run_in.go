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
	Use: "head",
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		return summariseRun(system.RunInCurrentBranch(command, &lib.FilterOptions{Name: name, Fuzzy: fuzzy}, lib.CmdOptionsWithStdIO(runCmdStageCB)))
	}),
}

var runInBranch = &cobra.Command{
	Use: "branch <branch>",
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		branch := "master"
		if len(args) > 0 {
			branch = args[0]
		}

		return summariseRun(system.RunInBranch(command, branch, &lib.FilterOptions{Name: name, Fuzzy: fuzzy}, lib.CmdOptionsWithStdIO(runCmdStageCB)))
	}),
}

var runInPr = &cobra.Command{
	Use: "pr --src <branch> --dst <branch>",
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
	Use: "diff --from <sha> --to <sha>",
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
	Use: "commit <sha>",
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
	Use: "local [--all]",
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
	Short: docText("run-in-summary"),
	Long:  docText("run-in"),
}
