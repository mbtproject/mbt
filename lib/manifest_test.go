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
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/mbtproject/mbt/e"
	"github.com/stretchr/testify/assert"
)

func TestSingleModDir(t *testing.T) {
	clean()

	repo := NewTestRepo(t, ".tmp/repo")
	err := repo.InitModule("app-a")
	check(t, err)
	err = repo.Commit("first")
	check(t, err)

	m, err := NewWorld(t, ".tmp/repo").System.ManifestByBranch("master")
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, "app-a", m.Modules[0].Name())
}

func TestNonModContent(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	err := repo.InitModule("app-a")
	check(t, err)
	err = repo.WriteContent("content/index.html", "hello")
	check(t, err)
	err = repo.Commit("first")
	check(t, err)

	m, err := NewWorld(t, ".tmp/repo").System.ManifestByBranch("master")
	check(t, err)

	assert.Len(t, m.Modules, 1)
}

func TestNestedAppDir(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("a/b/c/app-a"))
	check(t, repo.Commit("first"))

	m, err := NewWorld(t, ".tmp/repo").System.ManifestByBranch("master")
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, "a/b/c/app-a", m.Modules[0].Path())
}

func TestModsDirInModDir(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.InitModule("app-a/app-b"))
	check(t, repo.Commit("first"))

	m, err := NewWorld(t, ".tmp/repo").System.ManifestByBranch("master")
	check(t, err)

	assert.Len(t, m.Modules, 2)
	assert.Equal(t, "app-a", m.Modules[0].Name())
	assert.Equal(t, "app-a", m.Modules[0].Path())
	assert.Equal(t, "app-b", m.Modules[1].Name())
	assert.Equal(t, "app-a/app-b", m.Modules[1].Path())
}

func TestEmptyRepo(t *testing.T) {
	clean()

	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))

	m, err := NewWorld(t, ".tmp/repo").System.ManifestByBranch("master")
	check(t, err)

	assert.Len(t, m.Modules, 0)
}

func TestDiffingTwoBranches(t *testing.T) {
	clean()

	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))
	masterTip := repo.LastCommit

	check(t, repo.SwitchToBranch("feature"))
	check(t, repo.InitModule("app-b"))
	check(t, repo.Commit("second"))

	featureTip := repo.LastCommit

	m, err := NewWorld(t, ".tmp/repo").System.ManifestByPr("feature", "master")
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, featureTip.String(), m.Sha)
	assert.Equal(t, "app-b", m.Modules[0].Name())

	m, err = NewWorld(t, ".tmp/repo").System.ManifestByPr("master", "feature")
	check(t, err)

	assert.Len(t, m.Modules, 0)
	assert.Equal(t, masterTip.String(), m.Sha)
}

func TestDiffingTwoProgressedBranches(t *testing.T) {
	clean()

	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))
	check(t, repo.SwitchToBranch("feature"))
	check(t, repo.InitModule("app-b"))
	check(t, repo.Commit("second"))
	check(t, repo.SwitchToBranch("master"))
	check(t, repo.InitModule("app-c"))
	check(t, repo.Commit("third"))

	m, err := NewWorld(t, ".tmp/repo").System.ManifestByPr("feature", "master")
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, "app-b", m.Modules[0].Name())

	m, err = NewWorld(t, ".tmp/repo").System.ManifestByPr("master", "feature")
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, "app-c", m.Modules[0].Name())
}

func TestDiffingWithMultipleChangesToSameMod(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))
	check(t, repo.SwitchToBranch("feature"))
	check(t, repo.WriteContent("app-a/file1", "hello"))
	check(t, repo.Commit("second"))
	check(t, repo.WriteContent("app-a/file2", "world"))
	check(t, repo.Commit("third"))

	m, err := NewWorld(t, ".tmp/repo").System.ManifestByPr("feature", "master")
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, "app-a", m.Modules[0].Name())
}

func TestDiffingForDeletes(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteContent("app-a/file1", "hello world"))
	check(t, repo.Commit("first"))
	check(t, repo.SwitchToBranch("feature"))
	check(t, repo.Remove("app-a/file1"))
	check(t, repo.Commit("second"))

	m, err := NewWorld(t, ".tmp/repo").System.ManifestByPr("feature", "master")
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, "app-a", m.Modules[0].Name())
}

