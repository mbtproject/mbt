package cmd

import (
	"os"

	"github.com/mbtproject/mbt/lib"
	"github.com/sirupsen/logrus"
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
	all    bool
	debug  bool
	system lib.System
)

func init() {
	RootCmd.PersistentFlags().StringVar(&in, "in", "", "path to repo")
	RootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debugging")
}

// RootCmd is the main command.
var RootCmd = &cobra.Command{
	Use:   "mbt",
	Short: "Build utility for monorepos",
	Long: `Build utility for monorepos

Monorepo Build Tool (mbt) is a utility that supports differential builds,
dependency tracking and metadata management for monorepos stored in git.

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
			in = cwd
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
