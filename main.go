package main

import (
	"mbt/cmd"
	"os"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		println(err)
		os.Exit(1)
	}
}