func TestDiffingForRenames(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteContent("app-a/file1", "hello world"))
	check(t, repo.Commit("first"))
	check(t, repo.SwitchToBranch("feature"))
	check(t, repo.Rename("app-a/file1", "app-a/file2"))
	check(t, repo.Commit("second"))

	m, err := NewWorld(t, ".tmp/repo").System.ManifestByPr("feature", "master")
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, "app-a", m.Modules[0].Name())
}

func TestModuleOnRoot(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModuleWithOptions("", &Spec{Name: "root-app"}))
	check(t, repo.Commit("first"))

	m, err := NewWorld(t, ".tmp/repo").System.ManifestByBranch("master")
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, "root-app", m.Modules[0].Name())
	assert.Equal(t, "", m.Modules[0].Path())
	assert.Equal(t, repo.LastCommit.String(), m.Modules[0].Version())
}

func TestModuleOnRootWhenDiffing(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	// Test a change in the root dir
	check(t, repo.InitModuleWithOptions("", &Spec{Name: "root-app"}))
	check(t, repo.Commit("first"))
	first := repo.LastCommit.String()

	check(t, repo.WriteContent("foo", "bar"))
	check(t, repo.Commit("second"))
	second := repo.LastCommit.String()

	world := NewWorld(t, ".tmp/repo")
	m, err := world.System.ManifestByDiff(first, second)
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, "root-app", m.Modules[0].Name())
	assert.Equal(t, "", m.Modules[0].Path())
	assert.Equal(t, second, m.Modules[0].Version())

	// Test a change in a nested dir
	check(t, repo.WriteContent("dir/foo", "bar"))
	check(t, repo.Commit("third"))
	third := repo.LastCommit.String()

	m, err = world.System.ManifestByDiff(second, third)
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, "root-app", m.Modules[0].Name())
	assert.Equal(t, "", m.Modules[0].Path())
	assert.Equal(t, third, m.Modules[0].Version())

	// Test an empty diff
	m, err = world.System.ManifestByDiff(third, third)
	check(t, err)

	assert.Len(t, m.Modules, 0)
}

func TestNestedModules(t *testing.T) {
	/*
		Nesting modules is a rare scenario.
		Although it could be useful to implement common build logic
		for a sub tree.
		Current expectation is, if a file changes in a path
		where modules are nested, all impacted modules should
		be returned in manifest.
		Consider the following repo structure

		/
		|_ .mbt.yml (root module)
		|_ foo.txt
		|_ mod-a
		    |_ .mbt.yml
		    |_ bar.txt

		Change to foo.txt should return just root module in the manifest.
		Change to bar.txt on the other hand should return both root module
		and mod-a.
	*/
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModuleWithOptions("", &Spec{Name: "root-app"}))
	check(t, repo.Commit("first"))
	first := repo.LastCommit.String()

	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("second"))
	second := repo.LastCommit.String()

	world := NewWorld(t, ".tmp/repo")
	m, err := world.System.ManifestByDiff(first, second)
	check(t, err)

	assert.Len(t, m.Modules, 2)
	assert.Equal(t, "app-a", m.Modules[0].Name())
	assert.Equal(t, "app-a", m.Modules[0].Path())
	assert.NotEqual(t, second, m.Modules[0].Version())
	assert.Equal(t, "root-app", m.Modules[1].Name())
	assert.Equal(t, "", m.Modules[1].Path())
	assert.Equal(t, second, m.Modules[1].Version())
}

func TestManifestByDiff(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.InitModule("app-b"))
	check(t, repo.Commit("first"))
	c1 := repo.LastCommit

	check(t, repo.WriteContent("app-a/foo", "hello"))
	check(t, repo.Commit("second"))
	c2 := repo.LastCommit

	m, err := NewWorld(t, ".tmp/repo").System.ManifestByDiff(c1.String(), c2.String())
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, "app-a", m.Modules[0].Name())

	m, err = NewWorld(t, ".tmp/repo").System.ManifestByDiff(c2.String(), c1.String())
	check(t, err)

	assert.Len(t, m.Modules, 0)
}

func TestManifestByHead(t *testing.T) {
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))
	check(t, repo.SwitchToBranch("feature"))

	m, err := NewWorld(t, ".tmp/repo").System.ManifestByCurrentBranch()
	check(t, err)

	assert.Equal(t, "app-a", m.Modules[0].Name())
}

