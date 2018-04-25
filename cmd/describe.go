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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/mbtproject/mbt/lib"
	"github.com/spf13/cobra"
)

var (
	toJSON  bool
	toGraph bool
)

func init() {
	describePrCmd.Flags().StringVar(&src, "src", "", "Source branch")
	describePrCmd.Flags().StringVar(&dst, "dst", "", "Destination branch")

	describeIntersectionCmd.Flags().StringVar(&kind, "kind", "", "Kind of input for first and second args (available options are 'branch' and 'commit')")
	describeIntersectionCmd.Flags().StringVar(&first, "first", "", "First item")
	describeIntersectionCmd.Flags().StringVar(&second, "second", "", "Second item")

	describeDiffCmd.Flags().StringVar(&from, "from", "", "From commit")
	describeDiffCmd.Flags().StringVar(&to, "to", "", "To commit")
	describeLocalCmd.Flags().BoolVarP(&all, "all", "a", false, "Describe all")
	describeLocalCmd.Flags().StringVarP(&name, "name", "n", "", "Describe modules with a name that matches this value. Multiple names can be specified as a comma separated string.")
	describeLocalCmd.Flags().BoolVarP(&fuzzy, "fuzzy", "f", false, "Use fuzzy match when filtering")

	describeCommitCmd.Flags().BoolVarP(&content, "content", "c", false, "Describe the modules impacted by the changes in commit")
	describeCommitCmd.Flags().StringVarP(&name, "name", "n", "", "Describe modules with a name that matches this value. Multiple names can be specified as a comma separated string.")
	describeCommitCmd.Flags().BoolVarP(&fuzzy, "fuzzy", "f", false, "Use fuzzy match when filtering")

	describeBranchCmd.Flags().StringVarP(&name, "name", "n", "", "Describe modules with a name that matches this value. Multiple names can be specified as a comma separated string.")
	describeBranchCmd.Flags().BoolVarP(&fuzzy, "fuzzy", "f", false, "Use fuzzy match when filtering")

	describeHeadCmd.Flags().StringVarP(&name, "name", "n", "", "Describe modules with a name that matches this value. Multiple names can be specified as a comma separated string.")
	describeHeadCmd.Flags().BoolVarP(&fuzzy, "fuzzy", "f", false, "Use fuzzy match when filtering")

	describeCmd.PersistentFlags().BoolVar(&toJSON, "json", false, "Format output as json")
	describeCmd.PersistentFlags().BoolVar(&toGraph, "graph", false, "Format output as dot graph")
	describeCmd.AddCommand(describeCommitCmd)
	describeCmd.AddCommand(describeBranchCmd)
	describeCmd.AddCommand(describeHeadCmd)
	describeCmd.AddCommand(describeLocalCmd)
	describeCmd.AddCommand(describePrCmd)
	describeCmd.AddCommand(describeIntersectionCmd)
	describeCmd.AddCommand(describeDiffCmd)

	RootCmd.AddCommand(describeCmd)
}

var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe the modules in the repo",
	Long: `Describe the modules in the repo

Displays all modules discovered in repo according to the sub command
used. This can be used to understand the impact of executing the build
command.

You can control the output of this command using --json and --graph flags.

--json flag is self explanatory.
--graph flag outputs the dependency graph in dot format. This can be piped
into graphviz tools such as dot to produce a graphical representation of
the dependency graph of the repo.

e.g.
mbt describe head --graph | dot -Tpng > /tmp/graph.png && open /tmp/graph.png
`,
}

var describeBranchCmd = &cobra.Command{
	Use:   "branch <branch>",
	Short: "Describe all modules in the tip of the given branch",
	Long: `Describe all modules in the tip of the given branch

If branch name is not specified assumes 'master'.
`,
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		branch := "master"
		if len(args) > 0 {
			branch = args[0]
		}
		m, err := system.ManifestByBranch(branch)
		if err != nil {
			return err
		}

		m = m.ApplyFilters(&lib.FilterOptions{Name: name, Fuzzy: fuzzy})

		return output(m.Modules)
	}),
}

var describeHeadCmd = &cobra.Command{
	Use:   "head",
	Short: "Describe all modules in the branch pointed by head",
	Long: `Describe all modules in the branch pointed by head

`,
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		m, err := system.ManifestByCurrentBranch()
		if err != nil {
			return err
		}

		m = m.ApplyFilters(&lib.FilterOptions{Name: name, Fuzzy: fuzzy})
		return output(m.Modules)
	}),
}

