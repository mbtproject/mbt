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
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCmdExecution(t *testing.T) {
	clean()
	r := NewTestRepo(t, ".tmp/repo")

	r.InitModuleWithOptions("app-a", &Spec{
		Name: "app-a",
		Commands: map[string]*UserCmd{
			"echo": {Cmd: "echo", Args: []string{"hello"}},
		},
	})

	w := NewWorld(t, ".tmp/repo")

	// Test new workspace
	buff := new(bytes.Buffer)
	result, err := w.System.RunInWorkspace("echo", NoFilter, stdTestCmdOptions(buff))
	check(t, err)

	assert.Len(t, result.Completed, 1)
	assert.Len(t, result.Skipped, 0)
	assert.Len(t, result.Failures, 0)
	assert.Equal(t, "app-a", result.Completed[0].Name())
	assert.Equal(t, "hello\n", buff.String())

	// Test new workspace diff
	buff = new(bytes.Buffer)
	result, err = w.System.RunInWorkspaceChanges("echo", stdTestCmdOptions(buff))

	check(t, err)
	assert.Len(t, result.Completed, 1)
	assert.Len(t, result.Skipped, 0)
	assert.Len(t, result.Failures, 0)
	assert.Equal(t, "app-a", result.Completed[0].Name())
	assert.Equal(t, "hello\n", buff.String())
}
func TestRunInDirtyWorkingDir(t *testing.T) {
	clean()
	r := NewTestRepo(t, ".tmp/repo")

	r.InitModuleWithOptions("app-a", &Spec{
		Name: "app-a",
		Commands: map[string]*UserCmd{
			"echo": {Cmd: "echo", Args: []string{"hello"}},
		},
	})

	w := NewWorld(t, ".tmp/repo")

	buff := new(bytes.Buffer)
	result, err := w.System.RunInCurrentBranch("echo", NoFilter, stdTestCmdOptions(buff))

	assert.Nil(t, result)
	assert.EqualError(t, err, msgDirtyWorkingDir)
}

func TestRunInBranch(t *testing.T) {
	clean()
	r := NewTestRepo(t, ".tmp/repo")

	r.InitModuleWithOptions("app-a", &Spec{
		Name: "app-a",
		Commands: map[string]*UserCmd{
			"echo": {Cmd: "echo", Args: []string{"hello-master"}},
		},
	})

	r.Commit("first")

	r.SwitchToBranch("feature")
	r.Remove("app-a")
	r.InitModuleWithOptions("app-a", &Spec{
		Name: "app-a",
		Commands: map[string]*UserCmd{
			"echo": {Cmd: "echo", Args: []string{"hello-feature"}},
		},
	})

	r.Commit("second")

	w := NewWorld(t, ".tmp/repo")

	buff := new(bytes.Buffer)
	result, err := w.System.RunInBranch("echo", "master", NoFilter, stdTestCmdOptions(buff))

	check(t, err)
	assert.Len(t, result.Completed, 1)
	assert.Len(t, result.Skipped, 0)
	assert.Len(t, result.Failures, 0)
	assert.Equal(t, "app-a", result.Completed[0].Name())
	assert.Equal(t, "hello-master\n", buff.String())

	buff = new(bytes.Buffer)
	result, err = w.System.RunInBranch("echo", "feature", NoFilter, stdTestCmdOptions(buff))

	check(t, err)
	assert.Len(t, result.Completed, 1)
	assert.Len(t, result.Skipped, 0)
	assert.Len(t, result.Failures, 0)
	assert.Equal(t, "app-a", result.Completed[0].Name())
	assert.Equal(t, "hello-feature\n", buff.String())
}

func TestRunInPr(t *testing.T) {
	clean()
	r := NewTestRepo(t, ".tmp/repo")

	r.InitModuleWithOptions("app-a", &Spec{
		Name: "app-a",
		Commands: map[string]*UserCmd{
			"echo": {Cmd: "echo", Args: []string{"hello-app-a"}},
		},
	})

	r.Commit("first")

	r.SwitchToBranch("feature")
	r.InitModuleWithOptions("app-b", &Spec{
		Name: "app-b",
		Commands: map[string]*UserCmd{
			"echo": {Cmd: "echo", Args: []string{"hello-app-b"}},
		},
	})

	r.Commit("second")

	w := NewWorld(t, ".tmp/repo")

	buff := new(bytes.Buffer)
	result, err := w.System.RunInPr("echo", "feature", "master", stdTestCmdOptions(buff))

	check(t, err)
	assert.Len(t, result.Completed, 1)
	assert.Len(t, result.Skipped, 0)
	assert.Len(t, result.Failures, 0)
	assert.Equal(t, "app-b", result.Completed[0].Name())
	assert.Equal(t, "hello-app-b\n", buff.String())

	buff = new(bytes.Buffer)
	result, err = w.System.RunInPr("echo", "master", "feature", stdTestCmdOptions(buff))

	check(t, err)
	assert.Len(t, result.Completed, 0)
	assert.Len(t, result.Skipped, 0)
	assert.Len(t, result.Failures, 0)
	assert.Equal(t, "", buff.String())
}