func TestManifestByLocalDirForUpdates(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteContent("app-a/test.txt", "test contents"))
	check(t, repo.Commit("first"))

	m, err := NewWorld(t, ".tmp/repo").System.ManifestByWorkspaceChanges()
	check(t, err)
	expectedPath, err := filepath.Abs(".tmp/repo")
	check(t, err)

	assert.Equal(t, "local", m.Sha)
	assert.Equal(t, expectedPath, m.Dir)

	// currently no modules changed locally
	assert.Equal(t, 0, len(m.Modules))

	// change the file, expect 1 module to be returned
	check(t, repo.WriteContent("app-a/test.txt", "amended contents"))

	m, err = NewWorld(t, ".tmp/repo").System.ManifestByWorkspaceChanges()
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, "app-a", m.Modules[0].Name())
	assert.Equal(t, "local", m.Modules[0].Version())
}

func TestManifestByLocalDirForAddition(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteContent("app-a/test.txt", "test contents"))
	check(t, repo.Commit("first"))

	check(t, repo.InitModule("app-b"))
	m, err := NewWorld(t, ".tmp/repo").System.ManifestByWorkspaceChanges()
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, "app-b", m.Modules[0].Name())
	assert.Equal(t, "local", m.Modules[0].Version())
}

func TestManifestByLocalDirForConversion(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.WriteContent("app-a/test.txt", "test contents"))
	check(t, repo.Commit("first"))

	check(t, repo.InitModule("app-a"))
	m, err := NewWorld(t, ".tmp/repo").System.ManifestByWorkspaceChanges()
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, "app-a", m.Modules[0].Name())
}

func TestManifestByLocalDirForNestedModules(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.WriteContent("README.md", "test contents"))
	check(t, repo.Commit("first"))
	check(t, repo.InitModule("src/app-a"))
	check(t, repo.WriteContent("src/app-a/test.txt", "test contents"))

	m, err := NewWorld(t, ".tmp/repo").System.ManifestByWorkspaceChanges()
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, "app-a", m.Modules[0].Name())
}

func TestManifestByLocalDirForAnEmptyRepo(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))

	m, err := NewWorld(t, ".tmp/repo").System.ManifestByWorkspaceChanges()
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, "app-a", m.Modules[0].Name())
	assert.Equal(t, "local", m.Modules[0].Version())
}

func TestManifestByLocalForFilesEndingWithSpecFileName(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.WriteContent("dummy/interesting.mbt.yml", "hello"))

	m, err := NewWorld(t, ".tmp/repo").System.ManifestByWorkspace()
	check(t, err)

	assert.Len(t, m.Modules, 0)
}

func TestManifestByLocalForDirsEndingWithSpecFileName(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.WriteContent(".mbt.yml/hello", "world"))

	m, err := NewWorld(t, ".tmp/repo").System.ManifestByWorkspace()
	check(t, err)

	assert.Len(t, m.Modules, 0)
}

func TestVersionOfLocalDirManifest(t *testing.T) {
	// All modules should have the fixed version string "local" as
	// for manifest derived from local directory.
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteContent("app-a/test.txt", "test contents"))
	check(t, repo.InitModuleWithOptions("app-b", &Spec{
		Name:         "app-b",
		Dependencies: []string{"app-a"},
	}))
	check(t, repo.Commit("first"))

	m, err := NewWorld(t, ".tmp/repo").System.ManifestByWorkspace()
	check(t, err)

	expectedPath, err := filepath.Abs(".tmp/repo")
	check(t, err)

	assert.Equal(t, "local", m.Sha)
	assert.Equal(t, expectedPath, m.Dir)

	assert.Equal(t, 2, len(m.Modules))
	assert.Equal(t, "app-a", m.Modules[0].Name())
	assert.Equal(t, "local", m.Modules[0].Version())
	assert.Equal(t, "app-b", m.Modules[1].Name())
	assert.Equal(t, "local", m.Modules[1].Version())
}

