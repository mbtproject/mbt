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
	"fmt"
	"testing"

	"github.com/mbtproject/mbt/e"
	"github.com/stretchr/testify/assert"
)

func TestInvalidBranch(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))

	_, err := NewWorld(t, ".tmp/repo").Repo.BranchCommit("foo")

	assert.EqualError(t, err, fmt.Sprintf(msgFailedBranchLookup, "foo"))
	assert.EqualError(t, (err.(*e.E)).InnerError(), "no reference found for shorthand 'foo'")
	assert.Equal(t, ErrClassUser, (err.(*e.E)).Class())
}

func TestBranchName(t *testing.T) {
	clean()

	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))
	check(t, repo.SwitchToBranch("feature"))

	commit, err := NewWorld(t, ".tmp/repo").Repo.BranchCommit("feature")
	check(t, err)

	assert.Equal(t, repo.LastCommit.String(), commit.ID())
}

func TestDiffByIndex(t *testing.T) {
	clean()

	repo := NewTestRepo(t, ".tmp/repo")

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
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteContent("app-a/foo", "a"))

	err := NewWorld(t, ".tmp/repo").Repo.EnsureSafeWorkspace()

	assert.EqualError(t, err, msgDirtyWorkingDir)
}

func TestDirtyWorkspaceForUpdates(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteContent("app-a/foo", "a"))
	check(t, repo.Commit("first"))

	check(t, repo.WriteContent("app-a/foo", "b"))

	err := NewWorld(t, ".tmp/repo").Repo.EnsureSafeWorkspace()

	assert.EqualError(t, err, msgDirtyWorkingDir)
}

func TestDirtyWorkspaceForRenames(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteContent("app-a/foo", "a"))
	check(t, repo.Commit("first"))

	check(t, repo.Rename("app-a/foo", "app-a/bar"))

	err := NewWorld(t, ".tmp/repo").Repo.EnsureSafeWorkspace()

	assert.EqualError(t, err, msgDirtyWorkingDir)
}

func TestChangesOfFirstCommit(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.WriteContent("readme.md", "hello"))
	check(t, repo.Commit("first"))

	r := NewWorld(t, ".tmp/repo").Repo
	commit, err := r.GetCommit(repo.LastCommit.String())
	check(t, err)
	d, err := r.Changes(commit)
	check(t, err)

	assert.Empty(t, d)
}

func TestChangesOfCommitWithOneParent(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.WriteContent("readme.md", "hello"))
	check(t, repo.Commit("first"))
	check(t, repo.WriteContent("contributing.md", "hello"))
	check(t, repo.Commit("second"))

	r := NewWorld(t, ".tmp/repo").Repo
	commit, err := r.GetCommit(repo.LastCommit.String())
	check(t, err)

	d, err := r.Changes(commit)
	check(t, err)

	assert.Len(t, d, 1)
	assert.Equal(t, "contributing.md", d[0].NewFile)
}

func TestChangesOfCommitWithMultipleParents(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.WriteContent("readme.md", "hello"))
	check(t, repo.Commit("first"))

	check(t, repo.SwitchToBranch("feature"))
	check(t, repo.WriteContent("foo", "hello"))
	check(t, repo.Commit("second"))
	check(t, repo.WriteContent("bar", "hello"))
	check(t, repo.Commit("third"))

	check(t, repo.SwitchToBranch("master"))
	check(t, repo.WriteContent("index.md", "hello"))
	check(t, repo.Commit("fourth"))

	mergeCommitID, err := repo.SimpleMerge("feature", "master")
	check(t, err)

	check(t, repo.SwitchToBranch("master"))

	r := NewWorld(t, ".tmp/repo").Repo
	mergeCommit, err := r.GetCommit(mergeCommitID.String())
	check(t, err)
	delta, err := r.Changes(mergeCommit)
	check(t, err)

	assert.Len(t, delta, 2)
	assert.Equal(t, "bar", delta[0].NewFile)
	assert.Equal(t, "foo", delta[1].NewFile)
}

func TestChangesOfCommitWhereOneParentIsAMergeCommit(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.WriteContent("readme.md", "hello"))
	check(t, repo.Commit("first"))
	check(t, repo.SwitchToBranch("feature-b"))

	check(t, repo.SwitchToBranch("feature-a"))
	check(t, repo.WriteContent("foo", "hello"))
	check(t, repo.Commit("second"))
	check(t, repo.WriteContent("bar", "hello"))
	check(t, repo.Commit("third"))

	check(t, repo.SwitchToBranch("master"))
	check(t, repo.WriteContent("index.md", "hello"))
	check(t, repo.Commit("fourth"))

	check(t, repo.SwitchToBranch("feature-b"))
	check(t, repo.WriteContent("car", "hello"))
	check(t, repo.Commit("fifth"))

	_, err := repo.SimpleMerge("feature-a", "master")
	check(t, err)

	mergeCommitID, err := repo.SimpleMerge("feature-b", "master")
	check(t, err)
	check(t, repo.SwitchToBranch("master"))

	r := NewWorld(t, ".tmp/repo").Repo
	mergeCommit, err := r.GetCommit(mergeCommitID.String())
	check(t, err)
	delta, err := r.Changes(mergeCommit)
	check(t, err)

	assert.Len(t, delta, 1)
	assert.Equal(t, "car", delta[0].NewFile)
}

func TestGitIgnoreRulesInSafeWorkspace(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.WriteContent("readme.md", "hello"))
	check(t, repo.WriteContent(".gitignore", "!*.[Cc]ache/"))
	check(t, repo.WriteContent("subdir/.gitignore", ".cache/"))
	check(t, repo.Commit("first"))
	check(t, repo.WriteContent("subdir/.cache/foo", "hello"))

	w := NewWorld(t, ".tmp/repo")
	check(t, w.Repo.EnsureSafeWorkspace())
}
