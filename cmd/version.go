package cmd

import "gopkg.in/spf13/cobra.v0"

func init() {
	RootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Displays the version of mbt",
	Run: func(cmd *cobra.Command, args []string) {
		println("mbt - monorepo build tool 0.1 build #development#")
	},
}
