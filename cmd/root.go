package cmd

import (
	"os"

	"github.com/mbtproject/mbt/lib"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Flags available to all commands.
var (
	in      string
	src     string
	dst     string
	from    string
	to      string
	first   string
	second  string
	kind    string
	name    string
	all     bool
	debug   bool
	content bool
	fuzzy   bool
	system  lib.System
)

func init() {
	RootCmd.PersistentFlags().StringVar(&in, "in", "", "Path to repo")
	RootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debugging")
}

// RootCmd is the main command.
var RootCmd = &cobra.Command{
	Use:   "mbt",
	Short: "Monorepo Build Tool",
	Long: `Monorepo Build Tool
The most flexible build tool for monorepo.

See help for individual commands for more information.

	`,
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Use == "version" {
			return nil
		}

		if in == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			in, err = lib.GitRepoRoot(cwd)
			if err != nil {
				return err
			}
		}

		level := lib.LogLevelNormal
		if debug {
			logrus.SetLevel(logrus.DebugLevel)
			level = lib.LogLevelDebug
		}

		var err error
		system, err = lib.NewSystem(in, level)
		return err
	},
}
