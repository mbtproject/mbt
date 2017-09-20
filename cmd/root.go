package cmd

import "github.com/spf13/cobra"

var RootCmd = &cobra.Command{
	Use:   "mbt",
	Short: "Monorepo build tool",
	Long:  "Build utility for monorepos",
	Run: func(cmd *cobra.Command, args []string) {
		println("yolo")
	},
}
