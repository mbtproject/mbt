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
