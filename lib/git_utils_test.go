package lib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusCountOnNew(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitApplication("app-a"))
	check(t, repo.WriteContent("app-a/foo", "a"))

	count, err := statusCount(repo.Repo)
	check(t, err)

	assert.Equal(t, 1, count)
}

func TestStatusCountOnEdit(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitApplication("app-a"))
	check(t, repo.WriteContent("app-a/foo", "a"))
	check(t, repo.Commit("first"))

	check(t, repo.WriteContent("app-a/foo", "b"))

	count, err := statusCount(repo.Repo)
	check(t, err)

	assert.Equal(t, 1, count)
}

func TestStatusCountOnRename(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitApplication("app-a"))
	check(t, repo.WriteContent("app-a/foo", "a"))
	check(t, repo.Commit("first"))

	check(t, repo.Rename("app-a/foo", "app-a/bar"))

	count, err := statusCount(repo.Repo)
	check(t, err)

	assert.Equal(t, 2, count)
}
