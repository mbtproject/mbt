package main

import (
	"os"

	"github.com/mbtproject/mbt/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		println(err)
		os.Exit(1)
	}
}
