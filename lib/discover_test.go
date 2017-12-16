package lib

import (
	"os"
	"testing"

	"github.com/libgit2/git2go"

	"github.com/stretchr/testify/assert"
)

func TestDependencyLinks(t *testing.T) {
	a := newModuleMetadata("app-a", "a", &Spec{Name: "app-a", Dependencies: []string{"app-b"}})
	b := newModuleMetadata("app-b", "b", &Spec{Name: "app-b", Dependencies: []string{"app-c"}})
	c := newModuleMetadata("app-c", "c", &Spec{Name: "app-c"})

	s := moduleMetadataSet{a, b, c}
	mods, err := s.toModules()
	check(t, err)
	m := mods.indexByName()

	assert.Len(t, mods, 3)
	assert.Equal(t, m["app-b"], m["app-a"].Requires().Front().Value)
	assert.Equal(t, m["app-c"], m["app-b"].Requires().Front().Value)
	assert.Equal(t, 0, m["app-c"].Requires().Len())
	assert.Equal(t, m["app-b"], m["app-c"].RequiredBy().Front().Value)
	assert.Equal(t, m["app-a"], m["app-b"].RequiredBy().Front().Value)
}

func TestVersionCalculation(t *testing.T) {
	a := newModuleMetadata("app-a", "a", &Spec{Name: "app-a", Dependencies: []string{"app-b"}})
	b := newModuleMetadata("app-b", "b", &Spec{Name: "app-b"})

	s := moduleMetadataSet{a, b}
	mods, err := s.toModules()
	check(t, err)
	m := mods.indexByName()

	assert.Equal(t, "b", m["app-b"].Version())
	assert.Equal(t, "da23614e02469a0d7c7bd1bdab5c9c474b1904dc", m["app-a"].Version())
}

func TestMalformedSpec(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteContent("app-a/.mbt.yml", "blah:blah\nblah::"))
	check(t, repo.Commit("first"))

	commit, err := repo.Repo.LookupCommit(repo.LastCommit)
	check(t, err)

	metadata, err := discoverMetadata(repo.Repo, commit)
	assert.Nil(t, metadata)
	assert.EqualError(t, err, "mbt: error while parsing the spec at app-a/.mbt.yml - yaml: line 1: mapping values are not allowed in this context")
}

func TestMissingBlobs(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))

	commit, err := repo.Repo.LookupCommit(repo.LastCommit)
	check(t, err)

	check(t, os.RemoveAll(".tmp/repo/.git/objects"))
	check(t, os.Mkdir(".tmp/repo/.git/objects", 0755))

	metadata, err := discoverMetadata(repo.Repo, commit)

	assert.Nil(t, metadata)
	assert.EqualError(t, err, "mbt: error while fetching the blob object for app-a/.mbt.yml - object not found - no match for id (5ed8e79fc340352ac6b4655390b78d12d03a4462)")
}

func TestMissingTreeObject(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))
	check(t, os.RemoveAll(".tmp/repo/.git/objects/30"))

	r, err := git.OpenRepository(".tmp/repo")
	check(t, err)
	commit, err := r.LookupCommit(repo.LastCommit)
	check(t, err)

	metadata, err := discoverMetadata(r, commit)

	assert.Nil(t, metadata)
	assert.EqualError(t, err, "mbt: failed to walk the tree object - object not found - no match for id (308607113d927f0f2dc511a05c0efb5c96260d08)")
}

func TestMissingDependencies(t *testing.T) {
	s := moduleMetadataSet{&moduleMetadata{
		dir:  "app-a",
		hash: "a",
		spec: &Spec{
			Name:         "app-a",
			Dependencies: []string{"app-b"},
		},
	}}

	mods, err := s.toModules()

	assert.Nil(t, mods)
	assert.EqualError(t, err, "mbt: dependency not found app-a -> app-b")
}
