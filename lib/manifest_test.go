package lib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSingleAppDir(t *testing.T) {
	clean()
	// defer clean()

	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	err = repo.InitApplication("app-a")
	check(t, err)

	err = repo.Commit("first")
	check(t, err)

	m, err := ManifestByBranch(".tmp/repo", "master")
	check(t, err)

	assert.Len(t, m.Applications, 1)
	assert.Equal(t, "app-a", m.Applications[0].Name)
}

func TestNonAppContent(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	err = repo.InitApplication("app-a")
	check(t, err)

	err = repo.WriteContent("content/index.html", "hello")
	check(t, err)

	err = repo.Commit("first")
	check(t, err)

	m, err := ManifestByBranch(".tmp/repo", "master")
	check(t, err)

	assert.Len(t, m.Applications, 1)
}

func TestNestedAppDir(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitApplication("a/b/c/app-a"))
	check(t, repo.Commit("first"))

	m, err := ManifestByBranch(".tmp/repo", "master")
	check(t, err)

	assert.Len(t, m.Applications, 1)
	assert.Equal(t, "a/b/c/app-a", m.Applications[0].Path)
}

func TestAppsDirInAppDir(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitApplication("app-a"))
	check(t, repo.InitApplication("app-a/app-b"))
	check(t, repo.Commit("first"))

	m, err := ManifestByBranch(".tmp/repo", "master")
	check(t, err)

	assert.Len(t, m.Applications, 2)
	assert.Equal(t, "app-a", m.Applications[0].Name)
	assert.Equal(t, "app-a", m.Applications[0].Path)
	assert.Equal(t, "app-b", m.Applications[1].Name)
	assert.Equal(t, "app-a/app-b", m.Applications[1].Path)
}

func TestEmptyRepo(t *testing.T) {
	clean()

	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitApplication("app-a"))

	m, err := ManifestByBranch(".tmp/repo", "master")
	check(t, err)

	assert.Len(t, m.Applications, 0)
}

func TestDiffingTwoBranches(t *testing.T) {
	clean()

	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitApplication("app-a"))
	check(t, repo.Commit("first"))
	check(t, repo.SwitchToBranch("feature"))
	check(t, repo.InitApplication("app-b"))
	check(t, repo.Commit("second"))

	m, err := ManifestByPr(".tmp/repo", "feature", "master")
	check(t, err)

	assert.Len(t, m.Applications, 1)
	assert.Equal(t, "app-b", m.Applications[0].Name)

	m, err = ManifestByPr(".tmp/repo", "master", "feature")
	check(t, err)

	assert.Len(t, m.Applications, 0)
}

func TestDiffingTwoProgressedBranches(t *testing.T) {
	clean()

	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitApplication("app-a"))
	check(t, repo.Commit("first"))
	check(t, repo.SwitchToBranch("feature"))
	check(t, repo.InitApplication("app-b"))
	check(t, repo.Commit("second"))
	check(t, repo.SwitchToBranch("master"))
	check(t, repo.InitApplication("app-c"))
	check(t, repo.Commit("third"))

	m, err := ManifestByPr(".tmp/repo", "feature", "master")
	check(t, err)

	assert.Len(t, m.Applications, 1)
	assert.Equal(t, "app-b", m.Applications[0].Name)

	m, err = ManifestByPr(".tmp/repo", "master", "feature")
	check(t, err)

	assert.Len(t, m.Applications, 1)
	assert.Equal(t, "app-c", m.Applications[0].Name)
}

func TestDiffingWithMultipleChangesToSameApp(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitApplication("app-a"))
	check(t, repo.Commit("first"))
	check(t, repo.SwitchToBranch("feature"))
	check(t, repo.WriteContent("app-a/file1", "hello"))
	check(t, repo.Commit("second"))
	check(t, repo.WriteContent("app-a/file2", "world"))
	check(t, repo.Commit("third"))

	m, err := ManifestByPr(".tmp/repo", "feature", "master")
	check(t, err)

	assert.Len(t, m.Applications, 1)
	assert.Equal(t, "app-a", m.Applications[0].Name)
}

func TestDiffingForDeletes(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitApplication("app-a"))
	check(t, repo.WriteContent("app-a/file1", "hello world"))
	check(t, repo.Commit("first"))
	check(t, repo.SwitchToBranch("feature"))
	check(t, repo.Remove("app-a/file1"))
	check(t, repo.Commit("second"))

	m, err := ManifestByPr(".tmp/repo", "feature", "master")
	check(t, err)

	assert.Len(t, m.Applications, 1)
	assert.Equal(t, "app-a", m.Applications[0].Name)
}

func TestDiffingForRenames(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitApplication("app-a"))
	check(t, repo.WriteContent("app-a/file1", "hello world"))
	check(t, repo.Commit("first"))
	check(t, repo.SwitchToBranch("feature"))
	check(t, repo.Rename("app-a/file1", "app-a/file2"))
	check(t, repo.Commit("second"))

	m, err := ManifestByPr(".tmp/repo", "feature", "master")
	check(t, err)

	assert.Len(t, m.Applications, 1)
	assert.Equal(t, "app-a", m.Applications[0].Name)
}
