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

	"github.com/mbtproject/mbt/e"
	"github.com/mbtproject/mbt/lib"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Flags available to all commands.
var (
	in       string
	src      string
	dst      string
	from     string
	to       string
	first    string
	second   string
	kind     string
	name     string
	command  string
	all      bool
	debug    bool
	content  bool
	fuzzy    bool
	failFast bool
	system   lib.System
)

func init() {
	RootCmd.PersistentFlags().StringVar(&in, "in", "", "Path to repo")
	RootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug output")
}

// RootCmd is the main command.
var RootCmd = &cobra.Command{
	Use:          "mbt",
	Short:        docText("main-summary"),
	Long:         docText("main"),
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

		parent := cmd.Parent()
		if parent != nil && parent.Name() == "run-in" && command == "" {
			return e.NewError(lib.ErrClassUser, "--command (-m) is not specified")
		}
		if parent != nil && parent.Name() == "describe" && dependents && name == "" {
			return e.NewError(lib.ErrClassUser, "--dependents flag can only be specified with the --name (-n) flag")
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
