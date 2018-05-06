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
	"io"
	"os"

	"github.com/spf13/cobra"
)

var (
	out string
)

func init() {
	applyCmd.PersistentFlags().StringVar(&to, "to", "", "Template to apply")
	applyCmd.PersistentFlags().StringVar(&out, "out", "", "Output path")
	applyCmd.AddCommand(applyBranchCmd)
	applyCmd.AddCommand(applyCommitCmd)
	applyCmd.AddCommand(applyHeadCmd)
	applyCmd.AddCommand(applyLocal)
	RootCmd.AddCommand(applyCmd)
}

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: docText("apply-summary"),
	Long:  docText("apply"),
}

var applyBranchCmd = &cobra.Command{
	Use: "branch <branch>",
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		branch := "master"
		if len(args) > 0 {
			branch = args[0]
		}

		return applyCore(func(to string, output io.Writer) error {
			return system.ApplyBranch(to, branch, output)
		})
	}),
}

var applyCommitCmd = &cobra.Command{
	Use: "commit <sha>",
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("requires the commit sha")
		}

		commit := args[0]

		return applyCore(func(to string, output io.Writer) error {
			return system.ApplyCommit(commit, to, output)
		})
	}),
}

var applyHeadCmd = &cobra.Command{
	Use: "head",
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		return applyCore(func(to string, output io.Writer) error {
			return system.ApplyHead(to, output)
		})
	}),
}

var applyLocal = &cobra.Command{
	Use: "local",
	RunE: buildHandler(func(cmd *cobra.Command, args []string) error {
		return applyCore(func(to string, output io.Writer) error {
			return system.ApplyLocal(to, output)
		})
	}),
}

type applyFunc func(to string, output io.Writer) error

func applyCore(f applyFunc) error {
	if to == "" {
		return errors.New("requires the path to template, specify --to argument")
	}

	output, err := getOutput(out)
	if err != nil {
		return err
	}

	return f(to, output)
}

func getOutput(out string) (io.Writer, error) {
	if out == "" {
		return os.Stdout, nil
	}
	return os.Create(out)
}
