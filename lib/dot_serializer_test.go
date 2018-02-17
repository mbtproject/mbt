package lib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGraph(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	// Create indirect dependency graph
	check(t, repo.InitModule("lib-a"))

	check(t, repo.InitModuleWithOptions("lib-b", &Spec{
		Name:         "lib-b",
		Dependencies: []string{"lib-a"},
	}))

	check(t, repo.InitModuleWithOptions("app-a", &Spec{
		Name:         "app-a",
		Dependencies: []string{"lib-b"},
	}))

	// Create an isolated node
	check(t, repo.InitModule("app-b"))

	check(t, repo.Commit("first"))

	m, err := NewWorld(t, ".tmp/repo").ManifestBuilder.ByCurrentBranch()
	check(t, err)

	s := m.Modules.SerializeAsDot()

	assert.Equal(t, `digraph mbt {
  "lib-a"
  "lib-b" -> "lib-a"
  "app-a" -> "lib-b"
  "app-b"
}`, s)
}

func TestSerializeAsDotOfAnEmptyRepo(t *testing.T) {
	clean()
	_, err := createTestRepository(".tmp/repo")
	check(t, err)

	m, err := NewWorld(t, ".tmp/repo").ManifestBuilder.ByCurrentBranch()
	check(t, err)

	s := m.Modules.SerializeAsDot()

	assert.Equal(t, `digraph mbt {
  
}`, s)
}