func TestRunInDiff(t *testing.T) {
	clean()
	r := NewTestRepo(t, ".tmp/repo")

	r.InitModuleWithOptions("app-a", &Spec{
		Name: "app-a",
		Commands: map[string]*UserCmd{
			"echo": {Cmd: "echo", Args: []string{"hello-app-a"}},
		},
	})

	r.Commit("first")
	c1 := r.LastCommit

	r.InitModuleWithOptions("app-b", &Spec{
		Name: "app-b",
		Commands: map[string]*UserCmd{
			"echo": {Cmd: "echo", Args: []string{"hello-app-b"}},
		},
	})

	r.Commit("second")
	c2 := r.LastCommit

	w := NewWorld(t, ".tmp/repo")

	buff := new(bytes.Buffer)
	result, err := w.System.RunInDiff("echo", c1.String(), c2.String(), stdTestCmdOptions(buff))

	check(t, err)
	assert.Len(t, result.Completed, 1)
	assert.Len(t, result.Skipped, 0)
	assert.Len(t, result.Failures, 0)
	assert.Equal(t, "app-b", result.Completed[0].Name())
	assert.Equal(t, "hello-app-b\n", buff.String())

	buff = new(bytes.Buffer)
	result, err = w.System.RunInDiff("echo", c2.String(), c1.String(), stdTestCmdOptions(buff))

	check(t, err)
	assert.Len(t, result.Completed, 0)
	assert.Len(t, result.Skipped, 0)
	assert.Len(t, result.Failures, 0)
	assert.Equal(t, "", buff.String())
}

func TestRunInCurrentBranch(t *testing.T) {
	clean()
	r := NewTestRepo(t, ".tmp/repo")

	r.InitModuleWithOptions("app-a", &Spec{
		Name: "app-a",
		Commands: map[string]*UserCmd{
			"echo": {Cmd: "echo", Args: []string{"hello"}},
		},
	})

	r.Commit("first")

	w := NewWorld(t, ".tmp/repo")
	buff := new(bytes.Buffer)
	result, err := w.System.RunInCurrentBranch("echo", NoFilter, stdTestCmdOptions(buff))

	check(t, err)
	assert.Len(t, result.Completed, 1)
	assert.Len(t, result.Skipped, 0)
	assert.Len(t, result.Failures, 0)
	assert.Equal(t, "app-a", result.Completed[0].Name())
	assert.Equal(t, "hello\n", buff.String())
}

func TestRunInCommit(t *testing.T) {
	clean()
	r := NewTestRepo(t, ".tmp/repo")

	r.InitModuleWithOptions("app-a", &Spec{
		Name: "app-a",
		Commands: map[string]*UserCmd{
			"echo": {Cmd: "echo", Args: []string{"hello-c1"}},
		},
	})

	r.Commit("first")
	c1 := r.LastCommit

	r.Remove("app-a")
	r.InitModuleWithOptions("app-a", &Spec{
		Name: "app-a",
		Commands: map[string]*UserCmd{
			"echo": {Cmd: "echo", Args: []string{"hello-c2"}},
		},
	})
	r.Commit("second")
	c2 := r.LastCommit

	w := NewWorld(t, ".tmp/repo")

	buff := new(bytes.Buffer)
	result, err := w.System.RunInCommit("echo", c1.String(), NoFilter, stdTestCmdOptions(buff))

	check(t, err)
	assert.Len(t, result.Completed, 1)
	assert.Len(t, result.Skipped, 0)
	assert.Len(t, result.Failures, 0)
	assert.Equal(t, "app-a", result.Completed[0].Name())
	assert.Equal(t, "hello-c1\n", buff.String())

	buff = new(bytes.Buffer)
	result, err = w.System.RunInCommit("echo", c2.String(), NoFilter, stdTestCmdOptions(buff))

	check(t, err)
	assert.Len(t, result.Completed, 1)
	assert.Len(t, result.Skipped, 0)
	assert.Len(t, result.Failures, 0)
	assert.Equal(t, "app-a", result.Completed[0].Name())
	assert.Equal(t, "hello-c2\n", buff.String())
}

func TestRunInCommitContent(t *testing.T) {
	clean()
	r := NewTestRepo(t, ".tmp/repo")

	r.InitModuleWithOptions("app-a", &Spec{
		Name: "app-a",
		Commands: map[string]*UserCmd{
			"echo": {Cmd: "echo", Args: []string{"hello-app-a"}},
		},
	})

	r.Commit("first")

	r.InitModuleWithOptions("app-b", &Spec{
		Name: "app-b",
		Commands: map[string]*UserCmd{
			"echo": {Cmd: "echo", Args: []string{"hello-app-b"}},
		},
	})

	r.Commit("second")

	w := NewWorld(t, ".tmp/repo")

	buff := new(bytes.Buffer)
	result, err := w.System.RunInCommitContent("echo", r.LastCommit.String(), stdTestCmdOptions(buff))

	check(t, err)
	assert.Len(t, result.Completed, 1)
	assert.Len(t, result.Skipped, 0)
	assert.Len(t, result.Failures, 0)
	assert.Equal(t, "app-b", result.Completed[0].Name())
	assert.Equal(t, "hello-app-b\n", buff.String())
}

