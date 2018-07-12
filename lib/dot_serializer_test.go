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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGraph(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

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
  node [shape=box fillcolor=powderblue style=filled fontcolor=black];
  "lib-a"
  "lib-b" -> "lib-a"
  "app-a" -> "lib-b"
  "app-b"
}`, s)
}

func TestGroupedGraph(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

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

	filterOptions := ExactMatchDependentsFilter("lib-b")
	m, err = m.ApplyFilters(filterOptions)
	check(t, err)

	mods := m.Modules
	s := mods.GroupedSerializeAsDot()

	assert.Equal(t, `digraph mbt {
  node [shape=box fillcolor=red style=filled fontcolor=black];
  "lib-b"
  "app-a"
  node [shape=box fillcolor=powderblue style=filled fontcolor=black];
  "lib-b" -> "lib-a"
  "app-a" -> "lib-b"
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
  node [shape=box fillcolor=powderblue style=filled fontcolor=black];
  
}`, s)
}

func TestTransitiveDependencies(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-c"))
	check(t, repo.InitModuleWithOptions("app-b", &Spec{
		Name:         "app-b",
		Dependencies: []string{"app-c"},
	}))

	check(t, repo.InitModuleWithOptions("app-a", &Spec{
		Name:         "app-a",
		Dependencies: []string{"app-b"},
	}))

	check(t, repo.Commit("first"))

	m, err := NewWorld(t, ".tmp/repo").ManifestBuilder.ByCurrentBranch()
	check(t, err)

	s := m.Modules.SerializeAsDot()

	assert.Equal(t, `digraph mbt {
  node [shape=box fillcolor=powderblue style=filled fontcolor=black];
  "app-c"
  "app-b" -> "app-c"
  "app-a" -> "app-b"
}`, s)
}

func TestTransitiveDependenciesWithFiltering(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-c"))
	check(t, repo.InitModuleWithOptions("app-b", &Spec{
		Name:         "app-b",
		Dependencies: []string{"app-c"},
	}))

	check(t, repo.InitModuleWithOptions("app-a", &Spec{
		Name:         "app-a",
		Dependencies: []string{"app-b"},
	}))

	check(t, repo.Commit("first"))

	m, err := NewWorld(t, ".tmp/repo").ManifestBuilder.ByCurrentBranch()
	check(t, err)

	m, err = m.ApplyFilters(&FilterOptions{Name: "app-a"})
	check(t, err)

	s := m.Modules.SerializeAsDot()

	// Should only include filtered modules in the graph
	assert.Equal(t, `digraph mbt {
  node [shape=box fillcolor=powderblue style=filled fontcolor=black];
  "app-a" -> "app-b"
}`, s)
}
