package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

var globalArgIn string

func init() {
	RootCmd.PersistentFlags().StringVar(&globalArgIn, "in", "", "path to repo")
}

var RootCmd = &cobra.Command{
	Use:   "mbt",
	Short: "Monorepo build tool",
	Long:  "Build utility for monorepos",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if globalArgIn == "" {
			return errors.New("requires the path to repo")
		}
		return nil
	},
}
