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
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	repoDir, rootDir string
)

func init() {
	var err error
	repoDir, err = filepath.Abs(".tmp/repo")
	if err != nil {
		panic(err)
	}

	rootDir, err = filepath.Abs("/")
	if err != nil {
		panic(err)
	}
}

func TestGitRepoRootForPathWithoutAGitRepo(t *testing.T) {
	path, err := GitRepoRoot("/tmp/this/does/not/exist")
	assert.NoError(t, err)
	assert.Equal(t, rootDir, path)
}

func TestGitRepoRootForRepoDir(t *testing.T) {
	clean()
	NewTestRepo(t, ".tmp/repo")
	path, err := GitRepoRoot(".tmp/repo")
	assert.NoError(t, err)
	assert.Equal(t, repoDir, path)
}

func TestGitRepoRootForChildDir(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	check(t, repo.InitModule("app-a"))
	check(t, repo.InitModule("a/b/c/app-b"))

	path, err := GitRepoRoot(".tmp/repo/app-a")
	assert.NoError(t, err)
	assert.Equal(t, repoDir, path)

	path, err = GitRepoRoot(".tmp/repo/a/b/c/app-b")
	assert.NoError(t, err)
	assert.Equal(t, repoDir, path)
}

func TestGitRepoRootForConflictingFileName(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	check(t, repo.WriteContent("a/b/.git", "hello"))

	path, err := GitRepoRoot(".tmp/repo/a/b")
	assert.NoError(t, err)
	assert.Equal(t, repoDir, path)
}

func TestGitRepoRootForAbsPath(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	check(t, repo.InitModule("app-a"))

	abs, err := filepath.Abs(".tmp/repo")
	check(t, err)
	path, err := GitRepoRoot(filepath.Join(abs, "app-a"))

	assert.NoError(t, err)
	assert.Equal(t, repoDir, path)
}
