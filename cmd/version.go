package cmd

import "github.com/spf13/cobra"

func init() {
	RootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Displays the version of mbt",
	Run: func(cmd *cobra.Command, args []string) {
		println("0.1.11")
	},
}
