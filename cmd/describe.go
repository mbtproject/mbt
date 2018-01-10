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
	formatAsJson bool
)

func init() {
	describePrCmd.Flags().StringVar(&src, "src", "", "source branch")
	describePrCmd.Flags().StringVar(&dst, "dst", "", "destination branch")

	describeIntersectionCmd.Flags().StringVar(&kind, "kind", "", "kind of input for first and second args (available options are 'branch' and 'commit')")
	describeIntersectionCmd.Flags().StringVar(&first, "first", "", "first item")
	describeIntersectionCmd.Flags().StringVar(&second, "second", "", "second item")

	describeCmd.PersistentFlags().BoolVar(&formatAsJson, "json", false, "format output as json")
	describeCmd.AddCommand(describeCommitCmd)
	describeCmd.AddCommand(describeBranchCmd)
	describeCmd.AddCommand(describeHeadCmd)
	describeCmd.AddCommand(describePrCmd)
	describeCmd.AddCommand(describeIntersectionCmd)

	RootCmd.AddCommand(describeCmd)
}

var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describes the manifest of a repo",
	Long: `Describes the manifest of a repo

Displays all modules discovered in repo according to the sub command 
used. This can be used to understand the impact of executing the build 
command and also to pipe mbt manifest to external tools.
	`,
}

var describeBranchCmd = &cobra.Command{
	Use:   "branch <branch>",
	Short: "Describes the manifest for the given branch",
	Long: `Describes the manifest for the given branch

Displays all modules as of the tip of specified branch.
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

		return handle(output(m.Modules))
	},
}

var describeHeadCmd = &cobra.Command{
	Use:   "head",
	Short: "Describes the manifest for the branch pointed at head",
	Long: `Describes the manifest for the branch pointed at head

Displays all modules as of the tip of the branch pointed at head.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := lib.ManifestByHead(in)
		if err != nil {
			return handle(err)
		}

		return handle(output(m.Modules))
	},
}

var describePrCmd = &cobra.Command{
	Use:   "pr --src <branch> --dst <branch>",
	Short: "Describes the manifest for diff between src and dst branches",
	Long: `Describes the manifest for diff between src and dst branches

Works out the merge base for src and dst branches and 
displays all modules which have been changed between merge base and 
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

		return handle(output(m.Modules))
	},
}

var describeCommitCmd = &cobra.Command{
	Use:   "commit <sha>",
	Short: "Describes the manifest of a specified commit",
	Long: `Describes the manifest of a specified commit

Displays all modules as of the specified commit.

Commit SHA must be the complete 40 character SHA1 string.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("requires the commit sha")
		}

		commit := args[0]

		m, err := lib.ManifestBySha(in, commit)
		if err != nil {
			return handle(err)
		}

		return handle(output(m.Modules))
	},
}

var describeIntersectionCmd = &cobra.Command{
	Use:   "intersection --kind <branch|commit> --first <first> --second <second>",
	Short: "Describes the intersection between two commit trees",
	Long: `Describes the intersection between two commit trees
	
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
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
			mods, err = lib.IntersectionByBranch(in, first, second)
		case "commit":
			mods, err = lib.IntersectionByCommit(in, first, second)
		default:
			err = errors.New("not a valid kind - available options are 'branch' and 'commit'")
		}

		if err != nil {
			return handle(err)
		}

		return handle(output(mods))
	},
}

const columnWidth = 30

func output(mods lib.Modules) error {
	if formatAsJson {
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
	} else {
		if len(mods) == 0 {
			fmt.Println("No modules found")
			return nil
		}

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
