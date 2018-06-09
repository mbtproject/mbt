// +build ignore

package main

import (
	"os"
	"path"

	"github.com/mbtproject/mbt/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	err = os.RemoveAll(path.Join(wd, "docs"))
	if err != nil {
		panic(err)
	}

	err = os.Mkdir(path.Join(wd, "docs"), 0755)
	if err != nil {
		panic(err)
	}

	err = doc.GenMarkdownTree(cmd.RootCmd, path.Join(wd, "docs"))
	if err != nil {
		panic(err)
	}

	err = os.Link(path.Join(wd, "docs", "mbt.md"), path.Join(wd, "docs", "index.md"))
	if err != nil {
		panic(err)
	}
}
