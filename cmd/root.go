package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

// Flags available to all commands.
var (
	in     string
	src    string
	dst    string
	from   string
	to     string
	first  string
	second string
	kind   string
)

func init() {
	RootCmd.PersistentFlags().StringVar(&in, "in", "", "path to repo")
}

// RootCmd is the main command.
var RootCmd = &cobra.Command{
	Use:   "mbt",
	Short: "Build utility for monorepos",
	Long: `Build utility for monorepos

mbt is a cli tool for building monorepos stored in git. 
It supports differential builds, dependency tracking and 
transforming templates based on data inferred from the state of 
the repository.

All commands in mbt should specify the path to the repository via 
--in argument.

See help for individual commands for more information.

	`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Use != "version" && in == "" {
			return errors.New("requires the path to repo")
		}
		return nil
	},
}
