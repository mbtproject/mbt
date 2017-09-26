package lib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplicationDetection(t *testing.T) {
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

func TestNonApplicationDirectories(t *testing.T) {
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