func TestLocalDependencyChange(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.InitModuleWithOptions("app-b", &Spec{
		Name:         "app-b",
		Dependencies: []string{"app-a"},
	}))
	check(t, repo.InitModuleWithOptions("app-c", &Spec{
		Name:         "app-c",
		Dependencies: []string{"app-b"},
	}))
	check(t, repo.InitModuleWithOptions("app-d", &Spec{
		Name:         "app-d",
		Dependencies: []string{"app-c"},
	}))
	check(t, repo.Commit("first"))

	check(t, repo.WriteContent("app-b/foo", "bar"))

	m, err := NewWorld(t, ".tmp/repo").System.ManifestByWorkspaceChanges()
	check(t, err)

	assert.Equal(t, 3, len(m.Modules))
	assert.Equal(t, "app-b", m.Modules[0].Name())
	assert.Equal(t, "local", m.Modules[0].Version())
	assert.Equal(t, "app-c", m.Modules[1].Name())
	assert.Equal(t, "local", m.Modules[1].Version())
	assert.Equal(t, "app-d", m.Modules[2].Name())
	assert.Equal(t, "local", m.Modules[2].Version())
}

func TestDependencyChange(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.InitModuleWithOptions("app-b", &Spec{
		Name:         "app-b",
		Dependencies: []string{"app-a"},
	}))
	check(t, repo.Commit("first"))
	c1 := repo.LastCommit

	check(t, repo.WriteContent("app-a/foo", "hello"))
	check(t, repo.Commit("second"))
	c2 := repo.LastCommit

	m, err := NewWorld(t, ".tmp/repo").System.ManifestByDiff(c1.String(), c2.String())
	check(t, err)

	assert.Len(t, m.Modules, 2)
	assert.Equal(t, "app-a", m.Modules[0].Name())
	assert.Equal(t, "app-b", m.Modules[1].Name())
}

func TestIndirectDependencyChange(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.InitModuleWithOptions("app-b", &Spec{
		Name:         "app-b",
		Dependencies: []string{"app-a"},
	}))
	check(t, repo.InitModuleWithOptions("app-c", &Spec{
		Name:         "app-c",
		Dependencies: []string{"app-b"},
	}))
	check(t, repo.Commit("first"))
	c1 := repo.LastCommit

	check(t, repo.WriteContent("app-a/foo", "hello"))
	check(t, repo.Commit("second"))
	c2 := repo.LastCommit

	m, err := NewWorld(t, ".tmp/repo").System.ManifestByDiff(c1.String(), c2.String())
	check(t, err)

	assert.Len(t, m.Modules, 3)
	assert.Equal(t, "app-a", m.Modules[0].Name())
	assert.Equal(t, "app-b", m.Modules[1].Name())
	assert.Equal(t, "app-c", m.Modules[2].Name())
}

func TestDiffOfDependentChange(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.InitModuleWithOptions("app-b", &Spec{
		Name:         "app-b",
		Dependencies: []string{"app-a"},
	}))
	check(t, repo.Commit("first"))
	c1 := repo.LastCommit

	check(t, repo.WriteContent("app-b/foo", "hello"))
	check(t, repo.Commit("second"))
	c2 := repo.LastCommit

	m, err := NewWorld(t, ".tmp/repo").System.ManifestByDiff(c1.String(), c2.String())
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, "app-b", m.Modules[0].Name())
}

func TestVersionOfIndependentModules(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.InitModule("app-b"))
	check(t, repo.Commit("first"))
	c1 := repo.LastCommit

	m1, err := NewWorld(t, ".tmp/repo").System.ManifestByCommit(c1.String())
	check(t, err)

	check(t, repo.WriteContent("app-a/foo", "hello"))
	check(t, repo.Commit("second"))
	c2 := repo.LastCommit

	m2, err := NewWorld(t, ".tmp/repo").System.ManifestByCommit(c2.String())
	check(t, err)

	assert.Equal(t, m1.Modules[1].Version(), m2.Modules[1].Version())
	assert.NotEqual(t, m1.Modules[0].Version(), m2.Modules[0].Version())
}

func TestVersionOfDependentModules(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.InitModuleWithOptions("app-b", &Spec{
		Name:         "app-b",
		Dependencies: []string{"app-a"},
	}))
	check(t, repo.Commit("first"))
	c1 := repo.LastCommit

	m1, err := NewWorld(t, ".tmp/repo").System.ManifestByCommit(c1.String())
	check(t, err)

	check(t, repo.WriteContent("app-a/foo", "hello"))
	check(t, repo.Commit("second"))
	c2 := repo.LastCommit

	m2, err := NewWorld(t, ".tmp/repo").System.ManifestByCommit(c2.String())
	check(t, err)

	assert.NotEqual(t, m1.Modules[0].Version(), m2.Modules[0].Version())
	assert.NotEqual(t, m1.Modules[1].Version(), m2.Modules[1].Version())
}

