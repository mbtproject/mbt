module github.com/mbtproject/mbt

go 1.15

replace github.com/libgit2/git2go/v34 => ./git2go

require (
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/libgit2/git2go/v34 v34.0.0
	github.com/mattn/goveralls v0.0.11 // indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.1
	github.com/stretchr/testify v1.8.2
	golang.org/x/crypto v0.7.0 // indirect
	golang.org/x/tools/cmd/cover v0.1.0-deprecated // indirect
)