func TestOSConstraints(t *testing.T) {
	clean()
	r := NewTestRepo(t, ".tmp/repo")

	r.InitModuleWithOptions("unix-app", &Spec{
		Name: "unix-app",
		Commands: map[string]*UserCmd{
			"echo": {Cmd: "./script.sh", OS: []string{"linux", "darwin"}},
		},
	})
	r.WriteShellScript("unix-app/script.sh", "echo hello-unix-app")

	r.InitModuleWithOptions("windows-app", &Spec{
		Name: "windows-app",
		Commands: map[string]*UserCmd{
			"echo": {Cmd: "powershell", Args: []string{"-ExecutionPolicy", "Bypass", "-File", ".\\script.ps1"}, OS: []string{"windows"}},
		},
	})
	r.WritePowershellScript("windows-app/script.ps1", "echo hello-windows-app")
	r.Commit("first")

	w := NewWorld(t, ".tmp/repo")

	buff := new(bytes.Buffer)
	result, err := w.System.RunInCurrentBranch("echo", NoFilter, stdTestCmdOptions(buff))

	check(t, err)
	switch runtime.GOOS {
	case "linux", "darwin":
		assert.Len(t, result.Completed, 1)
		assert.Len(t, result.Skipped, 1)
		assert.Len(t, result.Failures, 0)
		assert.Equal(t, "unix-app", result.Completed[0].Name())
		assert.Equal(t, "windows-app", result.Skipped[0].Name())
		assert.Equal(t, "hello-unix-app\n", buff.String())
	case "windows":
		assert.Len(t, result.Completed, 1)
		assert.Len(t, result.Skipped, 1)
		assert.Len(t, result.Failures, 0)
		assert.Equal(t, "windows-app", result.Completed[0].Name())
		assert.Equal(t, "unix-app", result.Skipped[0].Name())
		assert.Equal(t, "hello-windows-app\r\n", buff.String())
	}
}

func TestFailingCommands(t *testing.T) {
	clean()
	r := NewTestRepo(t, ".tmp/repo")

	r.InitModuleWithOptions("app-a", &Spec{
		Name: "app-a",
		Commands: map[string]*UserCmd{
			"echo": {Cmd: "echo", Args: []string{"app-a"}},
		},
	})
	r.InitModuleWithOptions("app-b", &Spec{
		Name: "app-b",
		Commands: map[string]*UserCmd{
			"echo": {Cmd: "bad_command", Args: []string{"app-b"}},
		},
	})
	r.InitModuleWithOptions("app-c", &Spec{
		Name: "app-c",
		Commands: map[string]*UserCmd{
			"echo": {Cmd: "echo", Args: []string{"app-c"}},
		},
	})

	r.Commit("first")

	w := NewWorld(t, ".tmp/repo")

	buff := new(bytes.Buffer)
	result, err := w.System.RunInCurrentBranch("echo", NoFilter, stdTestCmdOptions(buff))

	check(t, err)
	assert.Len(t, result.Completed, 2)
	assert.Len(t, result.Skipped, 0)
	assert.Len(t, result.Failures, 1)
	assert.Equal(t, "app-a", result.Completed[0].Name())
	assert.Equal(t, "app-c", result.Completed[1].Name())
	assert.Equal(t, "app-b", result.Failures[0].Module.Name())
	assert.Equal(t, "app-a\napp-c\n", buff.String())
}

func TestFailFast(t *testing.T) {
	clean()
	r := NewTestRepo(t, ".tmp/repo")

	r.InitModuleWithOptions("app-a", &Spec{
		Name: "app-a",
		Commands: map[string]*UserCmd{
			"echo": {Cmd: "echo", Args: []string{"app-a"}},
		},
	})
	r.InitModuleWithOptions("app-b", &Spec{
		Name: "app-b",
		Commands: map[string]*UserCmd{
			"echo": {Cmd: "bad_command", Args: []string{"app-b"}},
		},
	})
	r.InitModuleWithOptions("app-c", &Spec{
		Name: "app-c",
		Commands: map[string]*UserCmd{
			"echo": {Cmd: "echo", Args: []string{"app-c"}},
		},
	})

	r.Commit("first")

	w := NewWorld(t, ".tmp/repo")

	buff := new(bytes.Buffer)
	options := stdTestCmdOptions(buff)
	options.FailFast = true
	result, err := w.System.RunInCurrentBranch("echo", NoFilter, options)

	check(t, err)
	assert.Len(t, result.Completed, 1)
	assert.Len(t, result.Skipped, 1)
	assert.Len(t, result.Failures, 1)
	assert.Equal(t, "app-a", result.Completed[0].Name())
	assert.Equal(t, "app-b", result.Failures[0].Module.Name())
	assert.Equal(t, "app-c", result.Skipped[0].Name())
	assert.Equal(t, "app-a\n", buff.String())
}
