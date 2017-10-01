package lib

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildExecution(t *testing.T) {
	clean()

	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitApplication("app-a"))
	check(t, repo.WriteShellScript("app-a/build.sh", "echo app-a built"))
	check(t, repo.WritePowershellScript("app-a/build.ps1", "write-host \"app-a built\""))
	check(t, repo.Commit("first"))

	m, err := ManifestByBranch(".tmp/repo", "master")
	check(t, err)

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	check(t, Build(m, os.Stdin, stdout, stderr))

	assert.Equal(t, "app-a built\n", stdout.String())
}
