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
	Long:  "Build utility for monorepos",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Use != "version" && in == "" {
			return errors.New("requires the path to repo")
		}
		return nil
	},
}
