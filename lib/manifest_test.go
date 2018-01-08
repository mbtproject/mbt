package lib

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSingleModDir(t *testing.T) {
	clean()
	// defer clean()

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
