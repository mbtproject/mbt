package lib

import (
	"fmt"
	"os"
	"testing"

	"github.com/mbtproject/mbt/e"
	"github.com/stretchr/testify/assert"
)

func TestDependencyLinks(t *testing.T) {
	a := newModuleMetadata("app-a", "a", &Spec{Name: "app-a", Dependencies: []string{"app-b"}})
	b := newModuleMetadata("app-b", "b", &Spec{Name: "app-b", Dependencies: []string{"app-c"}})
	c := newModuleMetadata("app-c", "c", &Spec{Name: "app-c"})

	s := moduleMetadataSet{a, b, c}
	mods, err := toModules(s)
	check(t, err)
	m := mods.indexByName()

	assert.Len(t, mods, 3)
	assert.Equal(t, m["app-b"], m["app-a"].Requires()[0])
	assert.Equal(t, m["app-c"], m["app-b"].Requires()[0])
	assert.Equal(t, 0, len(m["app-c"].Requires()))
	assert.Equal(t, m["app-b"], m["app-c"].RequiredBy()[0])
	assert.Equal(t, m["app-a"], m["app-b"].RequiredBy()[0])
}

func TestVersionCalculation(t *testing.T) {
	a := newModuleMetadata("app-a", "a", &Spec{Name: "app-a", Dependencies: []string{"app-b"}})
	b := newModuleMetadata("app-b", "b", &Spec{Name: "app-b"})

	s := moduleMetadataSet{a, b}
	mods, err := toModules(s)
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

	world := NewWorld(t, ".tmp/repo")
	lc, err := world.Repo.GetCommit(repo.LastCommit.String())
	check(t, err)
	metadata, err := world.Discover.ModulesInCommit(lc)

	assert.Nil(t, metadata)
	assert.EqualError(t, err, "error while parsing the spec at app-a/.mbt.yml")
	assert.EqualError(t, (err.(*e.E).InnerError()), "yaml: line 1: mapping values are not allowed in this context")
	assert.Equal(t, ErrClassUser, (err.(*e.E)).Class())
}

func TestMissingBlobs(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))

	world := NewWorld(t, ".tmp/repo")
	lc, err := world.Repo.GetCommit(repo.LastCommit.String())
	check(t, err)

	check(t, os.RemoveAll(".tmp/repo/.git/objects"))
	check(t, os.Mkdir(".tmp/repo/.git/objects", 0755))

	metadata, err := world.Discover.ModulesInCommit(lc)
	assert.Nil(t, metadata)
	assert.EqualError(t, err, fmt.Sprintf(msgFailedTreeLoad, repo.LastCommit.String()))
	assert.EqualError(t, (err.(*e.E)).InnerError(), fmt.Sprintf("object not found - no match for id (32980cb34a5e42c0ff4e4920204206c492c8d487)"))
	assert.Equal(t, ErrClassInternal, (err.(*e.E)).Class())
}

func TestMissingTreeObject(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.Commit("first"))
	check(t, os.RemoveAll(".tmp/repo/.git/objects/f6"))

	world := NewWorld(t, ".tmp/repo")
	lc, err := world.Repo.GetCommit(repo.LastCommit.String())
	check(t, err)

	metadata, err := world.Discover.ModulesInCommit(lc)

	treeID := "32980cb34a5e42c0ff4e4920204206c492c8d487"
	assert.Nil(t, metadata)
	assert.EqualError(t, err, fmt.Sprintf(msgFailedTreeWalk, treeID))
	assert.EqualError(t, (err.(*e.E)).InnerError(), "object not found - no match for id (f6929fe5c1165232e1ee6c92532f1f2bcf936845)")
	assert.Equal(t, ErrClassInternal, (err.(*e.E)).Class())
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

	mods, err := toModules(s)

	assert.Nil(t, mods)
	assert.EqualError(t, err, "dependency not found app-a -> app-b")
	assert.Equal(t, ErrClassUser, (err.(*e.E)).Class())
}

func TestModuleNameConflicts(t *testing.T) {
	s := moduleMetadataSet{
		&moduleMetadata{
			dir:  "app-a",
			hash: "a",
			spec: &Spec{Name: "app-a"},
		},
		&moduleMetadata{
			dir:  "app-b",
			hash: "a",
			spec: &Spec{Name: "app-a"},
		},
	}

	mods, err := toModules(s)

	assert.Nil(t, mods)
	assert.EqualError(t, err, "Module name 'app-a' in directory 'app-b' conflicts with the module in 'app-a' directory")
	assert.Equal(t, ErrClassUser, (err.(*e.E)).Class())
}

func TestDirectoryEntriesCalledMbtYml(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteContent(".mbt.yml/foo", "blah"))
	check(t, repo.Commit("first"))

	world := NewWorld(t, ".tmp/repo")
	lc, err := world.Repo.GetCommit(repo.LastCommit.String())
	check(t, err)
	modules, err := world.Discover.ModulesInCommit(lc)
	check(t, err)

	assert.Len(t, modules, 1)
	assert.Equal(t, "app-a", modules[0].Name())
}