func TestVersionOfIndirectlyDependentModules(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.InitModuleWithOptions("app-b", &Spec{
		Name:         "app-b",
		Dependencies: []string{"app-a"},
	}))
	check(t, repo.InitModuleWithOptions("app-c", &Spec{
		Name:         "app-c",
		Dependencies: []string{"app-b"},
	}))

	check(t, repo.Commit("first"))
	c1 := repo.LastCommit

	m1, err := NewWorld(t, ".tmp/repo").System.ManifestByCommit(c1.String())
	check(t, err)

	check(t, repo.WriteContent("app-a/foo", "hello"))
	check(t, repo.Commit("second"))
	c2 := repo.LastCommit

	m2, err := NewWorld(t, ".tmp/repo").System.ManifestByCommit(c2.String())
	check(t, err)

	assert.NotEqual(t, m1.Modules[0].Version(), m2.Modules[0].Version())
	assert.NotEqual(t, m1.Modules[1].Version(), m2.Modules[1].Version())
	assert.NotEqual(t, m1.Modules[2].Version(), m2.Modules[2].Version())
}

func TestChangeToFileDependency(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.WriteContent("shared/file", "a"))
	check(t, repo.InitModule("app-a"))
	check(t, repo.InitModuleWithOptions("app-b", &Spec{
		Name:             "app-b",
		FileDependencies: []string{"shared/file"},
	}))

	check(t, repo.Commit("first"))
	c1 := repo.LastCommit.String()

	check(t, repo.WriteContent("shared/file", "b"))
	check(t, repo.Commit("second"))
	c2 := repo.LastCommit.String()

	m, err := NewWorld(t, ".tmp/repo").System.ManifestByDiff(c1, c2)
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, "app-b", m.Modules[0].Name())
}

func TestFileDependencyInADependentModule(t *testing.T) {
	/*
		Edge case: It does not make sense to have a file dependency to a file
		in a module that you already have a dependency on. We test the correct
		behavior nevertheless.
	*/
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteContent("app-a/file", "a"))

	check(t, repo.InitModuleWithOptions("app-b", &Spec{
		Name:             "app-b",
		Dependencies:     []string{"app-a"},
		FileDependencies: []string{"app-a/file"},
	}))

	check(t, repo.Commit("first"))
	c1 := repo.LastCommit.String()

	check(t, repo.WriteContent("app-a/file", "b"))
	check(t, repo.Commit("second"))
	c2 := repo.LastCommit.String()

	m, err := NewWorld(t, ".tmp/repo").System.ManifestByDiff(c1, c2)
	check(t, err)

	assert.Len(t, m.Modules, 2)
	assert.Equal(t, "app-a", m.Modules[0].Name())
	assert.Equal(t, "app-b", m.Modules[1].Name())
}

func TestDependentOfAModuleWithFileDependency(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.WriteContent("shared/file", "a"))
	check(t, repo.InitModuleWithOptions("app-a", &Spec{
		Name:             "app-a",
		FileDependencies: []string{"shared/file"},
	}))

	check(t, repo.InitModuleWithOptions("app-b", &Spec{
		Name:         "app-b",
		Dependencies: []string{"app-a"},
	}))

	check(t, repo.InitModule("app-c"))

	check(t, repo.Commit("first"))
	c1 := repo.LastCommit.String()

	check(t, repo.WriteContent("shared/file", "b"))
	check(t, repo.Commit("second"))
	c2 := repo.LastCommit.String()

	m, err := NewWorld(t, ".tmp/repo").System.ManifestByDiff(c1, c2)
	check(t, err)

	assert.Len(t, m.Modules, 2)
	assert.Equal(t, "app-a", m.Modules[0].Name())
	assert.Equal(t, "app-b", m.Modules[1].Name())
}

