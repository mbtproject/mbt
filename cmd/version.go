package cmd

import "github.com/spf13/cobra"

func init() {
	RootCmd.AddCommand(VersionCmd)
}

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Displays the version of mbt",
	Run: func(cmd *cobra.Command, args []string) {
		println("mbt - monorepo build tool 0.1 build #development#")
	},
}
