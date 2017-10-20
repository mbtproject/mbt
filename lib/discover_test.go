package lib

import "testing"
import "github.com/stretchr/testify/assert"

func TestDependencyLinks(t *testing.T) {
	a := newApplicationMetadata("app-a", "a", &Spec{Name: "app-a", Dependencies: []string{"app-b"}})
	b := newApplicationMetadata("app-b", "b", &Spec{Name: "app-b", Dependencies: []string{"app-c"}})
	c := newApplicationMetadata("app-c", "c", &Spec{Name: "app-c"})

	s := applicationMetadataSet{a, b, c}
	apps, err := s.toApplications(true)
	check(t, err)
	m := apps.indexByName()

	assert.Len(t, apps, 3)
	assert.Equal(t, m["app-b"], m["app-a"].Requires().Front().Value)
	assert.Equal(t, m["app-c"], m["app-b"].Requires().Front().Value)
	assert.Equal(t, 0, m["app-c"].Requires().Len())
	assert.Equal(t, m["app-b"], m["app-c"].RequiredBy().Front().Value)
	assert.Equal(t, m["app-a"], m["app-b"].RequiredBy().Front().Value)
}

func TestVersionCalculation(t *testing.T) {
	a := newApplicationMetadata("app-a", "a", &Spec{Name: "app-a", Dependencies: []string{"app-b"}})
	b := newApplicationMetadata("app-b", "b", &Spec{Name: "app-b"})

	s := applicationMetadataSet{a, b}
	apps, err := s.toApplications(true)
	check(t, err)
	m := apps.indexByName()

	assert.Equal(t, "b", m["app-b"].Version())
	assert.Equal(t, "da23614e02469a0d7c7bd1bdab5c9c474b1904dc", m["app-a"].Version())
}

func TestMalformedSpec(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitApplication("app-a"))
	check(t, repo.WriteContent("app-a/.mbt.yml", "blah:blah\nblah::"))
	check(t, repo.Commit("first"))

	commit, err := repo.Repo.LookupCommit(repo.LastCommit)
	metadata, err := discoverMetadata(repo.Repo, commit)
	assert.Nil(t, metadata)
	assert.EqualError(t, err, "discover: error while parsing the spec at app-a/.mbt.yml - yaml: line 1: mapping values are not allowed in this context")
}