var describeLocalCmd = &cobra.Command{
	Use:   "local",
	Short: "Describe all modules in current workspace",
	Long: `Describe all modules in current workspace

This includes the modules in uncommitted changes.
`,
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		var (
			m   *lib.Manifest
			err error
		)

		if all {
			m, err = system.ManifestByWorkspace()
			m = m.ApplyFilters(&lib.FilterOptions{Name: name, Fuzzy: fuzzy})
		} else {
			m, err = system.ManifestByWorkspaceChanges()
		}

		if err != nil {
			return err
		}

		return output(m.Modules)
	}),
}

var describePrCmd = &cobra.Command{
	Use:   "pr --src <branch> --dst <branch>",
	Short: "Describe the modules changed between src and dst branches",
	Long: `Describe the modules changed between src and dst branches

Works out the merge base for src and dst branches and
displays all modules which have been changed between merge base and
the tip of dst branch.
`,
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		if src == "" {
			return errors.New("requires source")
		}

		if dst == "" {
			return errors.New("requires dest")
		}

		m, err := system.ManifestByPr(src, dst)
		if err != nil {
			return err
		}

		return output(m.Modules)
	}),
}

var describeCommitCmd = &cobra.Command{
	Use:   "commit <sha>",
	Short: "Describe all modules in the specified commit",
	Long: `Describe all modules in the specified commit

Commit SHA must be the complete 40 character SHA1 string.

If --content flag is specified, this command will describe just the modules
impacted by the specified commit.
`,
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("requires the commit sha")
		}

		commit := args[0]

		var (
			m   *lib.Manifest
			err error
		)

		if content {
			m, err = system.ManifestByCommitContent(commit)
		} else {
			m, err = system.ManifestByCommit(commit)
		}

		if err != nil {
			return err
		}

		m = m.ApplyFilters(&lib.FilterOptions{Name: name, Fuzzy: fuzzy})

		return output(m.Modules)
	}),
}

var describeIntersectionCmd = &cobra.Command{
	Use:   "intersection --kind <branch|commit> --first <first> --second <second>",
	Short: "Describe the common modules changed in two commit trees",
	Long: `Describe the common modules changed in two commit trees
`,
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		if kind == "" {
			return errors.New("requires the kind argument")
		}

		if first == "" {
			return errors.New("requires the first argument")
		}

		if second == "" {
			return errors.New("requires the second argument")
		}

		var mods lib.Modules
		var err error

		switch kind {
		case "branch":
			mods, err = system.IntersectionByBranch(first, second)
		case "commit":
			mods, err = system.IntersectionByCommit(first, second)
		default:
			err = errors.New("not a valid kind - available options are 'branch' and 'commit'")
		}

		if err != nil {
			return err
		}

		return output(mods)
	}),
}

var describeDiffCmd = &cobra.Command{
	Use:   "diff --from <commit> --to <commit>",
	Short: "Describe the modules in the diff between from and to commits",
	Long: `Describe the modules in the diff between from and to commits

Works out the merge base for from and to commits and
displays all modules which have been changed between merge base and
the to commit.
`,
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		if from == "" {
			return errors.New("requires from commit")
		}

		if to == "" {
			return errors.New("requires to commit")
		}

		m, err := system.ManifestByDiff(from, to)
		if err != nil {
			return err
		}

		return output(m.Modules)
	}),
}

const columnWidth = 30

func output(mods lib.Modules) error {
	if toJSON {
		m := make(map[string]map[string]interface{})
		for _, a := range mods {
			v := make(map[string]interface{})
			v["Name"] = a.Name()
			v["Path"] = a.Path()
			v["Version"] = a.Version()
			v["Properties"] = a.Properties()
			m[a.Name()] = v
		}
		buff, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(buff))
	} else if toGraph {
		fmt.Println(mods.SerializeAsDot())
	} else {
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 4, ' ', 0)
		fmt.Fprintf(w, "Name\tPATH\tVERSION\n")
		for _, a := range mods {
			fmt.Fprintf(w, "%s\t%s\t%s\n", a.Name(), a.Path(), a.Version())
		}

		if err := w.Flush(); err != nil {
			panic(err)
		}
	}

	return nil
}
