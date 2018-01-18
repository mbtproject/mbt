package lib

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSingleModDir(t *testing.T) {
	clean()

	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	err = repo.InitModule("app-a")
	check(t, err)

	err = repo.Commit("first")
	check(t, err)

	m, err := ManifestByBranch(".tmp/repo", "master")
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, "app-a", m.Modules[0].Name())
}

func TestNonModContent(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	err = repo.InitModule("app-a")
	check(t, err)

	err = repo.WriteContent("content/index.html", "hello")
	check(t, err)

	err = repo.Commit("first")
	check(t, err)

	m, err := ManifestByBranch(".tmp/repo", "master")
	check(t, err)

	assert.Len(t, m.Modules, 1)
}

func TestNestedAppDir(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("a/b/c/app-a"))
	check(t, repo.Commit("first"))

	m, err := ManifestByBranch(".tmp/repo", "master")
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, "a/b/c/app-a", m.Modules[0].Path())
}

func TestModsDirInModDir(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.InitModule("app-a/app-b"))
	check(t, repo.Commit("first"))

	m, err := ManifestByBranch(".tmp/repo", "master")
	check(t, err)

	assert.Len(t, m.Modules, 2)
	assert.Equal(t, "app-a", m.Modules[0].Name())
	assert.Equal(t, "app-a", m.Modules[0].Path())
	assert.Equal(t, "app-b", m.Modules[1].Name())
	assert.Equal(t, "app-a/app-b", m.Modules[1].Path())
}

func TestEmptyRepo(t *testing.T) {
	clean()

	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))

	m, err := ManifestByBranch(".tmp/repo", "master")
	check(t, err)

	assert.Len(t, m.Modules, 0)
}

func TestDiffingTwoBranches(t *testing.T) {
	clean()

	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))
	masterTip := repo.LastCommit

	check(t, repo.SwitchToBranch("feature"))
	check(t, repo.InitModule("app-b"))
	check(t, repo.Commit("second"))

	featureTip := repo.LastCommit

	m, err := ManifestByPr(".tmp/repo", "feature", "master")
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, featureTip.String(), m.Sha)
	assert.Equal(t, "app-b", m.Modules[0].Name())

	m, err = ManifestByPr(".tmp/repo", "master", "feature")
	check(t, err)

	assert.Len(t, m.Modules, 0)
	assert.Equal(t, masterTip.String(), m.Sha)
}

func TestDiffingTwoProgressedBranches(t *testing.T) {
	clean()

	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))
	check(t, repo.SwitchToBranch("feature"))
	check(t, repo.InitModule("app-b"))
	check(t, repo.Commit("second"))
	check(t, repo.SwitchToBranch("master"))
	check(t, repo.InitModule("app-c"))
	check(t, repo.Commit("third"))

	m, err := ManifestByPr(".tmp/repo", "feature", "master")
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, "app-b", m.Modules[0].Name())

	m, err = ManifestByPr(".tmp/repo", "master", "feature")
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, "app-c", m.Modules[0].Name())
}

func TestDiffingWithMultipleChangesToSameMod(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))
	check(t, repo.SwitchToBranch("feature"))
	check(t, repo.WriteContent("app-a/file1", "hello"))
	check(t, repo.Commit("second"))
	check(t, repo.WriteContent("app-a/file2", "world"))
	check(t, repo.Commit("third"))

	m, err := ManifestByPr(".tmp/repo", "feature", "master")
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, "app-a", m.Modules[0].Name())
}

func TestDiffingForDeletes(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteContent("app-a/file1", "hello world"))
	check(t, repo.Commit("first"))
	check(t, repo.SwitchToBranch("feature"))
	check(t, repo.Remove("app-a/file1"))
	check(t, repo.Commit("second"))

	m, err := ManifestByPr(".tmp/repo", "feature", "master")
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, "app-a", m.Modules[0].Name())
}

func TestDiffingForRenames(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteContent("app-a/file1", "hello world"))
	check(t, repo.Commit("first"))
	check(t, repo.SwitchToBranch("feature"))
	check(t, repo.Rename("app-a/file1", "app-a/file2"))
	check(t, repo.Commit("second"))

	m, err := ManifestByPr(".tmp/repo", "feature", "master")
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, "app-a", m.Modules[0].Name())
}

