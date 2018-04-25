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
