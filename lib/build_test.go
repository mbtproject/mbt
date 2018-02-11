package lib

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"testing"

	"github.com/mbtproject/mbt/e"
	"github.com/stretchr/testify/assert"
)

func noopCb(a *Module, s BuildStage) {}
func TestBuildExecution(t *testing.T) {
	clean()

	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteShellScript("app-a/build.sh", "echo app-a built"))
	check(t, repo.WritePowershellScript("app-a/build.ps1", "write-host \"app-a built\""))
	check(t, repo.Commit("first"))

	m, err := ManifestByBranch(".tmp/repo", "master")
	check(t, err)

	stages := make([]BuildStage, 0)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	check(t, Build(m, os.Stdin, stdout, stderr, func(a *Module, s BuildStage) {
		stages = append(stages, s)
	}))

	assert.Equal(t, "app-a built\n", stdout.String())
	assert.EqualValues(t, []BuildStage{BuildStageBeforeBuild, BuildStageAfterBuild}, stages)
}

func TestBuildDirExecution(t *testing.T) {
	clean()

	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteShellScript("app-a/build.sh", "echo app-a built"))
	check(t, repo.WritePowershellScript("app-a/build.ps1", "write-host \"app-a built\""))
	check(t, repo.Commit("first"))
	m, err := ManifestByBranch(".tmp/repo", "master")
	check(t, err)

	stages := make([]BuildStage, 0)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	check(t, BuildDir(m, os.Stdin, stdout, stderr, func(a *Module, s BuildStage) {
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
		check(t, repo.InitModuleWithOptions("app-a", &Spec{
			Name:  "app-a",
			Build: map[string]*BuildCmd{"windows": {"powershell", []string{"-ExecutionPolicy", "Bypass", "-File", ".\\build.ps1"}}},
		}))
		check(t, repo.WritePowershellScript("app-a/build.ps1", "write-host built app-a"))
	case "windows":
		check(t, repo.InitModuleWithOptions("app-a", &Spec{
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

	check(t, Build(m, os.Stdin, buff, buff, func(a *Module, s BuildStage) {
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

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteShellScript("app-a/build.sh", "echo built app-a"))
	check(t, repo.WritePowershellScript("app-a/build.ps1", "write-host built app-a"))
	check(t, repo.Commit("first"))

	check(t, repo.SwitchToBranch("feature"))

	check(t, repo.InitModule("app-b"))
	check(t, repo.WriteShellScript("app-b/build.sh", "echo built app-b"))
	check(t, repo.WritePowershellScript("app-b/build.ps1", "write-host built app-b"))
	check(t, repo.Commit("second"))

	m, err := ManifestByBranch(".tmp/repo", "master")
	check(t, err)

	buff := new(bytes.Buffer)
	check(t, Build(m, os.Stdin, buff, buff, func(a *Module, s BuildStage) {}))

	assert.Equal(t, "built app-a\n", buff.String())

	m, err = ManifestByBranch(".tmp/repo", "feature")
	check(t, err)

	buff = new(bytes.Buffer)
	check(t, Build(m, os.Stdin, buff, buff, func(a *Module, s BuildStage) {}))

	assert.Equal(t, "built app-a\nbuilt app-b\n", buff.String())
}

func TestDirtyWorkingDir(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteContent("app-a/foo", "a"))
	check(t, repo.WriteShellScript("app-a/build.sh", "echo built app-a"))
	check(t, repo.WritePowershellScript("app-a/build.ps1", "write-host built app-a"))
	check(t, repo.Commit("first"))

	check(t, repo.WriteContent("app-a/foo", "b"))

	m, err := ManifestByBranch(".tmp/repo", "master")
	check(t, err)

	buff := new(bytes.Buffer)
	err = Build(m, os.Stdin, buff, buff, func(a *Module, s BuildStage) {})
	assert.Error(t, err)
	assert.Equal(t, "dirty working dir", err.Error())
	assert.Equal(t, ErrClassUser, (err.(*e.E)).Class())
}

func TestBuildEnvironment(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModuleWithOptions("app-a", &Spec{
		Name: "app-a",
		Build: map[string]*BuildCmd{
			"linux":   {Cmd: "./build.sh"},
			"darwin":  {Cmd: "./build.sh"},
			"windows": {Cmd: "powershell", Args: []string{"-ExecutionPolicy", "Bypass", "-File", ".\\build.ps1"}},
		},
		Properties: map[string]interface{}{"foo": "bar"},
	}))

	check(t, repo.WriteShellScript("app-a/build.sh", "echo $MBT_BUILD_COMMIT-$MBT_MODULE_VERSION-$MBT_MODULE_NAME-$MBT_MODULE_PROPERTY_FOO"))
	check(t, repo.WritePowershellScript("app-a/build.ps1", "write-host $Env:MBT_BUILD_COMMIT-$Env:MBT_MODULE_VERSION-$Env:MBT_MODULE_NAME-$Env:MBT_MODULE_PROPERTY_FOO"))
	check(t, repo.Commit("first"))

	m, err := ManifestByBranch(".tmp/repo", "master")
	check(t, err)

	buff := new(bytes.Buffer)
	err = Build(m, os.Stdin, buff, buff, noopCb)
	check(t, err)

	out := buff.String()
	assert.Equal(t, fmt.Sprintf("%s-%s-%s-%s\n", m.Sha, m.Modules[0].Version(), m.Modules[0].Name(), m.Modules[0].Properties()["foo"]), out)
}

func TestNonGitRepo(t *testing.T) {
	clean()
	check(t, os.MkdirAll(".tmp/repo", 0755))
	m := &Manifest{Dir: ".tmp/repo", Modules: []*Module{}, Sha: "a"}

	err := Build(m, os.Stdin, os.Stdout, os.Stderr, noopCb)

	assert.EqualError(t, err, "Unable to open repository in .tmp/repo")
	assert.EqualError(t, (err.(*e.E)).InnerError(), "could not find repository from '.tmp/repo'")
	assert.Equal(t, ErrClassUser, (err.(*e.E)).Class())
}

func TestBadSha(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))

	m := &Manifest{Dir: ".tmp/repo", Modules: []*Module{}, Sha: "a"}

	err = Build(m, os.Stdin, os.Stdout, os.Stderr, noopCb)

	assert.EqualError(t, err, fmt.Sprintf(msgInvalidSha, "a"))
	assert.EqualError(t, (err.(*e.E)).InnerError(), "encoding/hex: odd length hex string")
	assert.Equal(t, ErrClassUser, (err.(*e.E)).Class())
}

func TestMissingSha(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))

	sha := "22221c5e56794a2af5f59f94512df4c669c77a49"
	m := &Manifest{Dir: ".tmp/repo", Modules: []*Module{}, Sha: sha}

	err = Build(m, os.Stdin, os.Stdout, os.Stderr, noopCb)

	assert.EqualError(t, err, fmt.Sprintf(msgCommitShaNotFound, sha))
	assert.EqualError(t, (err.(*e.E)).InnerError(), "object not found - no match for id (22221c5e56794a2af5f59f94512df4c669c77a49)")
	assert.Equal(t, ErrClassUser, (err.(*e.E)).Class())
}