func TestModOnRoot(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModuleWithOptions("", &Spec{Name: "root-app"}))
	check(t, repo.Commit("first"))

	m, err := ManifestByBranch(".tmp/repo", "master")
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, "root-app", m.Modules[0].Name())
	assert.Equal(t, "", m.Modules[0].Path())
	assert.Equal(t, repo.LastCommit.String(), m.Modules[0].Version())
}

func TestManifestByDiff(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.InitModule("app-b"))
	check(t, repo.Commit("first"))
	c1 := repo.LastCommit

	check(t, repo.WriteContent("app-a/foo", "hello"))
	check(t, repo.Commit("second"))
	c2 := repo.LastCommit

	m, err := ManifestByDiff(".tmp/repo", c1.String(), c2.String())
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, "app-a", m.Modules[0].Name())

	m, err = ManifestByDiff(".tmp/repo", c2.String(), c1.String())
	check(t, err)

	assert.Len(t, m.Modules, 0)
}

func TestManifestByHead(t *testing.T) {
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))
	check(t, repo.SwitchToBranch("feature"))

	m, err := ManifestByHead(".tmp/repo")
	check(t, err)

	assert.Equal(t, "app-a", m.Modules[0].Name())
}

func TestManifestByLocalDir(t *testing.T) {
	clean()
	abs, err := filepath.Abs(".tmp/repo")
	check(t, err)

	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteContent("app-a/test.txt", "test contents"))
	check(t, repo.Commit("first"))

	m, err := ManifestByLocalDir(abs, false)
	check(t, err)

	assert.Equal(t, "local", m.Sha)
	assert.Equal(t, abs, m.Dir)

	// currently no modules changed locally
	assert.Equal(t, 0, len(m.Modules))

	// change the file, expect 1 module to be returned
	check(t, repo.WriteContent("app-a/test.txt", "amended contents"))

	m, err = ManifestByLocalDir(abs, false)
	check(t, err)
	assert.Len(t, m.Modules, 1)

	// add in an uncommitted module to ensure its found
	check(t, repo.InitModule("app-c"))
	m, err = ManifestByLocalDir(abs, false)
	assert.Len(t, m.Modules, 2)
	assert.Equal(t, "app-c", m.Modules[1].name)
	assert.Equal(t, "local", m.Modules[1].hash)
}

func TestManifestByLocalDirAll(t *testing.T) {
	clean()
	abs, err := filepath.Abs(".tmp/repo")
	check(t, err)

	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteContent("app-a/test.txt", "test contents"))
	check(t, repo.Commit("first"))

	m, err := ManifestByLocalDir(abs, true)
	check(t, err)

	assert.Equal(t, "local", m.Sha)
	assert.Equal(t, abs, m.Dir)

	// currently no modules changed locally
	assert.Equal(t, 1, len(m.Modules))
}

func TestDependencyChange(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

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

	m, err := ManifestByDiff(".tmp/repo", c1.String(), c2.String())
	check(t, err)

	assert.Len(t, m.Modules, 2)
	assert.Equal(t, "app-a", m.Modules[0].Name())
	assert.Equal(t, "app-b", m.Modules[1].Name())
}

func TestIndirectDependencyChange(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

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

	m, err := ManifestByDiff(".tmp/repo", c1.String(), c2.String())
	check(t, err)

	assert.Len(t, m.Modules, 3)
	assert.Equal(t, "app-a", m.Modules[0].Name())
	assert.Equal(t, "app-b", m.Modules[1].Name())
	assert.Equal(t, "app-c", m.Modules[2].Name())
}

func TestDiffOfDependentChange(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

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

	m, err := ManifestByDiff(".tmp/repo", c1.String(), c2.String())
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, "app-b", m.Modules[0].Name())
}

