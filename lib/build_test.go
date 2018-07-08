/*
Copyright 2018 MBT Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package lib

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/libgit2/git2go"
	"github.com/mbtproject/mbt/e"
	"github.com/stretchr/testify/assert"
)

//noinspection GoUnusedParameter
func noopCb(a *Module, s CmdStage, err error) {}

func stdTestCmdOptions(buff *bytes.Buffer) *CmdOptions {
	var stdout io.Writer = buff
	var stderr io.Writer = buff
	if buff == nil {
		stdout = os.Stdout
		stderr = os.Stderr
	}

	return &CmdOptions{Callback: noopCb, Stdin: os.Stdin, Stdout: stdout, Stderr: stderr}
}

func TestBuildExecution(t *testing.T) {
	clean()

	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteShellScript("app-a/build.sh", "echo app-a built"))
	check(t, repo.WritePowershellScript("app-a/build.ps1", "write-host \"app-a built\""))
	check(t, repo.Commit("first"))

	stages := make([]CmdStage, 0)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	_, err := NewWorld(t, ".tmp/repo").System.BuildCurrentBranch(NoFilter, &CmdOptions{
		Callback: func(a *Module, s CmdStage, err error) {
			stages = append(stages, s)
		},
		Stdin:  os.Stdin,
		Stdout: stdout,
		Stderr: stderr,
	})
	check(t, err)

	assert.Equal(t, "app-a built\n", stdout.String())
	assert.EqualValues(t, []CmdStage{CmdStageBeforeBuild, CmdStageAfterBuild}, stages)
}

func TestBuildDirExecution(t *testing.T) {
	clean()

	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteShellScript("app-a/build.sh", "echo app-a built"))
	check(t, repo.WritePowershellScript("app-a/build.ps1", "write-host \"app-a built\""))
	check(t, repo.Commit("first"))

	stages := make([]CmdStage, 0)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	_, err := NewWorld(t, ".tmp/repo").System.BuildWorkspace(NoFilter, &CmdOptions{
		Callback: func(a *Module, s CmdStage, err error) {
			stages = append(stages, s)
		},
		Stdin:  os.Stdin,
		Stdout: stdout,
		Stderr: stderr,
	})
	check(t, err)

	assert.Equal(t, "app-a built\n", stdout.String())
	assert.EqualValues(t, []CmdStage{CmdStageBeforeBuild, CmdStageAfterBuild}, stages)
}

func TestBuildSkip(t *testing.T) {
	clean()

	repo := NewTestRepo(t, ".tmp/repo")

	switch runtime.GOOS {
	case "linux", "darwin":
		check(t, repo.InitModuleWithOptions("app-a", &Spec{
			Name:  "app-a",
			Build: map[string]*Cmd{"windows": {"powershell", []string{"-ExecutionPolicy", "Bypass", "-File", ".\\build.ps1"}}},
		}))
		check(t, repo.WritePowershellScript("app-a/build.ps1", "write-host built app-a"))
	case "windows":
		check(t, repo.InitModuleWithOptions("app-a", &Spec{
			Name:  "app-a",
			Build: map[string]*Cmd{"darwin": {"./build.sh", []string{}}},
		}))
		check(t, repo.WriteShellScript("app-a/build.sh", "echo built app-a"))
	}

	check(t, repo.Commit("first"))

	skipped := make([]string, 0)
	other := make([]string, 0)
	buff := new(bytes.Buffer)

	_, err := NewWorld(t, ".tmp/repo").System.BuildCurrentBranch(NoFilter, &CmdOptions{
		Callback: func(a *Module, s CmdStage, err error) {
			if s == CmdStageSkipBuild {
				skipped = append(skipped, a.Name())
			} else {
				other = append(other, a.Name())
			}
		},
		Stdin:  os.Stdin,
		Stdout: buff,
		Stderr: buff,
	})
	check(t, err)

	assert.EqualValues(t, []string{"app-a"}, skipped)
	assert.EqualValues(t, []string{}, other)
}

func TestBuildBranch(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteShellScript("app-a/build.sh", "echo built app-a"))
	check(t, repo.WritePowershellScript("app-a/build.ps1", "write-host built app-a"))
	check(t, repo.Commit("first"))

	check(t, repo.SwitchToBranch("feature"))

	check(t, repo.InitModule("app-b"))
	check(t, repo.WriteShellScript("app-b/build.sh", "echo built app-b"))
	check(t, repo.WritePowershellScript("app-b/build.ps1", "write-host built app-b"))
	check(t, repo.Commit("second"))

	buff := new(bytes.Buffer)
	_, err := NewWorld(t, ".tmp/repo").System.BuildBranch("master", NoFilter, stdTestCmdOptions(buff))
	check(t, err)

	assert.Equal(t, "built app-a\n", buff.String())

	buff = new(bytes.Buffer)
	_, err = NewWorld(t, ".tmp/repo").System.BuildBranch("feature", NoFilter, stdTestCmdOptions(buff))
	check(t, err)

	assert.Equal(t, "built app-a\nbuilt app-b\n", buff.String())
}

func TestBuildDiff(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteShellScript("app-a/build.sh", "echo built app-a"))
	check(t, repo.WritePowershellScript("app-a/build.ps1", "write-host built app-a"))
	check(t, repo.Commit("first"))
	c1 := repo.LastCommit

	check(t, repo.SwitchToBranch("feature"))

	check(t, repo.InitModule("app-b"))
	check(t, repo.WriteShellScript("app-b/build.sh", "echo built app-b"))
	check(t, repo.WritePowershellScript("app-b/build.ps1", "write-host built app-b"))
	check(t, repo.Commit("second"))
	c2 := repo.LastCommit

	buff := new(bytes.Buffer)
	_, err := NewWorld(t, ".tmp/repo").System.BuildDiff(c1.String(), c2.String(), stdTestCmdOptions(buff))
	check(t, err)

	assert.Equal(t, "built app-b\n", buff.String())

	buff = new(bytes.Buffer)
	_, err = NewWorld(t, ".tmp/repo").System.BuildDiff(c2.String(), c1.String(), stdTestCmdOptions(buff))
	check(t, err)

	assert.Equal(t, "", buff.String())
}

func TestBuildPr(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteShellScript("app-a/build.sh", "echo built app-a"))
	check(t, repo.WritePowershellScript("app-a/build.ps1", "write-host built app-a"))
	check(t, repo.Commit("first"))

	check(t, repo.SwitchToBranch("feature"))

	check(t, repo.InitModule("app-b"))
	check(t, repo.WriteShellScript("app-b/build.sh", "echo built app-b"))
	check(t, repo.WritePowershellScript("app-b/build.ps1", "write-host built app-b"))
	check(t, repo.Commit("second"))

	buff := new(bytes.Buffer)
	_, err := NewWorld(t, ".tmp/repo").System.BuildPr("feature", "master", stdTestCmdOptions(buff))
	check(t, err)

	assert.Equal(t, "built app-b\n", buff.String())

	buff = new(bytes.Buffer)
	_, err = NewWorld(t, ".tmp/repo").System.BuildPr("master", "feature", stdTestCmdOptions(buff))
	check(t, err)

	assert.Equal(t, "", buff.String())
}

func TestBuildWorkspaceWithNameFilter(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteShellScript("app-a/build.sh", "echo built app-a"))
	check(t, repo.WritePowershellScript("app-a/build.ps1", "write-host built app-a"))
	check(t, repo.Commit("first"))

	check(t, repo.InitModule("app-b"))
	check(t, repo.WriteShellScript("app-b/build.sh", "echo built app-b"))
	check(t, repo.WritePowershellScript("app-b/build.ps1", "write-host built app-b"))

	buff := new(bytes.Buffer)
	_, err := NewWorld(t, ".tmp/repo").System.BuildWorkspace(&FilterOptions{Name: "app-a", Fuzzy: true}, stdTestCmdOptions(buff))
	check(t, err)

	assert.Equal(t, "built app-a\n", buff.String())
}

func TestBuildWorkspaceWithMultipleNameFilters(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteShellScript("app-a/build.sh", "echo built app-a"))
	check(t, repo.WritePowershellScript("app-a/build.ps1", "write-host built app-a"))
	check(t, repo.Commit("first"))

	check(t, repo.InitModule("app-b"))
	check(t, repo.WriteShellScript("app-b/build.sh", "echo built app-b"))
	check(t, repo.WritePowershellScript("app-b/build.ps1", "write-host built app-b"))

	check(t, repo.InitModule("app-c"))
	check(t, repo.WriteShellScript("app-c/build.sh", "echo built app-c"))
	check(t, repo.WritePowershellScript("app-c/build.ps1", "write-host built app-c"))

	buff := new(bytes.Buffer)
	_, err := NewWorld(t, ".tmp/repo").System.BuildWorkspace(&FilterOptions{Name: "app-a,app-c", Fuzzy: true}, stdTestCmdOptions(buff))
	check(t, err)

	assert.Equal(t, "built app-a\nbuilt app-c\n", buff.String())
}

func TestBuildWorkspaceWithNameFiltersMatchingSameModule(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteShellScript("app-a/build.sh", "echo built app-a"))
	check(t, repo.WritePowershellScript("app-a/build.ps1", "write-host built app-a"))
	check(t, repo.Commit("first"))

	check(t, repo.InitModule("app-b"))
	check(t, repo.WriteShellScript("app-b/build.sh", "echo built app-b"))
	check(t, repo.WritePowershellScript("app-b/build.ps1", "write-host built app-b"))

	buff := new(bytes.Buffer)
	_, err := NewWorld(t, ".tmp/repo").System.BuildWorkspace(&FilterOptions{Name: "app-a,app-a", Fuzzy: true}, stdTestCmdOptions(buff))
	check(t, err)

	assert.Equal(t, "built app-a\n", buff.String())
}

func TestBuildWorkspaceWithNameFilterThatDoesNotMatchAnyModule(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteShellScript("app-a/build.sh", "echo built app-a"))
	check(t, repo.WritePowershellScript("app-a/build.ps1", "write-host built app-a"))
	check(t, repo.Commit("first"))

	check(t, repo.InitModule("app-b"))
	check(t, repo.WriteShellScript("app-b/build.sh", "echo built app-b"))
	check(t, repo.WritePowershellScript("app-b/build.ps1", "write-host built app-b"))

	buff := new(bytes.Buffer)
	_, err := NewWorld(t, ".tmp/repo").System.BuildWorkspace(&FilterOptions{Name: "app-c", Fuzzy: true}, stdTestCmdOptions(buff))
	check(t, err)

	assert.Equal(t, "", buff.String())
}

func TestBuildWorkspace(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteShellScript("app-a/build.sh", "echo built app-a"))
	check(t, repo.WritePowershellScript("app-a/build.ps1", "write-host built app-a"))
	check(t, repo.Commit("first"))

	check(t, repo.InitModule("app-b"))
	check(t, repo.WriteShellScript("app-b/build.sh", "echo built app-b"))
	check(t, repo.WritePowershellScript("app-b/build.ps1", "write-host built app-b"))

	buff := new(bytes.Buffer)
	_, err := NewWorld(t, ".tmp/repo").System.BuildWorkspace(NoFilter, stdTestCmdOptions(buff))
	check(t, err)

	assert.Equal(t, "built app-a\nbuilt app-b\n", buff.String())
}
func TestBuildWorkspaceChanges(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteShellScript("app-a/build.sh", "echo built app-a"))
	check(t, repo.WritePowershellScript("app-a/build.ps1", "write-host built app-a"))
	check(t, repo.Commit("first"))

	check(t, repo.InitModule("app-b"))
	check(t, repo.WriteShellScript("app-b/build.sh", "echo built app-b"))
	check(t, repo.WritePowershellScript("app-b/build.ps1", "write-host built app-b"))

	buff := new(bytes.Buffer)
	_, err := NewWorld(t, ".tmp/repo").System.BuildWorkspaceChanges(stdTestCmdOptions(buff))
	check(t, err)

	assert.Equal(t, "built app-b\n", buff.String())
}

func TestBuildBranchForManifestFailure(t *testing.T) {
	clean()
	NewTestRepo(t, ".tmp/repo")

	w := NewWorld(t, ".tmp/repo")
	w.ManifestBuilder.Interceptor.Config("ByBranch").Return(nil, errors.New("doh"))
	_, err := w.System.BuildBranch("master", NoFilter, stdTestCmdOptions(nil))

	assert.EqualError(t, err, "doh")
}

func TestBuildPrForManifestFailure(t *testing.T) {
	clean()
	NewTestRepo(t, ".tmp/repo")

	w := NewWorld(t, ".tmp/repo")
	w.ManifestBuilder.Interceptor.Config("ByPr").Return(nil, errors.New("doh"))
	_, err := w.System.BuildPr("feature", "master", stdTestCmdOptions(nil))

	assert.EqualError(t, err, "doh")
}

func TestBuildDiffForManifestFailure(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))
	c := repo.LastCommit.String()

	w := NewWorld(t, ".tmp/repo")
	w.ManifestBuilder.Interceptor.Config("ByDiff").Return(nil, errors.New("doh"))
	_, err := w.System.BuildDiff(c, c, stdTestCmdOptions(nil))

	assert.EqualError(t, err, "doh")
}

func TestBuildCurrentBranchManifestFailure(t *testing.T) {
	clean()
	NewTestRepo(t, ".tmp/repo")

	w := NewWorld(t, ".tmp/repo")
	w.ManifestBuilder.Interceptor.Config("ByCurrentBranch").Return(nil, errors.New("doh"))
	_, err := w.System.BuildCurrentBranch(NoFilter, stdTestCmdOptions(nil))

	assert.EqualError(t, err, "doh")
}
func TestBuildCommitForManifestFailure(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))
	c := repo.LastCommit.String()

	w := NewWorld(t, ".tmp/repo")
	w.ManifestBuilder.Interceptor.Config("ByCommit").Return(nil, errors.New("doh"))
	_, err := w.System.BuildCommit(c, NoFilter, stdTestCmdOptions(nil))

	assert.EqualError(t, err, "doh")
}

func TestBuildWorkspaceForManifestFailure(t *testing.T) {
	clean()
	NewTestRepo(t, ".tmp/repo")

	w := NewWorld(t, ".tmp/repo")
	w.ManifestBuilder.Interceptor.Config("ByWorkspace").Return(nil, errors.New("doh"))
	_, err := w.System.BuildWorkspace(NoFilter, stdTestCmdOptions(nil))

	assert.EqualError(t, err, "doh")
}

func TestBuildWorkspaceChangesForManifestFailure(t *testing.T) {
	clean()
	NewTestRepo(t, ".tmp/repo")

	w := NewWorld(t, ".tmp/repo")
	w.ManifestBuilder.Interceptor.Config("ByWorkspaceChanges").Return(nil, errors.New("doh"))
	_, err := w.System.BuildWorkspaceChanges(stdTestCmdOptions(nil))

	assert.EqualError(t, err, "doh")
}
func TestDirtyWorkingDir(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteContent("app-a/foo", "a"))
	check(t, repo.WriteShellScript("app-a/build.sh", "echo built app-a"))
	check(t, repo.WritePowershellScript("app-a/build.ps1", "write-host built app-a"))
	check(t, repo.Commit("first"))

	check(t, repo.WriteContent("app-a/foo", "b"))

	buff := new(bytes.Buffer)
	_, err := NewWorld(t, ".tmp/repo").System.BuildCurrentBranch(NoFilter, stdTestCmdOptions(buff))
	assert.Error(t, err)
	assert.Equal(t, msgDirtyWorkingDir, err.Error())
	assert.Equal(t, ErrClassUser, (err.(*e.E)).Class())
}

func TestBuildEnvironment(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModuleWithOptions("app-a", &Spec{
		Name: "app-a",
		Build: map[string]*Cmd{
			"linux":   {Cmd: "./build.sh"},
			"darwin":  {Cmd: "./build.sh"},
			"windows": {Cmd: "powershell", Args: []string{"-ExecutionPolicy", "Bypass", "-File", ".\\build.ps1"}},
		},
		Properties: map[string]interface{}{"foo": "bar"},
	}))

	check(t, repo.WriteShellScript("app-a/build.sh", "echo $MBT_BUILD_COMMIT-$MBT_MODULE_VERSION-$MBT_MODULE_NAME-$MBT_MODULE_PATH-$MBT_REPO_PATH-$MBT_MODULE_PROPERTY_FOO"))
	check(t, repo.WritePowershellScript("app-a/build.ps1", "write-host $Env:MBT_BUILD_COMMIT-$Env:MBT_MODULE_VERSION-$Env:MBT_MODULE_NAME-$Env:MBT_MODULE_PATH-$Env:MBT_REPO_PATH-$Env:MBT_MODULE_PROPERTY_FOO"))
	check(t, repo.Commit("first"))

	m, err := NewWorld(t, ".tmp/repo").System.ManifestByCurrentBranch()
	check(t, err)

	buff := new(bytes.Buffer)
	_, err = NewWorld(t, ".tmp/repo").System.BuildCurrentBranch(NoFilter, stdTestCmdOptions(buff))
	check(t, err)

	expectedRepoPath, err := filepath.Abs(".tmp/repo")
	check(t, err)
	out := buff.String()
	assert.Equal(t, fmt.Sprintf("%s-%s-%s-%s-%s-%s\n", m.Sha, m.Modules[0].Version(), m.Modules[0].Name(), m.Modules[0].Path(), expectedRepoPath, m.Modules[0].Properties()["foo"]), out)
}

func TestDefaultBuild(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	check(t, repo.InitModuleWithOptions("app-a", &Spec{
		Name:  "app-a",
		Build: map[string]*Cmd{"default": {Cmd: "echo", Args: []string{"hello"}}},
	}))
	check(t, repo.Commit("first"))

	_, err := NewWorld(t, ".tmp/repo").System.ManifestByCurrentBranch()
	check(t, err)

	buff := new(bytes.Buffer)
	_, err = NewWorld(t, ".tmp/repo").System.BuildCurrentBranch(NoFilter, stdTestCmdOptions(buff))
	check(t, err)

	assert.Equal(t, "hello\n", buff.String())
}

func TestDefaultBuildOverride(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModuleWithOptions("app-a", &Spec{
		Name: "app-a",
		Build: map[string]*Cmd{
			"default": {Cmd: "doh"},
			"windows": {Cmd: "powershell", Args: []string{"-ExecutionPolicy", "Bypass", "-File", ".\\build.ps1"}},
			"darwin":  {Cmd: "./build.sh"},
			"linux":   {Cmd: "./build.sh"},
		},
	}))

	check(t, repo.WriteShellScript("app-a/build.sh", "echo foo"))
	check(t, repo.WritePowershellScript("app-a/build.ps1", "write-host bar"))
	check(t, repo.Commit("first"))

	_, err := NewWorld(t, ".tmp/repo").System.ManifestByCurrentBranch()
	check(t, err)

	buff := new(bytes.Buffer)
	_, err = NewWorld(t, ".tmp/repo").System.BuildCurrentBranch(NoFilter, stdTestCmdOptions(buff))
	check(t, err)

	out := buff.String()
	if runtime.GOOS == "windows" {
		assert.Equal(t, "bar\n", out)
	} else {
		assert.Equal(t, "foo\n", out)
	}
}

func TestBuildEnvironmentForAbsPath(t *testing.T) {
	clean()

	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModuleWithOptions("app-a", &Spec{
		Name: "app-a",
		Build: map[string]*Cmd{
			"linux":   {Cmd: "./build.sh"},
			"darwin":  {Cmd: "./build.sh"},
			"windows": {Cmd: "powershell", Args: []string{"-ExecutionPolicy", "Bypass", "-File", ".\\build.ps1"}},
		},
		Properties: map[string]interface{}{"foo": "bar"},
	}))

	check(t, repo.WriteShellScript("app-a/build.sh", "echo $MBT_BUILD_COMMIT-$MBT_MODULE_VERSION-$MBT_MODULE_NAME-$MBT_MODULE_PATH-$MBT_REPO_PATH-$MBT_MODULE_PROPERTY_FOO"))
	check(t, repo.WritePowershellScript("app-a/build.ps1", "write-host $Env:MBT_BUILD_COMMIT-$Env:MBT_MODULE_VERSION-$Env:MBT_MODULE_NAME-$Env:MBT_MODULE_PATH-$Env:MBT_REPO_PATH-$Env:MBT_MODULE_PROPERTY_FOO"))
	check(t, repo.Commit("first"))

	expectedRepoPath, err := filepath.Abs(".tmp/repo")
	check(t, err)
	m, err := NewWorld(t, expectedRepoPath).System.ManifestByCurrentBranch()
	check(t, err)

	buff := new(bytes.Buffer)
	_, err = NewWorld(t, ".tmp/repo").System.BuildCurrentBranch(NoFilter, stdTestCmdOptions(buff))
	check(t, err)
	out := buff.String()
	assert.Equal(t, fmt.Sprintf("%s-%s-%s-%s-%s-%s\n", m.Sha, m.Modules[0].Version(), m.Modules[0].Name(), m.Modules[0].Path(), expectedRepoPath, m.Modules[0].Properties()["foo"]), out)
}

func TestBadSha(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))

	_, err := NewWorld(t, ".tmp/repo").System.BuildCommit("a", NoFilter, stdTestCmdOptions(nil))

	assert.EqualError(t, err, fmt.Sprintf(msgInvalidSha, "a"))
	assert.EqualError(t, (err.(*e.E)).InnerError(), "encoding/hex: odd length hex string")
	assert.Equal(t, ErrClassUser, (err.(*e.E)).Class())
}

func TestMissingSha(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))

	sha := "22221c5e56794a2af5f59f94512df4c669c77a49"
	_, err := NewWorld(t, ".tmp/repo").System.BuildCommit(sha, NoFilter, stdTestCmdOptions(nil))

	assert.EqualError(t, err, fmt.Sprintf(msgCommitShaNotFound, sha))
	assert.EqualError(t, (err.(*e.E)).InnerError(), "object not found - no match for id (22221c5e56794a2af5f59f94512df4c669c77a49)")
	assert.Equal(t, ErrClassUser, (err.(*e.E)).Class())
}

func TestBuildCommitContent(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteShellScript("app-a/build.sh", "echo built app-a"))
	check(t, repo.WritePowershellScript("app-a/build.ps1", "write-host built app-a"))
	check(t, repo.Commit("first"))

	check(t, repo.InitModule("app-b"))
	check(t, repo.WriteShellScript("app-b/build.sh", "echo built app-b"))
	check(t, repo.WritePowershellScript("app-b/build.ps1", "write-host built app-b"))
	check(t, repo.Commit("second"))

	buff := new(bytes.Buffer)
	_, err := NewWorld(t, ".tmp/repo").System.BuildCommitContent(repo.LastCommit.String(), stdTestCmdOptions(buff))
	check(t, err)

	assert.Equal(t, "built app-b\n", buff.String())
}

func TestBuildCommitContentForManifestFailure(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))

	w := NewWorld(t, ".tmp/repo")
	w.ManifestBuilder.Interceptor.Config("ByCommitContent").Return((*Manifest)(nil), errors.New("doh"))

	buff := new(bytes.Buffer)
	_, err := w.System.BuildCommitContent(repo.LastCommit.String(), stdTestCmdOptions(buff))

	assert.EqualError(t, err, "doh")
}

func TestRestorationOfPristineWorkspace(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteShellScript("app-a/build.sh", "echo built app-a"))
	check(t, repo.WritePowershellScript("app-a/build.ps1", "write-host built app-a"))
	check(t, repo.WriteContent("app-a/foo", "bar1"))
	check(t, repo.Commit("first"))

	check(t, repo.SwitchToBranch("feature"))
	check(t, repo.AppendContent("app-a/foo", "bar2"))
	check(t, repo.AppendContent("app-a/foo2", "bar1"))
	check(t, repo.Commit("second"))

	check(t, repo.SwitchToBranch("master"))
	check(t, repo.AppendContent("app-a/foo", "bar3"))
	check(t, repo.AppendContent("app-a/foo2", "bar2"))
	check(t, repo.Commit("third"))

	w := NewWorld(t, ".tmp/repo")
	buff := new(bytes.Buffer)
	_, err := w.System.BuildBranch("feature", NoFilter, stdTestCmdOptions(buff))
	check(t, err)

	idx, err := repo.Repo.Index()
	check(t, err)

	diff, err := repo.Repo.DiffIndexToWorkdir(idx, &git.DiffOptions{
		Flags: git.DiffIncludeUntracked | git.DiffRecurseUntracked,
	})
	check(t, err)

	numDeltas, err := diff.NumDeltas()
	check(t, err)
	assert.Equal(t, 0, numDeltas)

	head, err := repo.Repo.Head()
	check(t, err)
	assert.Equal(t, "refs/heads/master", head.Name())
}

func TestBuildingDetachedHead(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteShellScript("app-a/build.sh", "echo built app-a"))
	check(t, repo.WritePowershellScript("app-a/build.ps1", "write-host built app-a"))
	check(t, repo.WriteContent("app-a/foo", "bar1"))
	check(t, repo.Commit("first"))
	first := repo.LastCommit

	check(t, repo.SwitchToBranch("feature"))
	check(t, repo.AppendContent("app-a/foo", "bar2"))
	check(t, repo.AppendContent("app-a/foo2", "bar1"))
	check(t, repo.Commit("second"))

	check(t, repo.SwitchToBranch("master"))
	check(t, repo.AppendContent("app-a/foo", "bar3"))
	check(t, repo.AppendContent("app-a/foo2", "bar2"))
	check(t, repo.Commit("third"))

	check(t, repo.CheckoutAndDetach(first.String()))

	w := NewWorld(t, ".tmp/repo")
	buff := new(bytes.Buffer)
	_, err := w.System.BuildBranch("feature", NoFilter, stdTestCmdOptions(buff))
	check(t, err)

	// Ensure that we don't touch this workspace
	idx, err := repo.Repo.Index()
	check(t, err)

	diff, err := repo.Repo.DiffIndexToWorkdir(idx, &git.DiffOptions{
		Flags: git.DiffIncludeUntracked | git.DiffRecurseUntracked,
	})
	check(t, err)

	numDeltas, err := diff.NumDeltas()
	check(t, err)
	assert.Equal(t, 0, numDeltas)

	detached, err := repo.Repo.IsHeadDetached()
	check(t, err)
	assert.True(t, detached)
}

func TestRestorationOnBuildFailure(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteShellScript("app-a/build.sh", "exit 1"))
	check(t, repo.WritePowershellScript("app-a/build.ps1", "throw \"foo\""))
	check(t, repo.Commit("first"))

	check(t, repo.SwitchToBranch("feature"))
	check(t, repo.AppendContent("app-a/foo", "bar2"))
	check(t, repo.AppendContent("app-a/foo2", "bar1"))
	check(t, repo.Commit("second"))

	w := NewWorld(t, ".tmp/repo")
	buff := new(bytes.Buffer)
	_, err := w.System.BuildBranch("feature", NoFilter, stdTestCmdOptions(buff))
	assert.EqualError(t, err, fmt.Sprintf(msgFailedBuild, "app-a"))

	idx, err := repo.Repo.Index()
	check(t, err)

	diff, err := repo.Repo.DiffIndexToWorkdir(idx, &git.DiffOptions{
		Flags: git.DiffIncludeUntracked | git.DiffRecurseUntracked,
	})
	check(t, err)

	numDeltas, err := diff.NumDeltas()
	check(t, err)
	assert.Equal(t, 0, numDeltas)
}
