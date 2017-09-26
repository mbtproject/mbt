package lib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSingleAppDir(t *testing.T) {
	clean()
	// defer clean()

	repo, err := createTestRepository(".tmp/repo")
	assert.NoError(t, err)

	err = repo.InitApplication("app-a")
	assert.NoError(t, err)

	err = repo.Commit("first")
	assert.NoError(t, err)

	m, err := fromBranch(repo.Repo, ".tmp/repo", "master")
	assert.NoError(t, err)

	assert.Len(t, m.Applications, 1)
	assert.Equal(t, "app-a", m.Applications[0].Name)
}

func TestNonAppContent(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	assert.NoError(t, err)

	err = repo.InitApplication("app-a")
	assert.NoError(t, err)

	err = repo.WriteContent("content/index.html", "hello")
	assert.NoError(t, err)

	err = repo.Commit("first")
	assert.NoError(t, err)

	m, err := fromBranch(repo.Repo, ".tmp/repo", "master")
	assert.NoError(t, err)

	assert.Len(t, m.Applications, 1)
}

func TestNestedAppDir(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	assert.NoError(t, err)

	assert.NoError(t, repo.InitApplication("a/b/c/app-a"))
	assert.NoError(t, repo.Commit("first"))

	m, err := fromBranch(repo.Repo, ".tmp/repo", "master")
	assert.NoError(t, err)

	assert.Len(t, m.Applications, 1)
	assert.Equal(t, "a/b/c/app-a", m.Applications[0].Path)
}

func TestAppsDirInAppDir(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	assert.NoError(t, err)

	assert.NoError(t, repo.InitApplication("app-a"))
	assert.NoError(t, repo.InitApplication("app-a/app-b"))
	assert.NoError(t, repo.Commit("first"))

	m, err := fromBranch(repo.Repo, ".tmp/repo", "master")
	assert.NoError(t, err)

	assert.Len(t, m.Applications, 2)
	assert.Equal(t, "app-b", m.Applications[0].Name)
	assert.Equal(t, "app-a/app-b", m.Applications[0].Path)
	assert.Equal(t, "app-a", m.Applications[1].Name)
	assert.Equal(t, "app-a", m.Applications[1].Path)
}