func TestVersionOfIndependentModules(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.InitModule("app-b"))
	check(t, repo.Commit("first"))
	c1 := repo.LastCommit

	m1, err := ManifestBySha(".tmp/repo", c1.String())
	check(t, err)

	check(t, repo.WriteContent("app-a/foo", "hello"))
	check(t, repo.Commit("second"))
	c2 := repo.LastCommit

	m2, err := ManifestBySha(".tmp/repo", c2.String())
	check(t, err)

	assert.Equal(t, m1.Modules[1].Version(), m2.Modules[1].Version())
	assert.NotEqual(t, m1.Modules[0].Version(), m2.Modules[0].Version())
}

func TestVersionOfDependentModules(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.InitModuleWithOptions("app-b", &Spec{
		Name:         "app-b",
		Dependencies: []string{"app-a"},
	}))
	check(t, repo.Commit("first"))
	c1 := repo.LastCommit

	m1, err := ManifestBySha(".tmp/repo", c1.String())
	check(t, err)

	check(t, repo.WriteContent("app-a/foo", "hello"))
	check(t, repo.Commit("second"))
	c2 := repo.LastCommit

	m2, err := ManifestBySha(".tmp/repo", c2.String())
	check(t, err)

	assert.NotEqual(t, m1.Modules[0].Version(), m2.Modules[0].Version())
	assert.NotEqual(t, m1.Modules[1].Version(), m2.Modules[1].Version())
}

func TestVersionOfIndirectlyDependentModules(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

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

	m1, err := ManifestBySha(".tmp/repo", c1.String())
	check(t, err)

	check(t, repo.WriteContent("app-a/foo", "hello"))
	check(t, repo.Commit("second"))
	c2 := repo.LastCommit

	m2, err := ManifestBySha(".tmp/repo", c2.String())
	check(t, err)

	assert.NotEqual(t, m1.Modules[0].Version(), m2.Modules[0].Version())
	assert.NotEqual(t, m1.Modules[1].Version(), m2.Modules[1].Version())
	assert.NotEqual(t, m1.Modules[2].Version(), m2.Modules[2].Version())
}

func TestChangeToFileDependency(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

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

	m, err := ManifestByDiff(".tmp/repo", c1, c2)
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
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

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

	m, err := ManifestByDiff(".tmp/repo", c1, c2)
	check(t, err)

	assert.Len(t, m.Modules, 2)
	assert.Equal(t, "app-a", m.Modules[0].Name())
	assert.Equal(t, "app-b", m.Modules[1].Name())
}

func TestDependentOfAModuleWithFileDependency(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

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

	m, err := ManifestByDiff(".tmp/repo", c1, c2)
	check(t, err)

	assert.Len(t, m.Modules, 2)
	assert.Equal(t, "app-a", m.Modules[0].Name())
	assert.Equal(t, "app-b", m.Modules[1].Name())
}

func TestManifestBySha(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))

	c1 := repo.LastCommit

	check(t, repo.InitModule("app-b"))
	check(t, repo.Commit("second"))

	c2 := repo.LastCommit

	m, err := ManifestBySha(".tmp/repo", c1.String())
	check(t, err)

	assert.Len(t, m.Modules, 1)
	assert.Equal(t, c1.String(), m.Sha)
	assert.Equal(t, "app-a", m.Modules[0].Name())

	m, err = ManifestBySha(".tmp/repo", c2.String())
	check(t, err)

	assert.Len(t, m.Modules, 2)
	assert.Equal(t, c2.String(), m.Sha)
	assert.Equal(t, "app-a", m.Modules[0].Name())
	assert.Equal(t, "app-b", m.Modules[1].Name())
}

func TestNonRepository(t *testing.T) {
	clean()
	check(t, os.MkdirAll(".tmp/repo", 0755))

	m, err := ManifestByBranch(".tmp/repo", "master")

	assert.Nil(t, m)
	assert.EqualError(t, err, "mbt: could not find repository from '.tmp/repo'")
}

func TestOrderOfModules(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

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

	m, err := ManifestByBranch(".tmp/repo", "master")
	check(t, err)

	assert.Len(t, m.Modules, 3)
	assert.Equal(t, "app-c", m.Modules[0].Name())
	assert.Equal(t, "app-b", m.Modules[1].Name())
	assert.Equal(t, "app-a", m.Modules[2].Name())
}
