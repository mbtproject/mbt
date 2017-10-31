package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/mbtproject/mbt/lib"
	"gopkg.in/spf13/cobra.v0"
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
	describeCmd.AddCommand(describePrCmd)
	describeCmd.AddCommand(describeIntersectionCmd)

	RootCmd.AddCommand(describeCmd)
}

var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describes the manifest of a repo",
}

var describeBranchCmd = &cobra.Command{
	Use:   "branch <branch>",
	Short: "Describes the manifest for the given branch",
	RunE: func(cmd *cobra.Command, args []string) error {
		branch := "master"
		if len(args) > 0 {
			branch = args[0]
		}
		m, err := lib.ManifestByBranch(in, branch)
		if err != nil {
			return handle(err)
		}

		output(m.Applications)
		return nil
	},
}

var describePrCmd = &cobra.Command{
	Use:   "pr --src <branch> --dst <branch>",
	Short: "Describes the manifest for a given pr",
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

		output(m.Applications)

		return nil
	},
}

var describeCommitCmd = &cobra.Command{
	Use:   "commit <sha>",
	Short: "Describes the manifest for a given commit",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("requires the commit sha")
		}

		commit := args[0]

		m, err := lib.ManifestBySha(in, commit)
		if err != nil {
			return handle(err)
		}

		output(m.Applications)

		return nil
	},
}

var describeIntersectionCmd = &cobra.Command{
	Use:   "intersection --kind <branch|commit> --first <first> --second <second>",
	Short: "Describes the intersection between two trees",
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

		var apps lib.Applications
		var err error

		switch kind {
		case "branch":
			apps, err = lib.IntersectionByBranch(in, first, second)
		case "commit":
			apps, err = lib.IntersectionByCommit(in, first, second)
		default:
			err = errors.New("not a valid kind - available options are 'branch' and 'commit'")
		}

		if err != nil {
			return handle(err)
		}

		return handle(output(apps))
	},
}

const columnWidth = 30

func formatRow(args ...interface{}) string {
	padded := make([]interface{}, len(args))
	for i, a := range args {
		requiredPadding := columnWidth - len(a.(string))
		if requiredPadding > 0 {
			padded[i] = fmt.Sprintf("%s%s", a, strings.Join(make([]string, requiredPadding), " "))
		} else {
			padded[i] = a
		}
	}
	return fmt.Sprintf("%s\t\t%s\t\t%s\n", padded...)
}

func output(apps lib.Applications) error {
	if formatAsJson {
		v := make(map[string]interface{})
		for _, a := range apps {
			v["Name"] = a.Name()
			v["Path"] = a.Path()
			v["Version"] = a.Version()
			v["Properties"] = a.Properties()
		}
		buff, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(buff))
	} else {
		fmt.Print(formatRow("NAME", "PATH", "VERSION"))
		for _, a := range apps {
			fmt.Printf(formatRow(a.Name(), a.Path(), a.Version()))
		}
	}

	return nil
}
