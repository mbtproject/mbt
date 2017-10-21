package lib

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func noopCb(a *Application, s BuildStage) {}
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

	stages := make([]BuildStage, 0)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	check(t, Build(m, os.Stdin, stdout, stderr, func(a *Application, s BuildStage) {
		stages = append(stages, s)
	}))

	assert.Equal(t, "app-a built\n", stdout.String())
	assert.EqualValues(t, []BuildStage{BuildStageBeforeBuild, BuildStageAfterBuild}, stages)
}

func TestBuildSkip(t *testing.T) {
	clean()

	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	switch runtime.GOOS {
	case "linux", "darwin":
		check(t, repo.InitApplicationWithOptions("app-a", &Spec{
			Name:  "app-a",
			Build: map[string]*BuildCmd{"windows": {"powershell", []string{"-ExecutionPolicy", "Bypass", "-File", ".\\build.ps1"}}},
		}))
		check(t, repo.WritePowershellScript("app-a/build.ps1", "write-host built app-a"))
	case "windows":
		check(t, repo.InitApplicationWithOptions("app-a", &Spec{
			Name:  "app-a",
			Build: map[string]*BuildCmd{"darwin": {"./build.sh", []string{}}},
		}))
		check(t, repo.WriteShellScript("app-a/build.sh", "echo built app-a"))
	}

	check(t, repo.Commit("first"))
	m, err := ManifestByBranch(".tmp/repo", "master")
	check(t, err)

	skipped := make([]string, 0)
	other := make([]string, 0)
	buff := new(bytes.Buffer)

	check(t, Build(m, os.Stdin, buff, buff, func(a *Application, s BuildStage) {
		if s == BuildStageSkipBuild {
			skipped = append(skipped, a.Name())
		} else {
			other = append(other, a.Name())
		}
	}))

	assert.EqualValues(t, []string{"app-a"}, skipped)
	assert.EqualValues(t, []string{}, other)
}

func TestBuildBranch(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitApplication("app-a"))
	check(t, repo.WriteShellScript("app-a/build.sh", "echo built app-a"))
	check(t, repo.WritePowershellScript("app-a/build.ps1", "write-host built app-a"))
	check(t, repo.Commit("first"))

	check(t, repo.SwitchToBranch("feature"))

	check(t, repo.InitApplication("app-b"))
	check(t, repo.WriteShellScript("app-b/build.sh", "echo built app-b"))
	check(t, repo.WritePowershellScript("app-b/build.ps1", "write-host built app-b"))
	check(t, repo.Commit("second"))

	m, err := ManifestByBranch(".tmp/repo", "master")
	check(t, err)

	buff := new(bytes.Buffer)
	check(t, Build(m, os.Stdin, buff, buff, func(a *Application, s BuildStage) {}))

	assert.Equal(t, "built app-a\n", buff.String())

	m, err = ManifestByBranch(".tmp/repo", "feature")
	check(t, err)

	buff = new(bytes.Buffer)
	check(t, Build(m, os.Stdin, buff, buff, func(a *Application, s BuildStage) {}))

	assert.Equal(t, "built app-a\nbuilt app-b\n", buff.String())
}

func TestDirtyWorkingDir(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitApplication("app-a"))
	check(t, repo.WriteContent("app-a/foo", "a"))
	check(t, repo.WriteShellScript("app-a/build.sh", "echo built app-a"))
	check(t, repo.WritePowershellScript("app-a/build.ps1", "write-host built app-a"))
	check(t, repo.Commit("first"))

	check(t, repo.WriteContent("app-a/foo", "b"))

	m, err := ManifestByBranch(".tmp/repo", "master")
	check(t, err)

	buff := new(bytes.Buffer)
	err = Build(m, os.Stdin, buff, buff, func(a *Application, s BuildStage) {})
	assert.Error(t, err)
	assert.Equal(t, "mbt build: dirty working dir", err.Error())
}

func TestBuildEnvironment(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitApplicationWithOptions("app-a", &Spec{
		Name: "app-a",
		Build: map[string]*BuildCmd{
			"linux":   {Cmd: "./build.sh"},
			"darwin":  {Cmd: "./build.sh"},
			"windows": {Cmd: "powershell", Args: []string{"-ExecutionPolicy", "Bypass", "-File", ".\\build.ps1"}},
		},
		Properties: map[string]interface{}{"foo": "bar"},
	}))

	check(t, repo.WriteShellScript("app-a/build.sh", "echo $MBT_BUILD_COMMIT-$MBT_APP_VERSION-$MBT_APP_NAME-$MBT_APP_PROPERTY_FOO"))
	check(t, repo.WritePowershellScript("app-a/build.ps1", "write-host $Env:MBT_BUILD_COMMIT-$Env:MBT_APP_VERSION-$Env:MBT_APP_NAME-$Env:MBT_APP_PROPERTY_FOO"))
	check(t, repo.Commit("first"))

	m, err := ManifestByBranch(".tmp/repo", "master")
	check(t, err)

	buff := new(bytes.Buffer)
	err = Build(m, os.Stdin, buff, buff, noopCb)
	check(t, err)

	out := buff.String()
	assert.Equal(t, fmt.Sprintf("%s-%s-%s-%s\n", m.Sha, m.Applications[0].Version(), m.Applications[0].Name(), m.Applications[0].Properties()["foo"]), out)
}

func TestNonGitRepo(t *testing.T) {
	clean()
	check(t, os.MkdirAll(".tmp/repo", 0755))
	m := &Manifest{Dir: ".tmp/repo", Applications: []*Application{}, Sha: "a"}

	err := Build(m, os.Stdin, os.Stdout, os.Stderr, noopCb)

	assert.EqualError(t, err, "mbt build: could not find repository from '.tmp/repo'")
}

func TestBadSha(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitApplication("app-a"))
	check(t, repo.Commit("first"))

	m := &Manifest{Dir: ".tmp/repo", Applications: []*Application{}, Sha: "a"}

	err = Build(m, os.Stdin, os.Stdout, os.Stderr, noopCb)

	assert.EqualError(t, err, "mbt build: encoding/hex: odd length hex string")
}

func TestMissingSha(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitApplication("app-a"))
	check(t, repo.Commit("first"))

	m := &Manifest{Dir: ".tmp/repo", Applications: []*Application{}, Sha: "22221c5e56794a2af5f59f94512df4c669c77a49"}

	err = Build(m, os.Stdin, os.Stdout, os.Stderr, noopCb)

	assert.EqualError(t, err, "mbt build: object not found - no match for id (22221c5e56794a2af5f59f94512df4c669c77a49)")
}
