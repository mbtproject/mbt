package lib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusCountOnNew(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteContent("app-a/foo", "a"))

	count, err := statusCount(repo.Repo)
	check(t, err)

	assert.Equal(t, 1, count)
}

func TestStatusCountOnEdit(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
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

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteContent("app-a/foo", "a"))
	check(t, repo.Commit("first"))

	check(t, repo.Rename("app-a/foo", "app-a/bar"))

	count, err := statusCount(repo.Repo)
	check(t, err)

	assert.Equal(t, 2, count)
}

func TestInvalidBranch(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))

	c, err := getBranchCommit(repo.Repo, "foo")

	assert.Nil(t, c)
	assert.EqualError(t, err, "mbt: no reference found for shorthand 'foo'")
}

func TestBranchName(t *testing.T) {
	clean()

	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))
	check(t, repo.SwitchToBranch("feature"))

	branch, err := getBranchName(repo.Repo)
	check(t, err)

	assert.Equal(t, "feature", branch)
}

func TestDiffByIndex(t *testing.T) {
	clean()

	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteContent("app-a/test.txt", "test contents"))
	check(t, repo.Commit("first"))

	check(t, repo.WriteContent("app-a/test.txt", "amend contents"))

	diff, err := getDiffFromIndex(repo.Repo)
	check(t, err)

	n, err := diff.NumDeltas()
	check(t, err)

	assert.Equal(t, 1, n)
}