func TestManifestBySha(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))

	c1 := repo.LastCommit

	check(t, repo.InitModule("app-b"))
	check(t, repo.Commit("second"))

	c2 := repo.LastCommit

	m, err := NewWorld(t, ".tmp/repo").System.ManifestByCommit(c1.String())
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, c1.String(), m.Sha)
	assert.Equal(t, "app-a", m.Modules[0].Name())

	m, err = NewWorld(t, ".tmp/repo").System.ManifestByCommit(c2.String())
	check(t, err)

	assert.Len(t, m.Modules, 2)
	assert.Equal(t, c2.String(), m.Sha)
	assert.Equal(t, "app-a", m.Modules[0].Name())
	assert.Equal(t, "app-b", m.Modules[1].Name())
}

func TestOrderOfModules(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModuleWithOptions("app-a", &Spec{
		Name:         "app-a",
		Dependencies: []string{"app-b"},
	}))

	check(t, repo.InitModuleWithOptions("app-b", &Spec{
		Name:         "app-b",
		Dependencies: []string{"app-c"},
	}))

	check(t, repo.InitModule("app-c"))
	check(t, repo.Commit("first"))

	m, err := NewWorld(t, ".tmp/repo").System.ManifestByBranch("master")
	check(t, err)

	assert.Len(t, m.Modules, 3)
	assert.Equal(t, "app-c", m.Modules[0].Name())
	assert.Equal(t, "app-b", m.Modules[1].Name())
	assert.Equal(t, "app-a", m.Modules[2].Name())
}

func TestAppsWithSamePrefix(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-aa"))
	check(t, repo.Commit("first"))

	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("second"))
	c1 := repo.LastCommit

	check(t, repo.WriteContent("app-aa/foo", "bar"))
	check(t, repo.Commit("third"))
	c2 := repo.LastCommit

	m, err := NewWorld(t, ".tmp/repo").System.ManifestByDiff(c1.String(), c2.String())
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, "app-aa", m.Modules[0].Name())
}

func TestDiffingForCaseSensitivityOfModulePath(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("App-A"))
	check(t, repo.Commit("first"))
	c1 := repo.LastCommit

	check(t, repo.WriteContent("App-A/foo", "bar"))
	check(t, repo.Commit("second"))
	c2 := repo.LastCommit

	m, err := NewWorld(t, ".tmp/repo").System.ManifestByDiff(c1.String(), c2.String())
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, "App-A", m.Modules[0].Name())
}

func TestCaseSensitivityOfFileDependency(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModuleWithOptions("App-A", &Spec{
		Name:             "App-A",
		FileDependencies: []string{"Dir1/Foo.js"},
	}))

	check(t, repo.WriteContent("dir/foo.js", "console.log('foo');"))
	check(t, repo.Commit("first"))

	m, err := NewWorld(t, ".tmp/repo").System.ManifestByCurrentBranch()

	assert.Nil(t, m)
	assert.EqualError(t, err, fmt.Sprintf(msgFileDependencyNotFound, "Dir1/Foo.js", "App-A", "App-A"))
}

func TestByDiffForDiscoverFailure(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))

	w := NewWorld(t, ".tmp/repo")
	w.Discover.Interceptor.Config("ModulesInCommit").Return(Modules(nil), errors.New("doh"))

	_, err := w.System.ManifestByDiff(repo.LastCommit.String(), repo.LastCommit.String())
	assert.EqualError(t, err, "doh")
}

func TestByDiffForMergeBaseFailure(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))

	w := NewWorld(t, ".tmp/repo")
	w.Repo.Interceptor.Config("DiffMergeBase").Return([]*DiffDelta(nil), errors.New("doh"))

	_, err := w.System.ManifestByDiff(repo.LastCommit.String(), repo.LastCommit.String())
	assert.EqualError(t, err, "doh")
}

func TestByDiffForReduceFailure(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))

	w := NewWorld(t, ".tmp/repo")
	w.Reducer.Interceptor.Config("Reduce").Return(Modules(nil), errors.New("doh"))

	_, err := w.System.ManifestByDiff(repo.LastCommit.String(), repo.LastCommit.String())
	assert.EqualError(t, err, "doh")
}

func TestByPrForInvalidSrcBranch(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))

	w := NewWorld(t, ".tmp/repo")
	_, err := w.System.ManifestByPr("master", "feature")
	assert.EqualError(t, err, fmt.Sprintf(msgFailedBranchLookup, "feature"))
}

