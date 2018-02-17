package lib

import (
	"fmt"
	"testing"

	"github.com/mbtproject/mbt/e"
	"github.com/stretchr/testify/assert"
)

func TestInvalidBranch(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))

	_, err = NewWorld(t, ".tmp/repo").Repo.BranchCommit("foo")

	assert.EqualError(t, err, fmt.Sprintf(msgFailedBranchLookup, "foo"))
	assert.EqualError(t, (err.(*e.E)).InnerError(), "no reference found for shorthand 'foo'")
	assert.Equal(t, ErrClassUser, (err.(*e.E)).Class())
}

func TestBranchName(t *testing.T) {
	clean()

	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))
	check(t, repo.SwitchToBranch("feature"))

	commit, err := NewWorld(t, ".tmp/repo").Repo.BranchCommit("feature")
	check(t, err)

	assert.Equal(t, repo.LastCommit.String(), commit.ID())
}

func TestDiffByIndex(t *testing.T) {
	clean()

	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteContent("app-a/test.txt", "test contents"))
	check(t, repo.Commit("first"))

	check(t, repo.WriteContent("app-a/test.txt", "amend contents"))

	diff, err := NewWorld(t, ".tmp/repo").Repo.DiffWorkspace()
	check(t, err)

	assert.Len(t, diff, 1)
}

func TestDirtyWorkspaceForUntracked(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteContent("app-a/foo", "a"))

	dirty, err := NewWorld(t, ".tmp/repo").Repo.IsDirtyWorkspace()
	check(t, err)

	assert.True(t, dirty)
}

func TestDirtyWorkspaceForUpdates(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteContent("app-a/foo", "a"))
	check(t, repo.Commit("first"))

	check(t, repo.WriteContent("app-a/foo", "b"))

	dirty, err := NewWorld(t, ".tmp/repo").Repo.IsDirtyWorkspace()
	check(t, err)

	assert.True(t, dirty)
}

func TestDirtyWorkspaceForRenames(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteContent("app-a/foo", "a"))
	check(t, repo.Commit("first"))

	check(t, repo.Rename("app-a/foo", "app-a/bar"))

	dirty, err := NewWorld(t, ".tmp/repo").Repo.IsDirtyWorkspace()
	check(t, err)

	assert.True(t, dirty)
}
