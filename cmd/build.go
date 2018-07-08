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
	Use: "head",
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		return summarise(system.BuildCurrentBranch(&lib.FilterOptions{Name: name, Fuzzy: fuzzy}, lib.CmdOptionsWithStdIO(buildStageCB)))
	}),
}

var buildBranch = &cobra.Command{
	Use: "branch <branch>",
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		branch := "master"
		if len(args) > 0 {
			branch = args[0]
		}

		return summarise(system.BuildBranch(branch, &lib.FilterOptions{Name: name, Fuzzy: fuzzy}, lib.CmdOptionsWithStdIO(buildStageCB)))
	}),
}

var buildPr = &cobra.Command{
	Use: "pr --src <branch> --dst <branch>",
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		if src == "" {
			return errors.New("requires source")
		}

		if dst == "" {
			return errors.New("requires dest")
		}

		return summarise(system.BuildPr(src, dst, lib.CmdOptionsWithStdIO(buildStageCB)))
	}),
}

var buildDiff = &cobra.Command{
	Use: "diff --from <sha> --to <sha>",
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		if from == "" {
			return errors.New("requires from commit")
		}

		if to == "" {
			return errors.New("requires to commit")
		}

		return summarise(system.BuildDiff(from, to, lib.CmdOptionsWithStdIO(buildStageCB)))
	}),
}

var buildCommit = &cobra.Command{
	Use: "commit <sha>",
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("requires the commit sha")
		}

		commit := args[0]

		if content {
			return summarise(system.BuildCommitContent(commit, lib.CmdOptionsWithStdIO(buildStageCB)))
		}
		return summarise(system.BuildCommit(commit, &lib.FilterOptions{Name: name, Fuzzy: fuzzy}, lib.CmdOptionsWithStdIO(buildStageCB)))
	}),
}

var buildLocal = &cobra.Command{
	Use: "local [--all]",
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		if all || name != "" {
			return summarise(system.BuildWorkspace(&lib.FilterOptions{Name: name, Fuzzy: fuzzy}, lib.CmdOptionsWithStdIO(buildStageCB)))
		}

		return summarise(system.BuildWorkspaceChanges(lib.CmdOptionsWithStdIO(buildStageCB)))
	}),
}

func buildStageCB(a *lib.Module, s lib.CmdStage, err error) {
	switch s {
	case lib.CmdStageBeforeBuild:
		logrus.Infof("BUILD %s in %s for %s", a.Name(), a.Path(), a.Version())
	case lib.CmdStageSkipBuild:
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
	Short: docText("build-summary"),
	Long:  docText("build"),
}