func TestByPrForInvalidDstBranch(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))

	w := NewWorld(t, ".tmp/repo")
	_, err := w.System.ManifestByPr("feature", "master")
	assert.EqualError(t, err, fmt.Sprintf(msgFailedBranchLookup, "feature"))
}

func TestByCommitForDiscoverFailure(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))

	w := NewWorld(t, ".tmp/repo")
	w.Discover.Interceptor.Config("ModulesInCommit").Return(Commit(nil), errors.New("doh"))

	_, err := w.System.ManifestByCommit(repo.LastCommit.String())
	assert.EqualError(t, err, "doh")
}

func TestByBranchForInvalidBranch(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))

	w := NewWorld(t, ".tmp/repo")
	_, err := w.System.ManifestByBranch("feature")
	assert.EqualError(t, err, fmt.Sprintf(msgFailedBranchLookup, "feature"))
}

func TestByCurrentBranchForResolutionFailure(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))

	w := NewWorld(t, ".tmp/repo")
	w.Repo.Interceptor.Config("CurrentBranch").Return("", errors.New("doh"))

	_, err := w.System.ManifestByCurrentBranch()
	assert.EqualError(t, err, "doh")
}

func TestByWorkspaceForDiscoverFailure(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))

	w := NewWorld(t, ".tmp/repo")
	w.Discover.Interceptor.Config("ModulesInWorkspace").Return((Modules)(nil), errors.New("doh"))

	_, err := w.System.ManifestByWorkspace()
	assert.EqualError(t, err, "doh")
}

func TestByWorkspaceChangesForDiscoverFailure(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))

	w := NewWorld(t, ".tmp/repo")
	w.Discover.Interceptor.Config("ModulesInWorkspace").Return((Modules)(nil), errors.New("doh"))

	_, err := w.System.ManifestByWorkspaceChanges()
	assert.EqualError(t, err, "doh")
}

func TestByWorkspaceChangesForDiffFailure(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))

	w := NewWorld(t, ".tmp/repo")
	w.Repo.Interceptor.Config("DiffWorkspace").Return([]*DiffDelta(nil), errors.New("doh"))

	_, err := w.System.ManifestByWorkspaceChanges()
	assert.EqualError(t, err, "doh")
}

func TestByWorkspaceChangesForReduceFailure(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))

	w := NewWorld(t, ".tmp/repo")
	w.Reducer.Interceptor.Config("Reduce").Return((Modules)(nil), errors.New("doh"))

	_, err := w.System.ManifestByWorkspaceChanges()
	assert.EqualError(t, err, "doh")
}

func TestByXxxForEmptyRepoEvaluationFailure(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))

	w := NewWorld(t, ".tmp/repo")
	w.Repo.Interceptor.Config("IsEmpty").Return(false, errors.New("doh"))

	_, err := w.System.ManifestByCurrentBranch()
	assert.EqualError(t, err, "doh")
}

func TestByCommitContentForFirstCommit(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))

	w := NewWorld(t, ".tmp/repo")

	m, err := w.System.ManifestByCommitContent(repo.LastCommit.String())
	check(t, err)

	assert.Equal(t, repo.LastCommit.String(), m.Sha)
	assert.Len(t, m.Modules, 1)
	assert.Equal(t, "app-a", m.Modules[0].Name())
}

func TestByCommitContentForCommitWithParent(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))
	check(t, repo.InitModule("app-b"))
	check(t, repo.Commit("second"))

	w := NewWorld(t, ".tmp/repo")

	m, err := w.System.ManifestByCommitContent(repo.LastCommit.String())
	check(t, err)

	assert.Equal(t, repo.LastCommit.String(), m.Sha)
	assert.Len(t, m.Modules, 1)
	assert.Equal(t, "app-b", m.Modules[0].Name())
}

func TestByCommitContentForDiscoverFailure(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))

	w := NewWorld(t, ".tmp/repo")

	w.Discover.Interceptor.Config("ModulesInCommit").Return(Modules(nil), errors.New("doh"))

	m, err := w.System.ManifestByCommitContent(repo.LastCommit.String())

	assert.Nil(t, m)
	assert.EqualError(t, err, "doh")
}

func TestByCommitContentForRepoFailure(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	repo.InitModule("app-a")
	repo.Commit("first")

	w := NewWorld(t, ".tmp/repo")
	w.Repo.Interceptor.Config("Changes").Return([]*DiffDelta(nil), errors.New("doh"))

	m, err := w.System.ManifestByCommitContent(repo.LastCommit.String())

	assert.Nil(t, m)
	assert.EqualError(t, err, "doh")
}

func TestByCommitContentForReducerFailure(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	repo.InitModule("app-a")
	repo.Commit("first")
	repo.InitModule("app-b")
	repo.Commit("second")

	w := NewWorld(t, ".tmp/repo")
	w.Reducer.Interceptor.Config("Reduce").Return(Modules(nil), errors.New("doh"))

	m, err := w.System.ManifestByCommitContent(repo.LastCommit.String())

	assert.Nil(t, m)
	assert.EqualError(t, err, "doh")
}

func TestCircularDependencyInCommit(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	check(t, repo.InitModuleWithOptions("app-a", &Spec{Name: "app-a", Dependencies: []string{"app-b"}}))
	check(t, repo.InitModuleWithOptions("app-b", &Spec{Name: "app-b", Dependencies: []string{"app-a"}}))
	check(t, repo.Commit("first"))

	w := NewWorld(t, ".tmp/repo")
	m, err := w.System.ManifestByCommit(repo.LastCommit.String())

	assert.Nil(t, m)
	assert.EqualError(t, err, "Could not produce the module graph due to a cyclic dependency in path: app-a -> app-b -> app-a")
}

func TestCircularDependencyInWorkspace(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	check(t, repo.InitModuleWithOptions("app-a", &Spec{Name: "app-a", Dependencies: []string{"app-b"}}))
	check(t, repo.InitModuleWithOptions("app-b", &Spec{Name: "app-b", Dependencies: []string{"app-a"}}))
	check(t, repo.Commit("first"))

	w := NewWorld(t, ".tmp/repo")
	m, err := w.System.ManifestByWorkspace()

	assert.Nil(t, m)
	assert.EqualError(t, err, "Could not produce the module graph due to a cyclic dependency in path: app-a -> app-b -> app-a")
	assert.Equal(t, ErrClassUser, (err.(*e.E)).Class())
}

func TestManifestByWorkspaceChangesForRootModule(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	check(t, repo.InitModuleWithOptions("", &Spec{Name: "root"}))

	w := NewWorld(t, ".tmp/repo")
	m, err := w.System.ManifestByWorkspaceChanges()
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, "root", m.Modules[0].Name())
}

func TestApplyFilter(t *testing.T) {
	appA := &Module{metadata: &moduleMetadata{spec: &Spec{Name: "app-a"}}}
	appB := &Module{
		metadata:   &moduleMetadata{spec: &Spec{Name: "app-b", Dependencies: []string{"app-a"}}},
		requiredBy: Modules{appA},
	}
	appAa := &Module{metadata: &moduleMetadata{spec: &Spec{Name: "app-aa"}}}

	m := &Manifest{Modules: []*Module{appA, appB, appAa}}

	m1, err := m.ApplyFilters(NoFilter)
	check(t, err)
	assert.Equal(t, m, m1)

	m1, err = m.ApplyFilters(ExactMatchFilter("app"))
	check(t, err)
	assert.Len(t, m1.Modules, 0)

	m1, err = m.ApplyFilters(ExactMatchFilter("app-a"))
	check(t, err)
	assert.Len(t, m1.Modules, 1)
	assert.Equal(t, "app-a", m1.Modules[0].Name())

	m1, err = m.ApplyFilters(FuzzyFilter("app-a"))
	check(t, err)
	assert.Len(t, m1.Modules, 2)
	assert.Equal(t, "app-a", m1.Modules[0].Name())
	assert.Equal(t, "app-aa", m1.Modules[1].Name())

	m1, err = m.ApplyFilters(ExactMatchDependentsFilter("app-b"))
	check(t, err)
	assert.Len(t, m1.Modules, 2)
	assert.Equal(t, "app-b", m1.Modules[0].Name())
	assert.Equal(t, "app-a", m1.Modules[1].Name())

	m1, err = m.ApplyFilters(FuzzyDependentsFilter("app-b"))
	check(t, err)
	assert.Len(t, m1.Modules, 2)
	assert.Equal(t, "app-b", m1.Modules[0].Name())
	assert.Equal(t, "app-a", m1.Modules[1].Name())
}
